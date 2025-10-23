package httpproxy

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func HTTPProxy(port string) {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		subdomain := strings.Split(r.Host, ".")[0]
		split := strings.Split(subdomain, "-")
		id := split[len(split)-1]

		reverseProxy, ok := GetProxy(id)
		if ok {
			reverseProxy.ProxyHandler(w, r)
			return
		}
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad gateway"))
	})

	log.Println("Reverse proxy server listening on port " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe ("+port+"): ", err)
	}
}
