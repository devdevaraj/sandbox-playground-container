package launcher

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/devdevaraj/firestarter/init_app"
	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

func StartMicroVM(
	ctx context.Context,
	vmID string,
	network []init_app.Network,
	kernelArgs string,
	kernelImagePath string,
	rootfsPath string,
	cpu *int,
	smt *bool,
	ram *int,
	enableOverlay *bool,
	isZFS *bool,
	ZFSPath string,
) (*firecracker.Machine, error) {
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
	primeryNetwork := getPrimery(network)

	kernelArg := fmt.Sprintf("%s ip=%s::%s:%s::%s:off", kernelArgs, *primeryNetwork.IP, primeryNetwork.Gateway, MaskToDotted(primeryNetwork.Mask), primeryNetwork.Name)
	log.Printf("%s", kernelArg)
	kernelArgsString := defaultString(
		&kernelArg,
		"console=ttyS0 reboot=k panic=1 pci=off hostname="+vmID+" overlay_root=vdb init=/sbin/overlay-init ip=172.16.0.2::172.16.0.1:255.255.255.0::eth0:off",
	)
	smtFlag := defaultBool(smt, false)
	overlay := defaultBool(enableOverlay, false)
	zfs := defaultBool(isZFS, false)

	path := rootfsPath
	if zfs {
		path = ZFSPath
	}

	networkMetadata := make(map[string]interface{})
	for _, net := range network {
		networkMetadata[net.Name] = map[string]interface{}{
			"is-primary": fmt.Sprintf("%v", net.IsPrimary),
			"name":       net.Name,
			"tap":        net.TAP,
			"bridge":     net.Bridge,
			"ip":         net.IP,
			"mask":       fmt.Sprintf("%d", net.Mask),
			"gateway":    net.Gateway,
			"mac":        net.MAC,
			"ns1":        net.Nameservers.NS1,
			"ns2":        net.Nameservers.NS2,
		}
	}

	var userDataBuilder strings.Builder
	userDataBuilder.WriteString("#cloud-config\n")
	userDataBuilder.WriteString(fmt.Sprintf("hostname: %s\n", vmID))
	userDataBuilder.WriteString("manage_etc_hosts: true\n")
	userDataBuilder.WriteString("users:\n")
	userDataBuilder.WriteString("  - name: ubuntu\n")
	userDataBuilder.WriteString("    ssh_authorized_keys:\n")
	userDataBuilder.WriteString("      - ssh-rsa AAAAB3NzaC1yc2E... your-public-key\n")
	userDataBuilder.WriteString("    sudo: ALL=(ALL) NOPASSWD:ALL\n")
	userDataBuilder.WriteString("    groups: sudo\n")
	userDataBuilder.WriteString("    shell: /bin/bash\n")
	userDataBuilder.WriteString("\n")
	userDataBuilder.WriteString("# Download and run network configuration script\n")
	userDataBuilder.WriteString("runcmd:\n")
	userDataBuilder.WriteString("  - curl -o /tmp/configure-network.sh http://169.254.169.254/latest/user-data/network-script\n")
	userDataBuilder.WriteString("  - chmod +x /tmp/configure-network.sh\n")
	userDataBuilder.WriteString("  - /tmp/configure-network.sh\n")
	userDataBuilder.WriteString(fmt.Sprintf("  - echo \"VM %s initialized\" > /tmp/cloud-init-done\n", vmID))

	metadata := map[string]interface{}{
		"latest": map[string]interface{}{
			"meta-data": map[string]interface{}{
				"instance-id":    vmID,
				"local-hostname": vmID,
				"public-keys": map[string]interface{}{
					"0": "ssh-rsa AAAAB3NzaC1yc2E... your-public-key",
				},
				"network": networkMetadata,
			},
			"user-data": userDataBuilder.String(),
		},
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
		NetworkInterfaces: CreateNetworkInterface(network),
		VMID:              vmID,
		LogLevel:          "Debug",
		LogPath:           filepath.Join(os.TempDir(), fmt.Sprintf("firecracker-%s.log", vmID)),
		MmdsVersion:       firecracker.MMDSv2,
		MmdsAddress:       net.ParseIP("169.254.169.254"),
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

		log.Println("Setting MMDS metadata...")
		if err := m.SetMetadata(ctx, metadata); err != nil {
			log.Printf("Warning: Failed to set metadata: %v", err)
		}
	}()

	return m, nil
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

func getPrimery(networks []init_app.Network) *init_app.Network {
	for _, cfg := range networks {
		if cfg.IsPrimary {
			return &cfg
		}
	}
	return nil
}

func MaskToDotted(mask int) string {
	m := net.CIDRMask(mask, 32)
	return net.IP(m).String()
}
