package httpproxy

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	wsserver "github.com/devdevaraj/firestarter/ws_server"
	"github.com/gorilla/mux"
)

func HTTPProxy(port string) {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Log for each request")
		subdomain := strings.Split(r.Host, ".")[0]
		split := strings.Split(subdomain, "-")
		id := split[len(split)-1]

		reverseProxy, ok := wsserver.GetProxy(id)
		if ok {
			reverseProxy.ProxyHandler(w, r)
			return
		}
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad gateway"))
	})

	log.Println("Reverse proxy server listening on port " + port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("ListenAndServe ("+port+"): ", err)
	}
}
