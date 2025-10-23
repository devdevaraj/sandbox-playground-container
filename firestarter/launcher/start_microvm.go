package launcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	kernelArgs string,
	kernelImagePath string,
	rootfsPath string,
	cpu *int,
	smt *bool,
	ram *int,
	enableOverlay *bool,
	isZFS *bool,
	ZFSPath string,
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
	kernelArgsString := defaultString(&kernelArgs, "console=ttyS0 reboot=k panic=1 pci=off hostname="+vmID+" overlay_root=vdb init=/sbin/overlay-init")
	smtFlag := defaultBool(smt, false)
	overlay := defaultBool(enableOverlay, false)
	zfs := defaultBool(isZFS, false)

	path := rootfsPath
	if zfs {
		path = ZFSPath
	}

	cfg := firecracker.Config{
		SocketPath:      socketPath,
		KernelImagePath: kernelImagePath,
		KernelArgs:      kernelArgsString,
		Drives: func() []models.Drive {
			drives := []models.Drive{
				{
					DriveID:      firecracker.String("rootfs"),
					PathOnHost:   firecracker.String(path),
					CacheType:    firecracker.String(models.DriveCacheTypeUnsafe),
					IsRootDevice: firecracker.Bool(true),
					IsReadOnly:   firecracker.Bool(overlay),
					RateLimiter:  nil,
				},
			}
			if overlay {
				drives = append(drives, models.Drive{
					DriveID:      firecracker.String("overlayfs"),
					PathOnHost:   firecracker.String(overlayfsPath),
					CacheType:    firecracker.String(models.DriveCacheTypeUnsafe),
					IsRootDevice: firecracker.Bool(false),
					IsReadOnly:   firecracker.Bool(false),
					RateLimiter:  nil,
				})
			}
			return drives
		}(),
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(int64(cpuCount)),
			MemSizeMib: firecracker.Int64(int64(ramSize)),
			Smt:        firecracker.Bool(smtFlag),
		},
		NetworkInterfaces: []firecracker.NetworkInterface{
			CreateNetworkInterface(tapName, ipAddr, gateway, macAddr),
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

func defaultString(ptr *string, defaultVal string) string {
	if ptr != nil && *ptr != "" {
		return *ptr
	}
	return defaultVal
}

func defaultBool(ptr *bool, defaultVal bool) bool {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}
