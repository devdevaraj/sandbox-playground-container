package startmicro_vm

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/devdevaraj/firestarter/create_ni"
	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

func StartMicroVM(
	ctx context.Context,
	vmID,
	tapName,
	ipAddr,
	gateway,
	macAddr string,
	kernelImagePath string,
	rootfsPath string,
	cpu *int,
	ram *int,
) {
	// Configure VM
	overlayfsPath := "/root/firecracker/overlayfs/" + vmID + "-overlay.ext4"
	socketPath := fmt.Sprintf("/tmp/firecracker-%s.sock", vmID)

	// Check if socket exists and remove it
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			log.Fatalf("Failed to remove existing socket: %v", err)
		}
	}

	// Clean up existing FIFOs
	fifoFiles := []string{"/tmp/firecracker.out.fifo", "/tmp/firecracker.metrics.fifo"}
	for _, fifo := range fifoFiles {
		if _, err := os.Stat(fifo); err == nil {
			if err := os.Remove(fifo); err != nil {
				log.Fatalf("Failed to remove existing FIFO %s: %v", fifo, err)
			}
			log.Printf("Removed existing FIFO: %s", fifo)
		}
	}

	cpuCount := defaultInt(cpu, 2)
	ramSize := defaultInt(ram, 2048)

	cfg := firecracker.Config{
		SocketPath:      socketPath,
		KernelImagePath: kernelImagePath,
		KernelArgs:      "console=ttyS0 reboot=k panic=1 pci=off hostname=" + vmID + " overlay_root=vdb init=/sbin/overlay-init",
		Drives: []models.Drive{
			{
				DriveID:      firecracker.String("rootfs"),
				PathOnHost:   firecracker.String(rootfsPath),
				CacheType:    firecracker.String(models.DriveCacheTypeUnsafe),
				IsRootDevice: firecracker.Bool(true),
				IsReadOnly:   firecracker.Bool(true),
				RateLimiter:  nil,
			},
			{
				DriveID:      firecracker.String("overlayfs"),
				PathOnHost:   firecracker.String(overlayfsPath),
				CacheType:    firecracker.String(models.DriveCacheTypeUnsafe),
				IsRootDevice: firecracker.Bool(false),
				IsReadOnly:   firecracker.Bool(false),
				RateLimiter:  nil,
			},
		},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(int64(cpuCount)),
			MemSizeMib: firecracker.Int64(int64(ramSize)),
			Smt:        firecracker.Bool(true),
		},
		NetworkInterfaces: []firecracker.NetworkInterface{
			create_ni.CreateNetworkInterface(tapName, ipAddr, gateway, macAddr),
		},
		VMID:     vmID,
		LogLevel: "Debug",
		LogPath:  filepath.Join(os.TempDir(), fmt.Sprintf("firecracker-%s.log", vmID)),
	}

	// Let's use a simpler approach without FIFOs
	cmd := firecracker.VMCommandBuilder{}.
		WithBin("firecracker").
		WithSocketPath(socketPath).
		// WithStdin(os.Stdin).
		// WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		Build(ctx)

	m, err := firecracker.NewMachine(ctx, cfg, firecracker.WithProcessRunner(cmd))
	if err != nil {
		log.Fatalf("Failed to create machine: %v", err)
	}

	// Start the VM
	log.Println("Starting Firecracker VM...")
	go func() {
		if err := m.Start(ctx); err != nil {
			log.Fatalf("Failed to start machine: %v", err)
		}
	}()
}

func defaultInt(ptr *int, defaultVal int) int {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}
