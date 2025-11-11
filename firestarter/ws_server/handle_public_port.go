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

var target_cache = TARGET_CACHE

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
	proxy_cache[id] = reverseProxy
	return reverseProxy, true
}
