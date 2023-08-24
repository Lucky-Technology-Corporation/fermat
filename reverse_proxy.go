package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func proxyPass(toPort string) http.Handler {
	target, _ := url.Parse(fmt.Sprintf("http://localhost:%s", toPort))

	proxy := httputil.NewSingleHostReverseProxy(target)

	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		req.Header.Set("Connection", req.Header.Get("Connection"))
		req.Header.Set("Upgrade", req.Header.Get("Upgrade"))
	}

	return proxy
}
