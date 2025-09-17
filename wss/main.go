package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/devdevaraj/wss/examiner"
	"github.com/devdevaraj/wss/handle_check"
	"github.com/devdevaraj/wss/handle_public_port"
	"github.com/devdevaraj/wss/handle_websocket"
	"github.com/gorilla/mux"
)

type Template struct {
	Kernel string `json:"kernel"`
	RootFS string `json:"rootfs"`
}

type Config struct {
	VMs       int        `json:"vms"`
	Templates []Template `json:"templates"`
}

func main() {
	args := os.Args
	data, err := os.ReadFile("/resourses/.configs/" + args[1] + ".json")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to unmarshal data: %v", err)
	}
	go func() {
		router := mux.NewRouter()

		router.HandleFunc("/examiner/test/{vm}/{test}", func(w http.ResponseWriter, r *http.Request) {
			examiner.ExaminerCheck(w, r)
		})
		for i := range cfg.VMs {
			router.HandleFunc("/vm"+strconv.Itoa(i+1)+"/{session}", func(w http.ResponseWriter, r *http.Request) {
				handle_websocket.HandleWebsocket(w, r, "172.16.0."+strconv.Itoa(i+2), "vm"+strconv.Itoa(i+1))
			})
		}
		router.HandleFunc("/wait-for-vms", func(w http.ResponseWriter, r *http.Request) {
			handle_check.HandleCheck(w, r, cfg.VMs)
		})
		router.HandleFunc("/open-close-port", func(w http.ResponseWriter, r *http.Request) {
			handle_public_port.HandlePublicPort(w, r)
		})
		router.HandleFunc("/open-close-port/{id}", func(w http.ResponseWriter, r *http.Request) {
			handle_public_port.HandlePublicPort(w, r)
		})

		// http.Handle("/", router)

		log.Println("WebSocket server listening on port 8080")
		if err := http.ListenAndServe(":8080", router); err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		subdomain := strings.Split(r.Host, ".")[0]
		split := strings.Split(subdomain, "-")
		id := split[len(split)-1]

		reverseProxy, ok := handle_public_port.GetProxy(id)
		if ok {
			reverseProxy.ProxyHandler(w, r)
			return
		}
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad gateway"))
	})

	log.Println("Reverse proxy server listening on port 80")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal("ListenAndServe (80): ", err)
	}
}
