package wsserver

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/devdevaraj/firestarter/init_app"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/ssh"
)

func WebSockerServer(port string, cfg init_app.Config) {
	router := mux.NewRouter()

	PopulateIp(cfg)

	key, err := os.ReadFile("/root/firecracker/keys/ubuntu-24.04.id_rsa")
	if err != nil {
		log.Fatalf("Unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Unable to parse private key: %v", err)
	}

	router.HandleFunc("/examiner/test/{vm}/{test}", func(w http.ResponseWriter, r *http.Request) {
		ExaminerCheck(w, r)
	})
	// for i, conf := range cfg.Templates {
	// 	router.HandleFunc("/vm"+strconv.Itoa(i+1)+"/{session}", func(w http.ResponseWriter, r *http.Request) {
	// 		HandleWebsocket(w, r, *conf.Network[0].IP, "vm"+strconv.Itoa(i+1), cfg.Templates[i])
	// 	})
	// }

	router.HandleFunc("/resize/{vm}/{session}", handleResize)

	for i, conf := range cfg.Templates {
		router.HandleFunc("/vm"+strconv.Itoa(i+1)+"/{session}", func(w http.ResponseWriter, r *http.Request) {
			handleWebSocket(w, r, *conf.Network[0].IP, "22", "vm"+strconv.Itoa(i+1), cfg.Templates[i], signer)
		})
	}

	router.HandleFunc("/wait-for-vms", func(w http.ResponseWriter, r *http.Request) {
		HandleCheck(w, r, len(cfg.Templates))
	})

	router.HandleFunc("/open-close-port", HandlePublicPort)

	router.HandleFunc("/open-close-port/{id}", HandlePublicPort)

	log.Println("WebSocket server listening on port " + port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
