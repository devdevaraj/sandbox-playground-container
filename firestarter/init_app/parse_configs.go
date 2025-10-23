package init_app

import (
	"encoding/json"
	"log"
	"os"
)

type Template struct {
	CPU           *int    `json:"cpu,omitempty"`
	SMT           *bool   `json:"smt,omitempty"`
	Multiplier    *int    `json:"multiplier,omitempty"`
	RAM           *int    `json:"ram,omitempty"`
	EnableOverlay *bool   `json:"enable-overlay"`
	OverlaySize   *int    `json:"overlay-size"`
	IsZFS         *bool   `json:"is-zfs,omitempty"`
	ZFSSnapshot   string  `json:"zfs-snapshot"`
	ZFSClonePath  string  `json:"zfs-clone-path"`
	KernelArgs    string  `json:"kernel-args"`
	Kernel        string  `json:"kernel"`
	RootFS        string  `json:"rootfs"`
	Username      *string `json:"username,omitempty"`
}

type Config struct {
	VMs       int        `json:"vms"`
	Templates []Template `json:"templates"`
}

func ParseConfigs(file string) Config {
	data, err := os.ReadFile("/resourses/.configs/" + file + ".json")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to unmarshal data: %v", err)
	}

	for i := range cfg.Templates {
		if cfg.Templates[i].IsZFS == nil {
			def := false
			cfg.Templates[i].IsZFS = &def
		}
	}

	for i := range cfg.Templates {
		if cfg.Templates[i].EnableOverlay == nil {
			def := false
			cfg.Templates[i].EnableOverlay = &def
		}
	}

	for i := range cfg.Templates {
		if cfg.Templates[i].Username == nil {
			def := "root"
			cfg.Templates[i].Username = &def
		}
	}

	return cfg
}
