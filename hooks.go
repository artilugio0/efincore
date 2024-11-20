package efincore

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type HookRequestRead interface {
	HookRead(*http.Request, uuid.UUID) error
}

type HookRequestReadFunc func(*http.Request, uuid.UUID) error

func (hf HookRequestReadFunc) HookRead(r *http.Request, id uuid.UUID) error {
	return hf(r, id)
}

type HookRequestMod interface {
	HookMod(*http.Request, uuid.UUID) error
}

type HookRequestModFunc func(*http.Request, uuid.UUID) error

func (hf HookRequestModFunc) HookMod(r *http.Request, id uuid.UUID) error {
	return hf(r, id)
}

type HookResponseRead interface {
	HookRead(*http.Response, uuid.UUID) error
}

type HookResponseReadFunc func(*http.Response, uuid.UUID) error

func (hf HookResponseReadFunc) HookRead(r *http.Response, id uuid.UUID) error {
	return hf(r, id)
}

type HookResponseMod interface {
	HookMod(*http.Response, uuid.UUID) error
}

type HookResponseModFunc func(*http.Response, uuid.UUID) error

func (hf HookResponseModFunc) HookMod(r *http.Response, id uuid.UUID) error {
	return hf(r, id)
}

type hooks struct {
	requestInReadHooks  []HookRequestRead
	requestModHooks     []HookRequestMod
	requestOutReadHooks []HookRequestRead

	responseInReadHooks  []HookResponseRead
	responseModHooks     []HookResponseMod
	responseOutReadHooks []HookResponseRead
}

func (h *hooks) RunRequestHooks(r *http.Request, id uuid.UUID) error {
	if h == nil {
		return nil
	}

	rbody := newRBody(r.Body)

	inReq := cloneRequest(r)
	go func() {
		inGroup, _ := errgroup.WithContext(inReq.Context())
		for _, hook := range h.requestInReadHooks {
			tHook := hook

			req := cloneRequest(inReq)
			req.Body = rbody.Clone()

			inGroup.Go(func() error {
				return tHook.HookRead(req, id)
			})
		}

		if err := inGroup.Wait(); err != nil {
			log.Printf("ERROR: requestIn read hooks failed for request '%s': %v", id.String(), err)
		}
	}()

	r.Body = rbody
	for _, hook := range h.requestModHooks {
		if err := hook.HookMod(r, id); err != nil {
			return err
		}

		if b, ok := r.Body.(*RBody); ok {
			b.Rewind()
		} else {
			r.Body = newRBody(r.Body)
		}
	}

	outReqRBody := r.Body.(*RBody)
	outReq := cloneRequest(r)
	go func() {
		outGroup, _ := errgroup.WithContext(r.Context())
		for _, hook := range h.requestOutReadHooks {
			tHook := hook

			req := cloneRequest(outReq)
			req.Body = outReqRBody.Clone()

			outGroup.Go(func() error {
				return tHook.HookRead(req, id)
			})
		}

		if err := outGroup.Wait(); err != nil {
			log.Printf("ERROR: requestOut read hooks failed for request '%s': %v", id.String(), err)
		}
	}()

	return nil
}

func (h *hooks) RunResponseHooks(r *http.Response, id uuid.UUID) error {
	if h == nil {
		return nil
	}

	rbody := newRBody(r.Body)

	inResp := cloneResponse(r)
	go func() {
		inGroup, _ := errgroup.WithContext(inResp.Request.Context())
		for _, hook := range h.responseInReadHooks {
			tHook := hook

			resp := cloneResponse(inResp)
			resp.Body = rbody.Clone()

			inGroup.Go(func() error {
				return tHook.HookRead(resp, id)
			})
		}

		if err := inGroup.Wait(); err != nil {
			log.Printf("ERROR: responseIn read hooks failed for response '%s': %v", id.String(), err)
		}
	}()

	r.Body = rbody
	for _, hook := range h.responseModHooks {
		if err := hook.HookMod(r, id); err != nil {
			return err
		}

		if b, ok := r.Body.(*RBody); ok {
			b.Rewind()
		} else {
			r.Body = newRBody(r.Body)
		}
	}

	outRespRBody := r.Body.(*RBody)
	outResp := cloneResponse(r)
	go func() {
		outGroup, _ := errgroup.WithContext(r.Request.Context())
		for _, hook := range h.responseOutReadHooks {
			tHook := hook

			resp := cloneResponse(outResp)
			resp.Body = outRespRBody.Clone()

			outGroup.Go(func() error {
				return tHook.HookRead(resp, id)
			})
		}

		if err := outGroup.Wait(); err != nil {
			log.Printf("ERROR: responseOut read hooks failed for response '%s': %v", id.String(), err)
		}
	}()

	return nil
}

func (h *hooks) clone() *hooks {
	if h == nil {
		return nil
	}

	return &hooks{
		requestInReadHooks:  append([]HookRequestRead{}, h.requestInReadHooks...),
		requestModHooks:     append([]HookRequestMod{}, h.requestModHooks...),
		requestOutReadHooks: append([]HookRequestRead{}, h.requestOutReadHooks...),

		responseInReadHooks:  append([]HookResponseRead{}, h.responseInReadHooks...),
		responseModHooks:     append([]HookResponseMod{}, h.responseModHooks...),
		responseOutReadHooks: append([]HookResponseRead{}, h.responseOutReadHooks...),
	}
}

func (h *hooks) AddRequestInReadHook(hook HookRequestRead) *hooks {
	newHooks := h.clone()
	if newHooks == nil {
		newHooks = &hooks{}
	}

	newHooks.requestInReadHooks = append(newHooks.requestInReadHooks, hook)

	return newHooks
}

func (h *hooks) AddRequestOutReadHook(hook HookRequestRead) *hooks {
	newHooks := h.clone()
	if newHooks == nil {
		newHooks = &hooks{}
	}

	newHooks.requestOutReadHooks = append(newHooks.requestOutReadHooks, hook)

	return newHooks
}

func (h *hooks) AddRequestModHook(hook HookRequestMod) *hooks {
	newHooks := h.clone()
	if newHooks == nil {
		newHooks = &hooks{}
	}

	newHooks.requestModHooks = append(newHooks.requestModHooks, hook)

	return newHooks
}

func (h *hooks) AddResponseInReadHook(hook HookResponseRead) *hooks {
	newHooks := h.clone()
	if newHooks == nil {
		newHooks = &hooks{}
	}

	newHooks.responseInReadHooks = append(newHooks.responseInReadHooks, hook)

	return newHooks
}

func (h *hooks) AddResponseOutReadHook(hook HookResponseRead) *hooks {
	newHooks := h.clone()
	if newHooks == nil {
		newHooks = &hooks{}
	}

	newHooks.responseOutReadHooks = append(newHooks.responseOutReadHooks, hook)

	return newHooks
}

func (h *hooks) AddResponseModHook(hook HookResponseMod) *hooks {
	newHooks := h.clone()
	if newHooks == nil {
		newHooks = &hooks{}
	}

	newHooks.responseModHooks = append(newHooks.responseModHooks, hook)

	return newHooks
}

func cloneRequest(r *http.Request) *http.Request {
	return r.Clone(r.Context())
}

func cloneResponse(r *http.Response) *http.Response {
	result := *r
	result.Header = r.Header.Clone()
	result.TransferEncoding = append([]string{}, r.TransferEncoding...)

	return &result
}
