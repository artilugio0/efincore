package efincore

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHTTPRequest(t *testing.T) {
	proxy := runTestProxy(t)

	reqHeader := "x-req-header"
	reqHeaderValue := "req-header-value"
	reqBody := "req-body-value"

	respHeader := "x-resp-header"
	respHeaderValue := "resp-header-value"
	respBody := "resp-body-value"
	respStatusCode := http.StatusCreated // different than 200 which is the default

	gotReqHeaderValue := ""
	gotReqBody := ""

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotReqHeaderValue = r.Header.Get(reqHeader)
		defer r.Body.Close()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read request body: %v", err)
		}
		gotReqBody = string(bodyBytes)

		w.Header().Add(respHeader, respHeaderValue)
		w.WriteHeader(respStatusCode)
		w.Write([]byte(respBody))
	})

	client := newTestClientProxy(t, proxy.URL().String())

	req, err := http.NewRequest(http.MethodGet, server.URL, strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	req.Header.Add(reqHeader, reqHeaderValue)

	response, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	gotRespHeaderValue := response.Header.Get(respHeader)

	defer response.Body.Close()
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("could not read response body")
	}
	gotRespBody := string(bodyBytes)

	if gotReqHeaderValue != reqHeaderValue {
		t.Errorf("request header: got '%s', expected '%s'", gotReqHeaderValue, reqHeaderValue)
	}

	if gotReqBody != reqBody {
		t.Errorf("request body: got '%s', expected '%s'", gotReqBody, reqBody)
	}

	if response.StatusCode != respStatusCode {
		t.Errorf("response status code: got '%d', expected '%d'", response.StatusCode, respStatusCode)
	}

	if gotRespHeaderValue != respHeaderValue {
		t.Errorf("response header: got '%s', expected '%s'", gotRespHeaderValue, respHeaderValue)
	}

	if gotRespBody != respBody {
		t.Errorf("response body: got '%s', expected '%s'", gotRespBody, respBody)
	}
}

func TestHTTPSRequest(t *testing.T) {
	proxy := runTestProxy(t)

	reqHeader := "x-req-header"
	reqHeaderValue := "req-header-value"
	reqBody := "req-body-value"

	respHeader := "x-resp-header"
	respHeaderValue := "resp-header-value"
	respBody := "resp-body-value"
	respStatusCode := http.StatusCreated // different than 200 which is the default

	gotReqHeaderValue := ""
	gotReqBody := ""

	server := newTestServerHTTPS(t, func(w http.ResponseWriter, r *http.Request) {
		gotReqHeaderValue = r.Header.Get(reqHeader)
		defer r.Body.Close()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read request body: %v", err)
		}
		gotReqBody = string(bodyBytes)

		w.Header().Add(respHeader, respHeaderValue)
		w.WriteHeader(respStatusCode)
		w.Write([]byte(respBody))
	})

	client := newTestClientProxy(t, proxy.URL().String())

	req, err := http.NewRequest(http.MethodGet, server.URL, strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	req.Header.Add(reqHeader, reqHeaderValue)

	response, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	gotRespHeaderValue := response.Header.Get(respHeader)

	defer response.Body.Close()
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("could not read response body")
	}
	gotRespBody := string(bodyBytes)

	if gotReqHeaderValue != reqHeaderValue {
		t.Errorf("request header: got '%s', expected '%s'", gotReqHeaderValue, reqHeaderValue)
	}

	if gotReqBody != reqBody {
		t.Errorf("request body: got '%s', expected '%s'", gotReqBody, reqBody)
	}

	if response.StatusCode != respStatusCode {
		t.Errorf("response status code: got '%d', expected '%d'", response.StatusCode, respStatusCode)
	}

	if gotRespHeaderValue != respHeaderValue {
		t.Errorf("response header: got '%s', expected '%s'", gotRespHeaderValue, respHeaderValue)
	}

	if gotRespBody != respBody {
		t.Errorf("response body: got '%s', expected '%s'", gotRespBody, respBody)
	}
}

