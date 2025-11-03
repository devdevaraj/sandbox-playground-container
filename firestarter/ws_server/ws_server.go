package wsserver

import (
	"log"
	"net/http"
	"strconv"

	"github.com/devdevaraj/firestarter/init_app"
	"github.com/gorilla/mux"
)

func WebSockerServer(port string, cfg init_app.Config) {
	router := mux.NewRouter()

	router.HandleFunc("/examiner/test/{vm}/{test}", func(w http.ResponseWriter, r *http.Request) {
		ExaminerCheck(w, r)
	})
	for i, conf := range cfg.Templates {
		router.HandleFunc("/vm"+strconv.Itoa(i+1)+"/{session}", func(w http.ResponseWriter, r *http.Request) {
			HandleWebsocket(w, r, *conf.Network[0].IP, "vm"+strconv.Itoa(i+1), cfg.Templates[i])
		})
	}
	router.HandleFunc("/wait-for-vms", func(w http.ResponseWriter, r *http.Request) {
		HandleCheck(w, r, len(cfg.Templates))
	})
	router.HandleFunc("/open-close-port", func(w http.ResponseWriter, r *http.Request) {
		HandlePublicPort(w, r)
	})
	router.HandleFunc("/open-close-port/{id}", func(w http.ResponseWriter, r *http.Request) {
		HandlePublicPort(w, r)
	})

	log.Println("WebSocket server listening on port " + port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
