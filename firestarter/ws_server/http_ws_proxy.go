package wsserver

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type HTTPWSProxy struct {
	Target string
}

func NewHTTPWSProxy(target string) *HTTPWSProxy {
	return &HTTPWSProxy{
		Target: target,
	}
}

func (p *HTTPWSProxy) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	connHdr := strings.ToLower(r.Header.Get("Connection"))
	upgradeHdr := strings.ToLower(r.Header.Get("Upgrade"))

	incomingScheme := "http"
	if r.TLS != nil {
		incomingScheme = "https"
	}

	isWS := strings.Contains(connHdr, "upgrade") && upgradeHdr == "websocket"

	if isWS {
		p.ProxyWebSocketRaw(w, r)
	} else {
		targetScheme := incomingScheme
		backendHost := p.Target
		backendURL := &url.URL{
			Scheme:   targetScheme,
			Host:     backendHost,
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		}
		p.ProxyHTTP(backendURL, w, r)
	}
}

// ------------------ HTTP PROXY ------------------
func (p *HTTPWSProxy) ProxyHTTP(target *url.URL, w http.ResponseWriter, r *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ResponseHeaderTimeout: 1800 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	}

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		req.URL.Path = target.Path
		req.URL.RawQuery = target.RawQuery

		req.Header.Set("X-Real-IP", clientIP(r))
		req.Header.Set("X-Forwarded-For", r.Header.Get("X-Forwarded-For")+","+clientIP(r))
		req.Header.Set("X-Forwarded-Proto", getScheme(r))
		req.Header.Set("Upgrade", r.Header.Get("Upgrade"))
		req.Header.Set("Connection", "upgrade")
		req.Header.Set("Host", r.Host)

	}
	proxy.FlushInterval = 100 * time.Millisecond
	proxy.ServeHTTP(w, r)
}

// ------------------ WEBSOCKET PROXY ------------------
func (p *HTTPWSProxy) ProxyWebSocketRaw(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		log.Printf("Hijack error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	backendConn, err := net.DialTimeout("tcp", p.Target, 30*time.Second)
	if err != nil {
		log.Printf("Backend dial error: %v", err)
		return
	}
	defer backendConn.Close()

	err = r.Write(backendConn)
	if err != nil {
		log.Printf("Error writing request to backend: %v", err)
		return
	}

	errc := make(chan error, 2)

	go func() {
		_, err := io.Copy(backendConn, clientConn)
		errc <- err
	}()

	go func() {
		_, err := io.Copy(clientConn, backendConn)
		errc <- err
	}()

	<-errc
}

func clientIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip
}

func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}
