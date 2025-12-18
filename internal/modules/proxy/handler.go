package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/mimamch/reverse-proxy/internal/utils"
)

var transport = &http.Transport{
	// ===== Connection reuse (keepalive) =====
	MaxIdleConns:        4096,
	MaxIdleConnsPerHost: 512,
	MaxConnsPerHost:     0,
	IdleConnTimeout:     120 * time.Second, // keepalive_timeout

	// ===== Timeouts =====
	// ResponseHeaderTimeout: 60 * time.Second, // proxy_read_timeout
	// ExpectContinueTimeout: 1 * time.Second,

	// ===== Dialing =====
	DialContext: (&net.Dialer{
		Timeout:   5 * time.Second,  // proxy_connect_timeout
		KeepAlive: 30 * time.Second, // tcp_keepalive_time
	}).DialContext,

	// ===== HTTP/2 =====
	ForceAttemptHTTP2: true, // nginx http2 on;

	// ===== TLS (upstream) =====
	TLSHandshakeTimeout: 10 * time.Second,
	TLSClientConfig: &tls.Config{
		MinVersion: tls.VersionTLS12,
	},

	// ===== Disable compression (match nginx proxy) =====
	DisableCompression: true,
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {

	target, err := h.service.GetTarget(r.Host)

	if err != nil {
		if errors.Is(err, ErrNoRouteFound) {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}
		log.Printf("Error getting target for host %s: %v", r.Host, err)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	// if http and target requires https, redirect
	if r.TLS == nil && target.ForceHTTPS {
		redirectURL := fmt.Sprintf("https://%s%s", r.Host, r.RequestURI)
		http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
		return
	}

	proxy := &httputil.ReverseProxy{
		Transport: transport,
		Director: func(req *http.Request) {

			req.URL.Scheme = target.Backend.Scheme

			req.URL.Host = target.Backend.Host
			if target.Backend.Port != 0 {
				req.URL.Host = net.JoinHostPort(target.Backend.Host, fmt.Sprintf("%d", target.Backend.Port))
			}

			// req.Host = target.Backend.Host
			// if _, ok := target.Headers["Host"]; ok {
			// 	req.Host = target.Headers["Host"]
			// }

			for key, val := range target.Headers {
				req.Header.Set(key, val)
			}

			realIp := utils.GetRealClientIP(r)

			req.Header.Set("X-Forwarded-Host", req.Host)
			req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)
			req.Header.Set("X-Forwarded-For", realIp.String())
			req.Header.Set("X-Real-IP", realIp.String())
		},
		ModifyResponse: func(r *http.Response) error {
			return nil
		},

		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {

			if errors.Is(err, context.Canceled) ||
				errors.Is(err, context.DeadlineExceeded) {
				return
			}

			log.Printf("Proxy Error: %v", err)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("Destination unreachable"))
		},
	}

	proxy.ServeHTTP(w, r)
}
