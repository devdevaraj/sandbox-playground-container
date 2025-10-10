package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/devdevaraj/firestarter/configure_network"
	startmicro_vm "github.com/devdevaraj/firestarter/start_microvm"
)

type Template struct {
	CPU           *int   `json:"cpu,omitempty"`
	SMT           *bool  `json:"smt,omitempty"`
	Multiplier    *int   `json:"multiplier,omitempty"`
	RAM           *int   `json:"ram,omitempty"`
	EnableOverlay *bool  `json:"enable-overlay"`
	IsZFS         *bool  `json:"is-zfs"`
	KernelArgs    string `json:"kernel-args"`
	Kernel        string `json:"kernel"`
	RootFS        string `json:"rootfs"`
}

type Config struct {
	VMs       int        `json:"vms"`
	Templates []Template `json:"templates"`
}

func main() {
	args := os.Args
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	data, err := os.ReadFile("/resourses/.configs/" + args[1] + ".json")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to unmarshal data: %v", err)
	}

	configure_network.ConfigureNetWork("172.16.0.1/24", "172.16.0.0/24", cfg.VMs)
	for i := range cfg.VMs {
		go startmicro_vm.StartMicroVM(
			ctx,
			"vm"+strconv.Itoa(i+1),
			"tap"+strconv.Itoa(i),
			"172.16.0."+strconv.Itoa(i+2),
			"172.16.0.1",
			"AA:FC:00:00:00:0"+strconv.Itoa(i+1),
			cfg.Templates[i].KernelArgs,
			cfg.Templates[i].Kernel,
			cfg.Templates[i].RootFS,
			cfg.Templates[i].CPU,
			cfg.Templates[i].SMT,
			cfg.Templates[i].RAM,
			cfg.Templates[i].EnableOverlay,
			cfg.Templates[i].IsZFS,
			args[2],
		)
	}
	select {}
}
