package efincore

import (
	"net/http"
	"regexp"
)

type RequestFilter interface {
	ShouldInterceptRequest(*http.Request) bool
}

type RequestFilterFunc func(*http.Request) bool

func (f RequestFilterFunc) ShouldInterceptRequest(r *http.Request) bool {
	return f(r)
}

type criteria struct {
	domainsRe *regexp.Regexp

	requestFilters []RequestFilter
}

func newCriteria() *criteria {
	return &criteria{}
}

func (c *criteria) shouldInterceptDomain(domain string) bool {
	if c == nil || c.domainsRe == nil {
		return true
	}

	return c.domainsRe.MatchString(domain)
}

func (c *criteria) shouldInterceptRequest(r *http.Request) bool {
	if c == nil {
		return true
	}

	for _, f := range c.requestFilters {
		if !f.ShouldInterceptRequest(r) {
			return false
		}
	}

	return true
}

func (c *criteria) WithDomainRegex(re *regexp.Regexp) *criteria {
	newCriteria := c.clone()
	newCriteria.domainsRe = re

	return newCriteria
}

func (c *criteria) AddRequestFilter(f RequestFilter) *criteria {
	newCriteria := c.clone()
	newCriteria.requestFilters = append(newCriteria.requestFilters, f)

	return newCriteria
}

func (c *criteria) clone() *criteria {
	if c == nil {
		return &criteria{}
	}

	var domainsRe *regexp.Regexp
	if c.domainsRe != nil {
		domainsRe = regexp.MustCompile(c.domainsRe.String())
	}

	return &criteria{
		domainsRe:      domainsRe,
		requestFilters: append([]RequestFilter{}, c.requestFilters...),
	}
}
