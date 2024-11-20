package efincore

import (
	"net/http"
	"regexp"
	"strings"
)

func ExcludeFileExtensions(regex string) (RequestFilter, error) {
	re, err := regexp.Compile(`(?i)^` + regex + `$`)
	if err != nil {
		return nil, err
	}

	return RequestFilterFunc(func(r *http.Request) bool {
		fields := strings.Split(r.URL.Path, ".")

		if len(fields) < 2 {
			// if no extension present, then accept the request
			return true
		}
		extension := fields[len(fields)-1]

		accept := !re.MatchString(extension)
		return accept
	}), nil
}

func ExcludeFileTypes(includeRegex, excludeRegex string) (RequestFilter, error) {
	includeRe, err := regexp.Compile(`(?i)` + includeRegex)
	if err != nil {
		return nil, err
	}

	excludeRe, err := regexp.Compile(`(?i)` + excludeRegex)
	if err != nil {
		return nil, err
	}

	return RequestFilterFunc(func(r *http.Request) bool {
		header := r.Header.Get("accept")
		return includeRe.MatchString(header) || !excludeRe.MatchString(header)
	}), nil
}
