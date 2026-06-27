package server

import (
	"net/http"
	"time"
)

// NewHTTPServer returns an http.Server configured with conservative timeouts so
// it is not exposed to slow or stalled connections.
func NewHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