func TestSingleRequestInReadHook_ServerReceivesBody(t *testing.T) {
	proxy := runTestProxy(t)

	var doneWg sync.WaitGroup
	doneWg.Add(1)

	proxy.AddRequestInReadHook(HookRequestReadFunc(func(r *http.Request, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		if _, err := io.ReadAll(r.Body); err != nil {
			t.Fatalf("could not read request body in hook: %v", err)
		}

		return nil
	}))

	var gotReqBody string
	reqBody := "request-body"

	server := newTestServerHTTPS(t, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read request body: %v", err)
		}

		gotReqBody = string(bodyBytes)
	})

	client := newTestClientProxy(t, proxy.URL().String())

	req, err := http.NewRequest(http.MethodGet, server.URL, strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	if _, err := client.Do(req); err != nil {
		t.Fatalf("request failed: %v", err)
	}
	doneWg.Wait()

	if gotReqBody != reqBody {
		t.Errorf("request body: got '%s', expected '%s'", gotReqBody, reqBody)
	}
}

func TestTwoRequestInReadHooks_HooksReadTheSameBody(t *testing.T) {
	proxy := runTestProxy(t)

	var gotBody1 string
	var gotBody2 string
	var doneWg sync.WaitGroup
	doneWg.Add(2)

	proxy.AddRequestInReadHook(HookRequestReadFunc(func(r *http.Request, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read request body in hook: %v", err)
		}
		gotBody1 = string(b)

		return nil
	}))

	proxy.AddRequestInReadHook(HookRequestReadFunc(func(r *http.Request, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read request body in hook: %v", err)
		}
		gotBody2 = string(b)

		return nil
	}))

	var gotReqBody string
	reqBody := "request-body"

	server := newTestServerHTTPS(t, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read request body: %v", err)
		}

		gotReqBody = string(bodyBytes)
	})

	client := newTestClientProxy(t, proxy.URL().String())

	req, err := http.NewRequest(http.MethodGet, server.URL, strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	if _, err := client.Do(req); err != nil {
		t.Fatalf("request failed: %v", err)
	}
	doneWg.Wait()

	if gotReqBody != reqBody {
		t.Errorf("request body on server: got '%s', expected '%s'", gotReqBody, reqBody)
	}

	if gotBody1 != reqBody {
		t.Errorf("request body on hook 1: got '%s', expected '%s'", gotBody1, reqBody)
	}

	if gotBody2 != reqBody {
		t.Errorf("request body on hook 2: got '%s', expected '%s'", gotBody2, reqBody)
	}
}

func TestRequestInModAndOutHooks_OutHookAndServerGetModifiedBody(t *testing.T) {
	proxy := runTestProxy(t)

	var gotBodyServer string
	var gotBodyHook1 string
	var gotBodyHook2 string
	var gotBodyHook3 string

	reqBody := "request-body"
	modifiedBody := "modified-request-body"

	var doneWg sync.WaitGroup
	doneWg.Add(3)

	proxy.AddRequestInReadHook(HookRequestReadFunc(func(r *http.Request, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read request body in hook: %v", err)
		}
		gotBodyHook1 = string(b)

		return nil
	}))

	proxy.AddRequestModHook(HookRequestModFunc(func(r *http.Request, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read request body in hook: %v", err)
		}
		gotBodyHook2 = string(b)

		r.Body = io.NopCloser(strings.NewReader(modifiedBody))
		r.Header.Set("Content-Length", strconv.Itoa(len(modifiedBody)))

		return nil
	}))

	proxy.AddRequestOutReadHook(HookRequestReadFunc(func(r *http.Request, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read request body in hook: %v", err)
		}
		gotBodyHook3 = string(b)

		return nil
	}))

	server := newTestServerHTTPS(t, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read request body: %v", err)
		}

		gotBodyServer = string(bodyBytes)
	})

	client := newTestClientProxy(t, proxy.URL().String())

	req, err := http.NewRequest(http.MethodGet, server.URL, strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	if _, err := client.Do(req); err != nil {
		t.Fatalf("request failed: %v", err)
	}
	doneWg.Wait()

	if gotBodyServer != modifiedBody {
		t.Errorf("request body on server: got '%s', expected '%s'", gotBodyServer, modifiedBody)
	}

	if gotBodyHook1 != reqBody {
		t.Errorf("request body on in read hook: got '%s', expected '%s'", gotBodyHook1, reqBody)
	}

	if gotBodyHook2 != reqBody {
		t.Errorf("request body on mod hook: got '%s', expected '%s'", gotBodyHook2, reqBody)
	}

	if gotBodyHook3 != modifiedBody {
		t.Errorf("request body on out hook: got '%s', expected '%s'", gotBodyHook3, modifiedBody)
	}
}

