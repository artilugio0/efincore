package efincore

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/google/uuid"
)

type mitm struct {
	ca            *CA
	client        *http.Client
	readyEndpoint string

	criteria      *criteria
	criteriaMutex *sync.Mutex

	hooks      *hooks
	hooksMutex *sync.Mutex
}

func newMitm() *mitm {
	return &mitm{
		client: &http.Client{},
		ca:     NewCA(),

		criteriaMutex: &sync.Mutex{},
		hooksMutex:    &sync.Mutex{},
	}
}

func (m *mitm) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	GetStatsService().Increase(StatActiveConnections)
	defer GetStatsService().Decrease(StatActiveConnections)

	if r.Method == http.MethodConnect {
		GetStatsService().Increase(StatActiveConnectRequests)
		m.serveConnect(w, r)
		GetStatsService().Decrease(StatActiveConnectRequests)

		return
	}

	if r.URL.Path == m.readyEndpoint {
		w.WriteHeader(http.StatusOK)
		return
	}

	m.servePlainRequest(w, r)
}

func (m *mitm) SetReadyEndpoint(e string) {
	m.readyEndpoint = e
}

func (m *mitm) serveConnect(w http.ResponseWriter, r *http.Request) {
	if !m.shouldInterceptDomain(r) {
		m.requestPassthrough(w, r)
		return
	}

	m.requestInspect(w, r)
}

func (m *mitm) requestInspect(w http.ResponseWriter, r *http.Request) {
	connectURL, _ := url.Parse(r.URL.String())

	srcConn, destConn, err := m.hijack(w, r)
	if err != nil {
		// TODO: make errors more specific and check error type to
		// return correct status if needed
		w.WriteHeader(http.StatusBadGateway)
		log.Printf("ERROR: could not hijack connection: %v", err)
		return
	}
	defer srcConn.Close()
	defer destConn.Close()

	srcBufReader := bufio.NewReader(srcConn)
	destBufReader := bufio.NewReader(destConn)

	for {
		req, err := http.ReadRequest(srcBufReader)
		if err != nil {
			if err != io.EOF {
				log.Printf("could not read request: %v", err)
			}
			return
		}

		shouldIntercept := m.shouldInterceptRequest(req)
		var reqID *uuid.UUID
		var reqBytes []byte
		var respBytes []byte

		if shouldIntercept {
			GetStatsService().Increase(StatInterceptedRequests)

			req, reqID, err = m.interceptRequestConnect(req, connectURL)
			if err != nil {
				log.Printf("intercept request failed: %v", err)
				return
			}
		}

		reqBytes, err = httputil.DumpRequest(req, true)
		if err != nil {
			log.Printf("could not dump request: %v", err)
			return
		}

		_, err = io.Copy(destConn, bytes.NewReader(reqBytes))
		if err != nil {
			log.Printf("could not send request bytes to destination: %v", err)
			return
		}

		resp, err := http.ReadResponse(destBufReader, req)
		if err != nil {
			if err != io.EOF {
				log.Printf("could not read response: %s: %v", req.URL.String(), err)
			}
			return
		}

		// TODO: also implement response filters
		if shouldIntercept {
			// TODO: take into account 101... in those cases should the body be read?
			//	Comment: apparently this works fine, because it looks like http.ReadResponse
			//	body reads the bytes until it finds a \r\n\r\n, then the passthrough
			//	implemented below handles the rest of the data correctly
			GetStatsService().Increase(StatInterceptedResponses)

			resp, err = m.interceptResponse(resp, *reqID)
			if err != nil {
				log.Printf("intercept response failed: %v", err)
				return
			}
		}

		respBytes, err = httputil.DumpResponse(resp, true)
		if err != nil {
			log.Printf("could not dump response: %v", err)
			return
		}

		_, err = io.Copy(srcConn, bytes.NewReader(respBytes))
		if err != nil {
			log.Printf("could not send response bytes to client: %v", err)
			return
		}

		if resp.StatusCode == 101 {
			GetStatsService().Increase(StatUpgradedRequests)
			GetStatsService().Increase(StatActiveUpgradedRequests)
			defer GetStatsService().Decrease(StatActiveUpgradedRequests)

			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer destConn.Close()
				defer srcConn.Close()
				defer wg.Done()

				if _, err := io.Copy(destConn, srcBufReader); err != nil {
					if err != io.EOF && !errors.Is(err, net.ErrClosed) {
						log.Printf("error in upgraded connection sending data from source to destintaion: %v", err)
					}
					return
				}
			}()

			go func() {
				defer destConn.Close()
				defer srcConn.Close()
				defer wg.Done()

				if _, err := io.Copy(srcConn, destBufReader); err != nil {
					if err != io.EOF && !errors.Is(err, net.ErrClosed) {
						log.Printf("error in upgraded connection sending data from destination to source: %v", err)
					}
					return
				}
			}()

			wg.Wait()
			return
		}
	}
}

