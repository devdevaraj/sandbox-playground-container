package http_ws_proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

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
	targetScheme := incomingScheme
	if isWS {
		if incomingScheme == "https" {
			targetScheme = "wss"
		} else {
			targetScheme = "ws"
		}
	}

	backendHost := p.Target
	backendURL := &url.URL{
		Scheme: targetScheme,
		Host:   backendHost,
		Path:   r.URL.Path,
	}

	if isWS {
		p.ProxyWebSocket(backendURL, w, r)
	} else {
		p.ProxyHTTP(backendURL, w, r)
	}
}

func (p *HTTPWSProxy) ProxyHTTP(target *url.URL, w http.ResponseWriter, r *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		req.URL.Path = target.Path
		req.Header.Set("X-Forwarded-For", "")
		req.Header.Set("X-Real-IP", "")
	}
	proxy.ServeHTTP(w, r)
}

func (p *HTTPWSProxy) ProxyWebSocket(target *url.URL, w http.ResponseWriter, r *http.Request) {
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade client websocket:", err)
		return
	}
	defer clientConn.Close()

	backendConn, _, err := websocket.DefaultDialer.Dial(target.String(), nil)
	if err != nil {
		log.Println("Failed to dial backend websocket:", err)
		return
	}
	defer backendConn.Close()

	errc := make(chan error, 2)

	go func() {
		for {
			mt, msg, err := clientConn.ReadMessage()
			if err != nil {
				errc <- err
				return
			}
			err = backendConn.WriteMessage(mt, msg)
			if err != nil {
				errc <- err
				return
			}
		}
	}()

	go func() {
		for {
			mt, msg, err := backendConn.ReadMessage()
			if err != nil {
				errc <- err
				return
			}
			err = clientConn.WriteMessage(mt, msg)
			if err != nil {
				errc <- err
				return
			}
		}
	}()

	<-errc
}
