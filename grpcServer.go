package efincore

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/artilugio0/efincore/proto"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	*proto.UnimplementedEfinProxyServer

	addr string

	requestInClientsMutex *sync.Mutex
	requestInClients      []chan requestData

	requestModClientsMutex *sync.Mutex
	requestModClients      []chan requestModData

	requestOutClientsMutex *sync.Mutex
	requestOutClients      []chan requestData

	responseInClientsMutex *sync.Mutex
	responseInClients      []chan responseData

	responseModClientsMutex *sync.Mutex
	responseModClients      []chan responseModData

	responseOutClientsMutex *sync.Mutex
	responseOutClients      []chan responseData
}

type requestData struct {
	r  *http.Request
	id uuid.UUID
}

type requestModData struct {
	r  *http.Request
	id uuid.UUID
	wg *sync.WaitGroup
}

type responseData struct {
	r  *http.Response
	id uuid.UUID
}

type responseModData struct {
	r  *http.Response
	id uuid.UUID
	wg *sync.WaitGroup
}

func NewGRPCServer(addr string) *GRPCServer {
	return &GRPCServer{
		addr: addr,

		requestInClientsMutex:  &sync.Mutex{},
		requestModClientsMutex: &sync.Mutex{},
		requestOutClientsMutex: &sync.Mutex{},

		responseInClientsMutex:  &sync.Mutex{},
		responseModClientsMutex: &sync.Mutex{},
		responseOutClientsMutex: &sync.Mutex{},
	}
}

func (s *GRPCServer) RequestInHook(r *http.Request, id uuid.UUID) error {
	group, _ := errgroup.WithContext(r.Context())

	rData := requestData{r, id}
	for _, c := range s.getRequestInClients() {
		thisC := c
		group.Go(func() error {
			// recover if the channel was closed and this function
			// writes to it. Can happen if client disconnects
			// right after getResponseInClients returned
			defer func() { recover() }()

			// TODO: get client answer and return it
			thisC <- rData
			return nil
		})
	}

	return group.Wait()
}

