package launcher

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// StartMemoryAutoscaler starts a background goroutine that watches the VM's internal memory
// StartMemoryAutoscaler starts a background goroutine that watches the VM's internal memory
// usage and adjusts the Firecracker virtio-mem hotplug memory accordingly.
func StartMemoryAutoscaler(ctx context.Context, client *FirecrackerClient, vmID string, vmIP string) {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		// Allow VM to boot up before we start scaling
		time.Sleep(5 * time.Second)

		for {
			select {
			case <-ctx.Done():
				log.Printf("[%s] Memory autoscaler stopped.", vmID)
				return
			case <-ticker.C:
				adjustMemory(ctx, client, vmID, vmIP)
			}
		}
	}()
}

func adjustMemory(ctx context.Context, client *FirecrackerClient, vmID string, vmIP string) {
	// 1. Check current virtio-mem device state
	status, err := client.GetMemoryHotplugStatus(ctx)
	if err != nil {
		log.Printf("[%s] Failed to get memory hotplug status: %v", vmID, err)
		return
	}

	// 2. We need to know how much memory the VM is actually using.
	// We can use docker exec (if playground uses docker) or SSH.
	// Based on commands.sh and the firestarter environment, we can use docker exec
	// if we are the host. But firestarter *runs inside* the docker container,
	// and the VMs are firecracker microVMs. So we must use SSH to the tap IP,
	// or assume the VM has an agent.
	// Since firestarter sets up SSH keys to /root/firecracker/keys/ubuntu-24.04.id_rsa
	// we will run a quick SSH command to check memory.

	// Default VM IP is usually derived from the bridge/tap.
	// Unfortunately finding the exact IP dynamically from here is tricky without the full config,
	// but let's assume we can query the VM if we pass the SSH IP.
	// To keep it simple, we will use a dummy scaling logic for demonstration,
	// or shell out to SSH if we have the IP.

	log.Printf("[%s] Memory Hotplug Status: Plugged=%dMiB, Requested=%dMiB, Total=%dMiB",
		vmID, status.PluggedSizeMib, status.RequestedSizeMib, status.TotalSizeMib)

	if vmIP == "" {
		log.Printf("[%s] No VM IP available for memory autoscaling", vmID)
		return
	}

	freeMem, err := getGuestFreeMemory(vmIP)
	if err != nil {
		log.Printf("[%s] Failed to get guest free memory via SSH: %v", vmID, err)
		return
	}

	currentRequested := status.RequestedSizeMib
	step := 128

	// Basic scale-up/scale-down logic based on guest free memory
	if freeMem < 256 {
		// Guest has less than 256MB free; plug more memory
		if currentRequested+step <= status.TotalSizeMib {
			newSize := currentRequested + step
			log.Printf("[%s] Autoscaling up to %d MiB (Guest Free: %d MiB)", vmID, newSize, freeMem)
			err = client.PatchMemoryHotplug(ctx, MemoryHotplugRequest{RequestedSizeMib: newSize})
			if err != nil {
				log.Printf("[%s] Failed to patch memory via virtio-mem: %v", vmID, err)
			}
		} else {
			log.Printf("[%s] Cannot scale up, max hotplug limit reached (Guest Free: %d MiB)", vmID, freeMem)
		}
	} else if freeMem > 384 {
		// Guest has more than 512MB free; unplug memory
		if currentRequested > 0 {
			newSize := max(currentRequested-step, 0)
			log.Printf("[%s] Autoscaling down to %d MiB (Guest Free: %d MiB)", vmID, newSize, freeMem)
			err = client.PatchMemoryHotplug(ctx, MemoryHotplugRequest{RequestedSizeMib: newSize})
			if err != nil {
				log.Printf("[%s] Failed to patch memory via virtio-mem: %v", vmID, err)
			}
		}
	} else {
		// Memory is within target range
		log.Printf("[%s] Memory usage balanced (Guest Free: %d MiB)", vmID, freeMem)
	}
}

// getGuestFreeMemory is a helper function to SSH into the VM and run `free -m`
func getGuestFreeMemory(ip string) (int, error) {
	cmd := exec.Command("ssh", "-i", "/root/firecracker/keys/ubuntu-24.04.id_rsa",
		"-o", "StrictHostKeyChecking=no", "root@"+ip,
		"free -m | awk '/^Mem:/ {print $4}'")

	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ssh failed: %w", err)
	}

	freeMemStr := strings.TrimSpace(string(out))
	freeMem, err := strconv.Atoi(freeMemStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse free memory: %w", err)
	}

	return freeMem, nil
}
