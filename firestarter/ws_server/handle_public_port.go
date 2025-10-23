package wsserver

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type PortID struct {
	Key    string `json:"key"`
	Target string `json:"target"`
}

var target_cache = map[string]string{
	"codeserverprt": "172.16.0.2:40000",
}

var proxy_cache = map[string]*HTTPWSProxy{}

func HandlePublicPort(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var entry PortID
		err := json.NewDecoder(r.Body).Decode(&entry)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		target_cache[entry.Key] = entry.Target
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Port opened"))
		return
	}
	if r.Method == "DELETE" {
		vars := mux.Vars(r)
		id := vars["id"]
		delete(target_cache, id)
		delete(proxy_cache, id)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Port closed"))
		return
	}
	if r.Method == "GET" {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		temp_list := target_cache
		delete(temp_list, "codeserverprt")
		if ok {
			data := PortID{
				Key:    id,
				Target: temp_list[id],
			}
			json.NewEncoder(w).Encode(data)
			return
		}
		json.NewEncoder(w).Encode(temp_list)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not found"))
}

func GetProxy(id string) (*HTTPWSProxy, bool) {
	target, ok := target_cache[id]
	if !ok || target == "" {
		return nil, false
	}

	cachedProxy, POk := proxy_cache[id]
	if POk {
		return cachedProxy, true
	}

	reverseProxy := NewHTTPWSProxy(target)

	// targetURL, err := url.Parse("http://" + target)
	// if err != nil {
	// 	log.Fatal("Error parsing target URL: ", err)
	// 	return nil, false
	// }
	// reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	// reverseProxy.Director = func(req *http.Request) {
	// 	req.URL.Scheme = targetURL.Scheme
	// 	req.URL.Host = targetURL.Host
	// 	req.Host = targetURL.Host

	// 	// Preserve WebSocket headers
	// 	if strings.ToLower(req.Header.Get("Connection")) == "upgrade" &&
	// 		strings.ToLower(req.Header.Get("Upgrade")) == "websocket" {
	// 		log.Printf("Web socket in")
	// 		req.Header.Set("Connection", "upgrade")
	// 		req.Header.Set("Upgrade", "websocket")
	// 	}
	// 	req.Header.Set("X-Forwarded-For", "")
	// 	req.Header.Set("X-Real-IP", "")
	// }
	proxy_cache[id] = reverseProxy

	return reverseProxy, true
}
