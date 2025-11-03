package wsserver

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func HandleCheck(w http.ResponseWriter, r *http.Request, no_vms int) {
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	timeout := 10 * time.Second
	retryInterval := 500 * time.Millisecond

	// vms := makeVMs(no_vms)

	for _, vm := range VMS_IP {
		wg.Add(1)
		go func(vmIP, vmName string) {
			defer wg.Done()
			log.Printf("Checking SSH readiness for %s (%s)...", vmName, vmIP)
			err := WaitForSSH(vmIP, 22, timeout, retryInterval)
			if err != nil {
				errChan <- fmt.Errorf("%s: %v", vmName, err)
			} else {
				log.Printf("SSH is ready on %s (%s)", vmName, vmIP)
			}
		}(vm.ip, vm.name)
	}

	wg.Wait()
	close(errChan)

	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		http.Error(w, "Some VMs are not ready:\n"+fmt.Sprintf("%v", errors), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Both VMs are ready."))
}

// func makeVMs(count int) []struct {
// 	ip   string
// 	name string
// } {
// 	vms := make([]struct {
// 		ip   string
// 		name string
// 	}, count)

// 	for i := range count {
// 		vms[i].ip = fmt.Sprintf("172.16.0.%d", 2+i)
// 		vms[i].name = fmt.Sprintf("vm%d", 1+i)
// 	}

// 	return vms
// }