func (s *GRPCServer) RequestModHook(r *http.Request, id uuid.UUID) error {
	for _, c := range s.getRequestModClients() {
		err := func() error {
			// TODO: get client answer and return it

			// recover if the channel was closed and this function
			// writes to it
			defer func() { recover() }()

			var wg sync.WaitGroup
			wg.Add(1)
			c <- requestModData{r, id, &wg}
			// ensure modification was done
			wg.Wait()

			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *GRPCServer) RequestOutHook(r *http.Request, id uuid.UUID) error {
	group, _ := errgroup.WithContext(r.Context())

	rData := requestData{r, id}
	for _, c := range s.getRequestOutClients() {
		thisC := c
		group.Go(func() error {
			// recover if the channel was closed and this function
			// writes to it
			defer func() { recover() }()

			// TODO: get client answer and return it
			thisC <- rData
			return nil
		})
	}

	return group.Wait()
}

func (s *GRPCServer) ResponseInHook(r *http.Response, id uuid.UUID) error {
	group, _ := errgroup.WithContext(r.Request.Context())

	rData := responseData{r, id}
	for _, c := range s.getResponseInClients() {
		thisC := c
		group.Go(func() error {
			// recover if the channel was closed and this function
			// writes to it
			defer func() { recover() }()

			// TODO: get client answer and return it
			thisC <- rData
			return nil
		})
	}

	return group.Wait()
}

func (s *GRPCServer) ResponseModHook(r *http.Response, id uuid.UUID) error {
	for _, c := range s.getResponseModClients() {
		err := func() error {
			// TODO: get client answer and return it

			// recover if the channel was closed and this function
			// writes to it
			defer func() { recover() }()

			var wg sync.WaitGroup
			wg.Add(1)
			c <- responseModData{r, id, &wg}
			// ensure modification was done
			wg.Wait()

			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *GRPCServer) ResponseOutHook(r *http.Response, id uuid.UUID) error {
	group, _ := errgroup.WithContext(r.Request.Context())

	rData := responseData{r, id}
	for _, c := range s.getResponseOutClients() {
		thisC := c
		group.Go(func() error {
			// recover if the channel was closed and this function
			// writes to it
			defer func() { recover() }()

			// TODO: get client answer and return it
			thisC <- rData
			return nil
		})
	}

	return group.Wait()
}

func (s *GRPCServer) getRequestInClients() []chan requestData {
	s.requestInClientsMutex.Lock()
	defer s.requestInClientsMutex.Unlock()

	result := make([]chan requestData, len(s.requestInClients))
	for i, c := range s.requestInClients {
		result[i] = c
	}

	return result
}

func (s *GRPCServer) addRequestInClient() <-chan requestData {
	s.requestInClientsMutex.Lock()
	defer s.requestInClientsMutex.Unlock()

	c := make(chan requestData)
	s.requestInClients = append(s.requestInClients, c)

	return c
}

func (s *GRPCServer) removeRequestInClient(cRemove <-chan requestData) {
	s.requestInClientsMutex.Lock()
	defer s.requestInClientsMutex.Unlock()

	newRequestInClients := []chan requestData{}
	for _, c := range s.requestInClients {
		if c == cRemove {
			close(c)
			continue
		}
		newRequestInClients = append(newRequestInClients, c)
	}

	s.requestInClients = newRequestInClients
}

func (s *GRPCServer) getRequestModClients() []chan requestModData {
	s.requestModClientsMutex.Lock()
	defer s.requestModClientsMutex.Unlock()

	result := make([]chan requestModData, len(s.requestModClients))
	for i, c := range s.requestModClients {
		result[i] = c
	}

	return result
}

func (s *GRPCServer) addRequestModClient() <-chan requestModData {
	s.requestModClientsMutex.Lock()
	defer s.requestModClientsMutex.Unlock()

	c := make(chan requestModData)
	s.requestModClients = append(s.requestModClients, c)

	return c
}

func (s *GRPCServer) removeRequestModClient(cRemove <-chan requestModData) {
	s.requestModClientsMutex.Lock()
	defer s.requestModClientsMutex.Unlock()

	newRequestModClients := []chan requestModData{}
	for _, c := range s.requestModClients {
		if c == cRemove {
			close(c)
			continue
		}
		newRequestModClients = append(newRequestModClients, c)
	}

	s.requestModClients = newRequestModClients
}

func (s *GRPCServer) getRequestOutClients() []chan requestData {
	s.requestOutClientsMutex.Lock()
	defer s.requestOutClientsMutex.Unlock()

	result := make([]chan requestData, len(s.requestOutClients))
	for i, c := range s.requestOutClients {
		result[i] = c
	}

	return result
}

func (s *GRPCServer) addRequestOutClient() <-chan requestData {
	s.requestOutClientsMutex.Lock()
	defer s.requestOutClientsMutex.Unlock()

	c := make(chan requestData)
	s.requestOutClients = append(s.requestOutClients, c)

	return c
}

func (s *GRPCServer) removeRequestOutClient(cRemove <-chan requestData) {
	s.requestOutClientsMutex.Lock()
	defer s.requestOutClientsMutex.Unlock()

	newRequestOutClients := []chan requestData{}
	for _, c := range s.requestOutClients {
		if c == cRemove {
			close(c)
			continue
		}
		newRequestOutClients = append(newRequestOutClients, c)
	}

	s.requestOutClients = newRequestOutClients
}

func (s *GRPCServer) getResponseInClients() []chan responseData {
	s.responseInClientsMutex.Lock()
	defer s.responseInClientsMutex.Unlock()

	result := make([]chan responseData, len(s.responseInClients))
	for i, c := range s.responseInClients {
		result[i] = c
	}

	return result
}

func (s *GRPCServer) addResponseInClient() <-chan responseData {
	s.responseInClientsMutex.Lock()
	defer s.responseInClientsMutex.Unlock()

	c := make(chan responseData)
	s.responseInClients = append(s.responseInClients, c)

	return c
}

func (s *GRPCServer) removeResponseInClient(cRemove <-chan responseData) {
	s.responseInClientsMutex.Lock()
	defer s.responseInClientsMutex.Unlock()

	newResponseInClients := []chan responseData{}
	for _, c := range s.responseInClients {
		if c == cRemove {
			close(c)
			continue
		}
		newResponseInClients = append(newResponseInClients, c)
	}

	s.responseInClients = newResponseInClients
}

func (s *GRPCServer) getResponseOutClients() []chan responseData {
	s.responseOutClientsMutex.Lock()
	defer s.responseOutClientsMutex.Unlock()

	result := make([]chan responseData, len(s.responseOutClients))
	for i, c := range s.responseOutClients {
		result[i] = c
	}

	return result
}

func (s *GRPCServer) addResponseOutClient() <-chan responseData {
	s.responseOutClientsMutex.Lock()
	defer s.responseOutClientsMutex.Unlock()

	c := make(chan responseData)
	s.responseOutClients = append(s.responseOutClients, c)

	return c
}

func (s *GRPCServer) removeResponseOutClient(cRemove <-chan responseData) {
	s.responseOutClientsMutex.Lock()
	defer s.responseOutClientsMutex.Unlock()

	newResponseOutClients := []chan responseData{}
	for _, c := range s.responseOutClients {
		if c == cRemove {
			close(c)
			continue
		}
		newResponseOutClients = append(newResponseOutClients, c)
	}

	s.responseOutClients = newResponseOutClients
}

func (s *GRPCServer) getResponseModClients() []chan responseModData {
	s.responseModClientsMutex.Lock()
	defer s.responseModClientsMutex.Unlock()

	result := make([]chan responseModData, len(s.responseModClients))
	for i, c := range s.responseModClients {
		result[i] = c
	}

	return result
}

func (s *GRPCServer) addResponseModClient() <-chan responseModData {
	s.responseModClientsMutex.Lock()
	defer s.responseModClientsMutex.Unlock()

	c := make(chan responseModData)
	s.responseModClients = append(s.responseModClients, c)

	return c
}

func (s *GRPCServer) removeResponseModClient(cRemove <-chan responseModData) {
	s.responseModClientsMutex.Lock()
	defer s.responseModClientsMutex.Unlock()

	newResponseModClients := []chan responseModData{}
	for _, c := range s.responseModClients {
		if c == cRemove {
			close(c)
			continue
		}
		newResponseModClients = append(newResponseModClients, c)
	}

	s.responseModClients = newResponseModClients
}

// GRPC server implementation
func (s *GRPCServer) GetStats(context.Context, *proto.GetStatsInput) (*proto.GetStatsOutput, error) {
	stats := GetStatsService().Get()
	result := &proto.GetStatsOutput{}

	for k, v := range stats {
		result.Stats = append(result.Stats, &proto.Stat{
			Name:  k,
			Value: int64(v),
		})
	}

	return result, nil
}

func (s *GRPCServer) GetRequestsIn(_ *proto.GetRequestsInInput, stream proto.EfinProxy_GetRequestsInServer) error {
	c := s.addRequestInClient()
	defer s.removeRequestInClient(c)

	for reqData := range c {
		req, err := toProtoRequest(reqData.r, reqData.id)
		if err != nil {
			return err
		}

		if err := stream.Send(req); err != nil {
			return err
		}
	}

	return nil
}

func (s *GRPCServer) RequestsMod(stream proto.EfinProxy_RequestsModServer) error {
	c := s.addRequestModClient()
	defer s.removeRequestModClient(c)

	for reqData := range c {
		// TODO: distinguish between recoverable errors and send them
		//	throug channel so that the proxy knows if it can continue processing
		//	the request or not

		req, err := toProtoRequest(reqData.r, reqData.id)
		if err != nil {
			reqData.wg.Done()
			return err
		}

		if err := stream.Send(req); err != nil {
			reqData.wg.Done()
			return err
		}

		modReq, err := stream.Recv()
		if err != nil {
			reqData.wg.Done()
			return err
		}

		// update original request
		parsedURL, err := url.Parse(modReq.Url)
		if err != nil {
			reqData.wg.Done()
			return err
		}

		modifiedHeaders := http.Header{}
		for _, h := range modReq.Headers {
			modifiedHeaders.Add(h.Name, h.Value)
		}

		reqData.r.Proto = modReq.Version
		reqData.r.Method = modReq.Method
		reqData.r.URL = parsedURL
		reqData.r.Header = modifiedHeaders
		reqData.r.Body = newRBody(io.NopCloser(bytes.NewReader(modReq.Body)))

		reqData.wg.Done()
	}

	return nil
}

func (s *GRPCServer) GetRequestsOut(_ *proto.GetRequestsOutInput, stream proto.EfinProxy_GetRequestsOutServer) error {
	c := s.addRequestOutClient()
	defer s.removeRequestOutClient(c)

	for reqData := range c {
		req, err := toProtoRequest(reqData.r, reqData.id)
		if err != nil {
			return err
		}

		if err := stream.Send(req); err != nil {
			return err
		}
	}

	return nil
}

func (s *GRPCServer) GetResponsesIn(_ *proto.GetResponsesInInput, stream proto.EfinProxy_GetResponsesInServer) error {
	c := s.addResponseInClient()
	defer s.removeResponseInClient(c)

	for reqData := range c {
		req, err := toProtoResponse(reqData.r, reqData.id)
		if err != nil {
			return err
		}

		if err := stream.Send(req); err != nil {
			return err
		}
	}

	return nil
}

func (s *GRPCServer) ResponsesMod(stream proto.EfinProxy_ResponsesModServer) error {
	c := s.addResponseModClient()
	defer s.removeResponseModClient(c)

	for respData := range c {
		// TODO: distinguish between recoverable errors and send them
		//	throug channel so that the proxy knows if it can continue processing
		//	the response or not

		resp, err := toProtoResponse(respData.r, respData.id)
		if err != nil {
			respData.wg.Done()
			return err
		}

		if err := stream.Send(resp); err != nil {
			respData.wg.Done()
			return err
		}

		modResp, err := stream.Recv()
		if err != nil {
			respData.wg.Done()
			return err
		}

		// update original response
		modifiedHeaders := http.Header{}
		for _, h := range modResp.Headers {
			modifiedHeaders.Add(h.Name, h.Value)
		}

		respData.r.Proto = modResp.Version
		respData.r.Header = modifiedHeaders
		respData.r.Body = newRBody(io.NopCloser(bytes.NewReader(modResp.Body)))
		respData.r.StatusCode = int(modResp.StatusCode)

		respData.wg.Done()
	}

	return nil
}

func (s *GRPCServer) GetResponsesOut(_ *proto.GetResponsesOutInput, stream proto.EfinProxy_GetResponsesOutServer) error {
	c := s.addResponseOutClient()
	defer s.removeResponseOutClient(c)

	for reqData := range c {
		req, err := toProtoResponse(reqData.r, reqData.id)
		if err != nil {
			return err
		}

		if err := stream.Send(req); err != nil {
			return err
		}
	}

	return nil
}

func toProtoRequest(r *http.Request, id uuid.UUID) (*proto.Request, error) {
	headers := []*proto.Header{}
	for h, vs := range r.Header {
		for _, v := range vs {
			headers = append(headers, &proto.Header{
				Name:  h,
				Value: v,
			})
		}
	}

	body, err := r.Body.(*RBody).GetBytes()
	if err != nil {
		return nil, err
	}

	return &proto.Request{
		Id:      id.String(),
		Version: r.Proto,
		Url:     r.URL.String(),
		Method:  r.Method,
		Headers: headers,
		Body:    body,
	}, nil
}

func toProtoResponse(resp *http.Response, id uuid.UUID) (*proto.Response, error) {
	headers := []*proto.Header{}
	for h, vs := range resp.Header {
		for _, v := range vs {
			headers = append(headers, &proto.Header{
				Name:  h,
				Value: v,
			})
		}
	}

	body, err := resp.Body.(*RBody).GetBytes()
	if err != nil {
		return nil, err
	}

	return &proto.Response{
		Id:         id.String(),
		Version:    resp.Proto,
		Headers:    headers,
		Body:       body,
		StatusCode: uint32(resp.StatusCode),
		Status:     http.StatusText(resp.StatusCode),
	}, nil
}

func (s *GRPCServer) Run() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	proto.RegisterEfinProxyServer(grpcServer, s)
	reflection.Register(grpcServer)

	return grpcServer.Serve(lis)
}