func TestRequestInModAndOutHooks_OutHookAndServerGetModifiedBody_UsingRBodyGetBytes(t *testing.T) {
	proxy := runTestProxy(t)

	var gotBodyServer string
	var gotBodyHook1 string
	var gotBodyHook2 string
	var gotBodyHook3 string

	reqBody := "request-body"
	modifiedBody := "modified-request-body"

	var doneWg sync.WaitGroup
	doneWg.Add(3)

	proxy.AddRequestInReadHook(HookRequestReadFunc(func(r *http.Request, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := r.Body.(*RBody).GetBytes()
		if err != nil {
			t.Fatalf("could not read request body in hook: %v", err)
		}
		gotBodyHook1 = string(b)

		return nil
	}))

	proxy.AddRequestModHook(HookRequestModFunc(func(r *http.Request, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := r.Body.(*RBody).GetBytes()
		if err != nil {
			t.Fatalf("could not read request body in hook: %v", err)
		}
		gotBodyHook2 = string(b)

		r.Body = io.NopCloser(strings.NewReader(modifiedBody))
		r.Header.Set("Content-Length", strconv.Itoa(len(modifiedBody)))

		return nil
	}))

	proxy.AddRequestOutReadHook(HookRequestReadFunc(func(r *http.Request, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := r.Body.(*RBody).GetBytes()
		if err != nil {
			t.Fatalf("could not read request body in hook: %v", err)
		}
		gotBodyHook3 = string(b)

		return nil
	}))

	server := newTestServerHTTPS(t, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read request body: %v", err)
		}

		gotBodyServer = string(bodyBytes)
	})

	client := newTestClientProxy(t, proxy.URL().String())

	req, err := http.NewRequest(http.MethodGet, server.URL, strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	if _, err := client.Do(req); err != nil {
		t.Fatalf("request failed: %v", err)
	}
	doneWg.Wait()

	if gotBodyServer != modifiedBody {
		t.Errorf("request body on server: got '%s', expected '%s'", gotBodyServer, modifiedBody)
	}

	if gotBodyHook1 != reqBody {
		t.Errorf("request body on in read hook: got '%s', expected '%s'", gotBodyHook1, reqBody)
	}

	if gotBodyHook2 != reqBody {
		t.Errorf("request body on mod hook: got '%s', expected '%s'", gotBodyHook2, reqBody)
	}

	if gotBodyHook3 != modifiedBody {
		t.Errorf("request body on out hook: got '%s', expected '%s'", gotBodyHook3, modifiedBody)
	}
}

// Responses hooks tests
func TestSingleResponseInReadHook_ClientReceivesBody(t *testing.T) {
	proxy := runTestProxy(t)

	var doneWg sync.WaitGroup
	doneWg.Add(1)

	proxy.AddResponseInReadHook(HookResponseReadFunc(func(r *http.Response, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		if _, err := io.ReadAll(r.Body); err != nil {
			t.Fatalf("could not read response body in hook: %v", err)
		}

		return nil
	}))

	respBody := "response-body"
	server := newTestServerHTTPS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(respBody))
	})

	client := newTestClientProxy(t, proxy.URL().String())

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	doneWg.Wait()
	defer resp.Body.Close()

	gotRespBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}
	gotRespBody := string(gotRespBodyBytes)

	if gotRespBody != respBody {
		t.Errorf("response body: got '%s', expected '%s'", gotRespBody, respBody)
	}
}

