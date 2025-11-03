package wsserver

import (
	"fmt"
	"log"
	"strconv"

	"github.com/devdevaraj/firestarter/init_app"
)

type VMInfo struct {
	ip   string
	name string
}

var VM_MAP = map[string]string{}
var VMS_IP = []VMInfo{}
var TARGET_CACHE = map[string]string{}

type TestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func findPrimary(vms []init_app.Network) (*init_app.Network, int) {
	for i := range vms {
		if vms[i].IsPrimary {
			return &vms[i], i
		}
	}
	return nil, -1 // not found
}

func PopulateIp(cfg init_app.Config) {
	for i, vm := range cfg.Templates {
		ni, index := findPrimary(vm.Network)
		if index > -1 {
			VM_MAP["vm"+strconv.Itoa(i+1)] = *ni.IP
			VMS_IP = append(VMS_IP, VMInfo{
				ip:   *ni.IP,
				name: fmt.Sprintf("vm%d", 1+i),
			})
			if *vm.EnableIDE {
				log.Printf(strconv.Itoa(*vm.IDEPort))
				TARGET_CACHE["codeserverprt"] = *ni.IP + ":" + strconv.Itoa(*vm.IDEPort)
			}
		}
	}
}
