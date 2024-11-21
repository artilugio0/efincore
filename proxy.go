package efincore

import (
	"net/http"
	"net/url"
	"regexp"
)

type Proxy struct {
	addr string
	mitm *mitm
}

func NewProxy(addr string) *Proxy {
	return &Proxy{
		addr: addr,
		mitm: newMitm(),
	}
}

func (p *Proxy) ListenAndServe() error {
	return http.ListenAndServe(p.addr, p.mitm)
}

func (p *Proxy) Addr() string {
	return p.addr
}

func (p *Proxy) URL() *url.URL {
	u, err := url.Parse("http://" + p.Addr())
	if err != nil {
		panic(err)
	}

	return u
}

func (p *Proxy) GetStats() Stats {
	return GetStatsService().Get()
}

func (p *Proxy) SetDomainRegex(re *regexp.Regexp) {
	criteria := p.mitm.GetCriteria()
	criteria = criteria.WithDomainRegex(re)
	p.mitm.SetCriteria(criteria)
}

func (p *Proxy) AddRequestFilter(f RequestFilter) {
	criteria := p.mitm.GetCriteria()
	criteria = criteria.AddRequestFilter(f)
	p.mitm.SetCriteria(criteria)
}

func (p *Proxy) AddRequestInHook(h HookRequestRead) {
	hooks := p.mitm.GetHooks()
	hooks = hooks.AddRequestInHook(h)
	p.mitm.SetHooks(hooks)
}

func (p *Proxy) AddRequestOutHook(h HookRequestRead) {
	hooks := p.mitm.GetHooks()
	hooks = hooks.AddRequestOutHook(h)
	p.mitm.SetHooks(hooks)
}

func (p *Proxy) AddRequestModHook(h HookRequestMod) {
	hooks := p.mitm.GetHooks()
	hooks = hooks.AddRequestModHook(h)
	p.mitm.SetHooks(hooks)
}

func (p *Proxy) AddResponseInHook(h HookResponseRead) {
	hooks := p.mitm.GetHooks()
	hooks = hooks.AddResponseInHook(h)
	p.mitm.SetHooks(hooks)
}

func (p *Proxy) AddResponseOutHook(h HookResponseRead) {
	hooks := p.mitm.GetHooks()
	hooks = hooks.AddResponseOutHook(h)
	p.mitm.SetHooks(hooks)
}

func (p *Proxy) AddResponseModHook(h HookResponseMod) {
	hooks := p.mitm.GetHooks()
	hooks = hooks.AddResponseModHook(h)
	p.mitm.SetHooks(hooks)
}
