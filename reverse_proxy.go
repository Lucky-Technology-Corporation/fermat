package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func theiaProxy(toPort string) http.Handler {
	target, _ := url.Parse(fmt.Sprintf("http://localhost:%s", toPort))
	// Create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Custom director to manage WebSocket headers and rewrite path
	proxy.Director = func(req *http.Request) {
		// Rewrite the path
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/theia")
		// Set headers
		req.Header.Set("Host", target.Host)
		req.Header.Set("Upgrade", req.Header.Get("Upgrade"))
		req.Header.Set("Connection", req.Header.Get("Connection"))
	}

	return proxy
}

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