func (m *mitm) requestPassthrough(w http.ResponseWriter, r *http.Request) {
	srcConn, destConn, err := m.hijack(w, r)
	if err != nil {
		// TODO: make errors more specific and check error type to
		// return correct status if needed
		w.WriteHeader(http.StatusBadGateway)
		log.Printf("ERROR: could not hijack connection: %v", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer destConn.Close()
		defer srcConn.Close()
		defer wg.Done()

		if _, err := io.Copy(destConn, srcConn); err != nil {
			if err != io.EOF && !errors.Is(err, net.ErrClosed) {
				log.Printf("error sending data from source to destintaion: %v", err)
			}
			return
		}
	}()

	go func() {
		defer destConn.Close()
		defer srcConn.Close()
		defer wg.Done()

		if _, err := io.Copy(srcConn, destConn); err != nil {
			if err != io.EOF && !errors.Is(err, net.ErrClosed) {
				log.Printf("error sending data from destination to source: %v", err)
			}
			return
		}
	}()

	wg.Wait()
}

func (m *mitm) hijack(w http.ResponseWriter, r *http.Request) (*tls.Conn, *tls.Conn, error) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("Server does not support connection hijacking")
	}

	conn, buf, err := hj.Hijack()
	_ = buf
	if err != nil {
		return nil, nil, fmt.Errorf("Connection hijacking failed: %v", err)
	}

	if _, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
		return nil, nil, fmt.Errorf("error writing status to client: %v", err)
	}

	domain := r.URL.Hostname()
	cert, err := m.ca.GetCertificateFor(domain)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create certificate: %v", err)
	}
	srcConn := tls.Server(conn, &tls.Config{
		Certificates: []tls.Certificate{cert},
	})

	// force http/1.1 requests
	destConn, err := tls.Dial("tcp", r.URL.Host, &tls.Config{
		NextProtos:         []string{"http/1.1"},
		ServerName:         domain,
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not connect to destination: %v", err)
	}

	return srcConn, destConn, nil
}

func (m *mitm) servePlainRequest(w http.ResponseWriter, r *http.Request) {
	request := r.Clone(r.Context())
	request.RequestURI = ""

	response, err := m.client.Do(request)

	if err != nil {
		log.Printf("error sending request upstream: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	// copy headers
	wHeader := w.Header()
	for k, vs := range response.Header {
		for _, v := range vs {
			wHeader.Add(k, v)
		}
	}

	w.WriteHeader(response.StatusCode)

	defer response.Body.Close()
	if _, err := io.Copy(w, response.Body); err != nil {
		log.Printf("error while sending response body downstream: %v", err)
	}
}

func (m *mitm) interceptRequestConnect(r *http.Request, connectURL *url.URL) (*http.Request, *uuid.UUID, error) {
	id := uuid.New()

	m.hooksMutex.Lock()
	hooks := m.hooks
	m.hooksMutex.Unlock()

	req := r.Clone(r.Context())

	// add domain data to the request for hooks to have
	// that data available
	origUrl, err := url.Parse(req.URL.String())
	if err != nil {
		return nil, nil, err
	}
	origHost := req.Host

	connectPort := connectURL.Port()
	connectDomain := connectURL.Hostname()
	// TODO: do not assume https
	req.URL.Scheme = "https"
	req.URL.Host = connectDomain
	req.Host = connectDomain
	if connectPort != "443" && connectPort != "" {
		req.URL.Host += ":" + connectPort
	}

	modUrl := req.URL.String()
	modHost := req.Host

	if err := hooks.RunRequestHooks(req, id); err != nil {
		return nil, nil, err
	}

	// if hooks did not make any modification to the url
	// keep the original url
	newUrl := req.URL.String()
	if newUrl == modUrl && req.Host == modHost {
		req.URL = origUrl
		req.Host = origHost
	}

	return req, &id, nil
}

func (m *mitm) interceptResponse(r *http.Response, id uuid.UUID) (*http.Response, error) {
	m.hooksMutex.Lock()
	hooks := m.hooks
	m.hooksMutex.Unlock()

	if err := hooks.RunResponseHooks(r, id); err != nil {
		return nil, err
	}

	return r, nil
}

func (m *mitm) shouldInterceptDomain(r *http.Request) bool {
	domain := r.URL.Hostname()

	m.criteriaMutex.Lock()
	crit := m.criteria
	m.criteriaMutex.Unlock()

	return crit.shouldInterceptDomain(domain)
}

func (m *mitm) shouldInterceptRequest(r *http.Request) bool {
	m.criteriaMutex.Lock()
	crit := m.criteria
	m.criteriaMutex.Unlock()

	return crit.shouldInterceptRequest(r)
}

func (m *mitm) SetCriteria(c *criteria) {
	m.criteriaMutex.Lock()
	m.criteria = c
	m.criteriaMutex.Unlock()
}

func (m *mitm) GetCriteria() *criteria {
	m.criteriaMutex.Lock()
	defer m.criteriaMutex.Unlock()

	return m.criteria.clone()
}

func (m *mitm) SetHooks(h *hooks) {
	m.hooksMutex.Lock()
	m.hooks = h
	m.hooksMutex.Unlock()
}

func (m *mitm) GetHooks() *hooks {
	m.hooksMutex.Lock()
	defer m.hooksMutex.Unlock()

	return m.hooks.clone()
}