func TestTwoResponseInReadHooks_HooksReadTheSameBody(t *testing.T) {
	proxy := runTestProxy(t)

	var gotBody1 string
	var gotBody2 string

	var doneWg sync.WaitGroup
	doneWg.Add(2)

	proxy.AddResponseInReadHook(HookResponseReadFunc(func(r *http.Response, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read response body in hook: %v", err)
		}
		gotBody1 = string(b)

		return nil
	}))

	proxy.AddResponseInReadHook(HookResponseReadFunc(func(r *http.Response, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read response body in hook: %v", err)
		}
		gotBody2 = string(b)

		return nil
	}))

	respBody := "response-body"
	server := newTestServerHTTPS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(respBody))
	})

	client := newTestClientProxy(t, proxy.URL().String())

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	doneWg.Wait()
	defer resp.Body.Close()

	gotRespBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}
	gotRespBody := string(gotRespBodyBytes)

	if gotRespBody != respBody {
		t.Errorf("response body on client: got '%s', expected '%s'", gotRespBody, respBody)
	}

	if gotBody1 != respBody {
		t.Errorf("response body on hook 1: got '%s', expected '%s'", gotBody1, respBody)
	}

	if gotBody2 != respBody {
		t.Errorf("response body on hook 2: got '%s', expected '%s'", gotBody2, respBody)
	}
}

func TestResponseInModAndOutHooks_OutHookAndServerGetModifiedBody(t *testing.T) {
	proxy := runTestProxy(t)

	var gotBodyHook1 string
	var gotBodyHook2 string
	var gotBodyHook3 string

	respBody := "response-body"
	modifiedBody := "modified-response-body"

	var doneWg sync.WaitGroup
	doneWg.Add(3)

	proxy.AddResponseInReadHook(HookResponseReadFunc(func(r *http.Response, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read response body in hook: %v", err)
		}
		gotBodyHook1 = string(b)

		return nil
	}))

	proxy.AddResponseModHook(HookResponseModFunc(func(r *http.Response, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read response body in hook: %v", err)
		}
		gotBodyHook2 = string(b)

		r.Body = io.NopCloser(strings.NewReader(modifiedBody))
		r.ContentLength = int64(len(modifiedBody))
		r.Header.Set("Content-Length", strconv.Itoa(len(modifiedBody)))

		return nil
	}))

	proxy.AddResponseOutReadHook(HookResponseReadFunc(func(r *http.Response, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("could not read response body in hook: %v", err)
		}
		gotBodyHook3 = string(b)

		return nil
	}))

	server := newTestServerHTTPS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(respBody))
	})

	client := newTestClientProxy(t, proxy.URL().String())

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	doneWg.Wait()
	defer resp.Body.Close()

	gotRespBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}
	gotBodyClient := string(gotRespBodyBytes)

	if gotBodyClient != modifiedBody {
		t.Errorf("response body on client: got '%s', expected '%s'", gotBodyClient, modifiedBody)
	}

	if gotBodyHook1 != respBody {
		t.Errorf("response body on in read hook: got '%s', expected '%s'", gotBodyHook1, respBody)
	}

	if gotBodyHook2 != respBody {
		t.Errorf("response body on mod hook: got '%s', expected '%s'", gotBodyHook2, respBody)
	}

	if gotBodyHook3 != modifiedBody {
		t.Errorf("response body on out hook: got '%s', expected '%s'", gotBodyHook3, modifiedBody)
	}
}

