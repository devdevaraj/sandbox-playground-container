package init_app

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

type Network struct {
	Name string `json:"name,omitempty"`
	TAP  string `json:"tap,omitempty"`
	IP   string `json:"ip,omitempty"`
	MAC  string `json:"mac,omitempty"`
}

type Template struct {
	CPU           *int      `json:"cpu,omitempty"`
	SMT           *bool     `json:"smt,omitempty"`
	Multiplier    *int      `json:"multiplier,omitempty"`
	RAM           *int      `json:"ram,omitempty"`
	EnableOverlay *bool     `json:"enable-overlay"`
	OverlaySize   *int      `json:"overlay-size"`
	IsZFS         *bool     `json:"is-zfs,omitempty"`
	ZFSSnapshot   string    `json:"zfs-snapshot"`
	ZFSClonePath  string    `json:"zfs-clone-path"`
	KernelArgs    string    `json:"kernel-args"`
	Kernel        string    `json:"kernel"`
	RootFS        string    `json:"rootfs"`
	Username      *string   `json:"username,omitempty"`
	Network       []Network `json:"network"`
}

type Config struct {
	Bridge    *string    `json:"bridge,omitempty"`
	BridgeIP  *string    `json:"bridge-ip,omitempty"`
	Network   *string    `json:"network,omitempty"`
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

	if cfg.Bridge == nil {
		def := "br0"
		cfg.Bridge = &def
	}

	if cfg.BridgeIP == nil {
		def := "172.16.0.1/24"
		cfg.BridgeIP = &def
	}

	if cfg.Network == nil {
		def := "172.16.0.0/24"
		cfg.Network = &def
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

	for i := range cfg.Templates {
		if len(cfg.Templates[i].Network) == 0 {
			cfg.Templates[i].Network = []Network{
				{
					Name: "eth0",
					TAP:  "tap" + strconv.Itoa(i),
					IP:   "172.16.0." + strconv.Itoa(i+2),
					MAC:  "AA:FC:00:00:00:0" + strconv.Itoa(i+1)},
			}
		}

		// for j := range cfg.Templates[i].Network {
		// 	if cfg.Templates[i].Network[j].Name == nil {
		// 		def := "172.16.0." + strconv.Itoa(i+2)
		// 		cfg.Templates[i].IP = &def
		// 	}
		// }
	}

	return cfg
}