func TestResponseInModAndOutHooks_OutHookAndServerGetModifiedBody_UsingRBodyGetBytes(t *testing.T) {
	proxy := runTestProxy(t)

	var gotBodyHook1 string
	var gotBodyHook2 string
	var gotBodyHook3 string

	respBody := "response-body"
	modifiedBody := "modified-response-body"

	var doneWg sync.WaitGroup
	doneWg.Add(3)

	proxy.AddResponseInReadHook(HookResponseReadFunc(func(r *http.Response, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := r.Body.(*RBody).GetBytes()
		if err != nil {
			t.Fatalf("could not read response body in hook: %v", err)
		}
		gotBodyHook1 = string(b)

		return nil
	}))

	proxy.AddResponseModHook(HookResponseModFunc(func(r *http.Response, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := r.Body.(*RBody).GetBytes()
		if err != nil {
			t.Fatalf("could not read response body in hook: %v", err)
		}
		gotBodyHook2 = string(b)

		r.Body = io.NopCloser(strings.NewReader(modifiedBody))
		r.ContentLength = int64(len(modifiedBody))
		r.Header.Set("Content-Length", strconv.Itoa(len(modifiedBody)))

		return nil
	}))

	proxy.AddResponseOutReadHook(HookResponseReadFunc(func(r *http.Response, id uuid.UUID) error {
		defer doneWg.Done()
		defer r.Body.Close()

		b, err := r.Body.(*RBody).GetBytes()
		if err != nil {
			t.Fatalf("could not read response body in hook: %v", err)
		}
		gotBodyHook3 = string(b)

		return nil
	}))

	server := newTestServerHTTPS(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(respBody))
	})

	client := newTestClientProxy(t, proxy.URL().String())

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	doneWg.Wait()
	defer resp.Body.Close()

	gotRespBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}
	gotBodyClient := string(gotRespBodyBytes)

	if gotBodyClient != modifiedBody {
		t.Errorf("response body on client: got '%s', expected '%s'", gotBodyClient, modifiedBody)
	}

	if gotBodyHook1 != respBody {
		t.Errorf("response body on in read hook: got '%s', expected '%s'", gotBodyHook1, respBody)
	}

	if gotBodyHook2 != respBody {
		t.Errorf("response body on mod hook: got '%s', expected '%s'", gotBodyHook2, respBody)
	}

	if gotBodyHook3 != modifiedBody {
		t.Errorf("response body on out hook: got '%s', expected '%s'", gotBodyHook3, modifiedBody)
	}
}

func runTestProxy(t *testing.T) *Proxy {
	port := getFreePort(t)
	addr := "127.0.0.1:" + strconv.Itoa(port)
	proxy := NewProxy(addr)
	proxy.mitm.SetReadyEndpoint("/ready")

	errChan := make(chan error)
	go func() {
		errChan <- proxy.ListenAndServe()
	}()

	go func() {
		var response *http.Response
		retries := 5
		for range retries {
			url := "http://" + proxy.Addr() + "/ready"
			r, err := http.Get(url)
			if err != nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			response = r
			break
		}

		if response == nil {
			errChan <- errors.New("could not establish connection to the proxy")
			return
		}

		if response.StatusCode != http.StatusOK {
			errChan <- fmt.Errorf("unexpected status code on ready endpoint: %d", response.StatusCode)
			return
		}

		errChan <- nil
	}()

	if err := <-errChan; err != nil {
		t.Fatalf("could not run test proxy: %v", err)
	}

	return proxy
}

func getFreePort(t *testing.T) int {
	t.Helper()

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := l.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return l.Addr().(*net.TCPAddr).Port
}

func newTestServer(h http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(h)
	return server
}

func newTestServerHTTPS(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()

	ca := NewCA()
	cert, err := ca.GetCertificateFor("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	server := httptest.NewUnstartedServer(h)
	server.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}
	server.StartTLS()

	return server
}

func newTestClientProxy(t *testing.T, proxyUrl string) *http.Client {
	t.Helper()

	pu, err := url.Parse(proxyUrl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(pu),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func TestChanComp(t *testing.T) {
	// test to see if channels can be compared
	// simmilar logic is used in grpc server for grpc hooks

	c := make(chan struct{})
	c3 := make(chan struct{})
	var c2 <-chan struct{} = c

	if c != c2 {
		t.Errorf("expected channels to be equal")
	}

	if c == c3 {
		t.Errorf("expected channels c and c3 to be different")
	}

	if c2 == c3 {
		t.Errorf("expected channels c2 and c3 to be different")
	}

	l := []chan struct{}{c, c3}
	newL := []chan struct{}{}
	for _, x := range l {
		if x == c {
			continue
		}
		newL = append(newL, x)
	}

	if len(newL) != 1 {
		t.Errorf("expected new channel list to have 1 element")
	}

	if len(newL) == 1 && newL[0] != c3 {
		t.Errorf("expected new channel list to have c3 as its only element")
	}
}

func TestClosedChanWritePanicHandling(t *testing.T) {
	defer func() {
		if recover() != nil {
			t.Errorf("expected a panic to be handled")
		}
	}()

	c := make(chan struct{})
	go func() {
		<-c
	}()
	close(c)

	func() {
		defer func() { recover() }()
		c <- struct{}{}
	}()
}
