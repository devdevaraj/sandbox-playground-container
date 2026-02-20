package launcher

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/devdevaraj/firestarter/init_app"
)

func StartMicroVM(
	ctx context.Context,
	vmID string,
	disks []init_app.Disk,
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
) error {
	// Configure VM paths
	overlayfsPath := "/root/firecracker/overlayfs/" + vmID + "-overlay.ext4"
	diskPath := "/root/firecracker/disks-" + vmID + "/"
	socketPath := fmt.Sprintf("/tmp/firecracker-%s.sock", vmID)
	logPath := filepath.Join(os.TempDir(), fmt.Sprintf("firecracker-%s.log", vmID))

	// Clean up existing socket
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			log.Printf("Failed to remove existing socket: %v", err)
		}
	}

	// Clean up existing log file
	if _, err := os.Stat(logPath); err == nil {
		if err := os.Remove(logPath); err != nil {
			log.Printf("Failed to remove existing log file: %v", err)
		}
	}

	// Clean up existing FIFOs (kept from original code logic, though maybe not needed if we don't use them)
	fifoFiles := []string{"/tmp/firecracker.out.fifo", "/tmp/firecracker.metrics.fifo"}
	for _, fifo := range fifoFiles {
		if _, err := os.Stat(fifo); err == nil {
			if err := os.Remove(fifo); err != nil {
				log.Printf("Failed to remove existing FIFO %s: %v", fifo, err)
			}
		}
	}

	// Start Firecracker process
	cmd := exec.CommandContext(ctx, "firecracker", "--api-sock", socketPath) //, "--log-path", logPath, "--level", "Debug")
	// Redirect stdout/stderr to a log file or os.Stderr for debugging
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		cmd.Stderr = logFile
		cmd.Stdout = logFile
		// defer logFile.Close() // Don't close immediately, let process write to it? exec.Command handles this?
		// Better to just set it.
	} else {
		log.Printf("Failed to open log file: %v, using os.Stderr", err)
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start firecracker process: %w", err)
	}

	// Wait for socket to be ready
	ready := false
	for range 50 { // Wait up to 5 seconds
		if _, err := os.Stat(socketPath); err == nil {
			ready = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !ready {
		_ = cmd.Process.Kill()
		return fmt.Errorf("firecracker socket %s not ready after 5 seconds", socketPath)
	}

	// Initialize Client
	client := NewFirecrackerClient(socketPath)

	// Prepare Configuration
	cpuCount := defaultInt(cpu, 2)
	ramSize := defaultInt(ram, 2048)
	smtFlag := defaultBool(smt, false)
	overlay := defaultBool(enableOverlay, false)
	zfs := defaultBool(isZFS, false)

	rootDiskPath := rootfsPath
	if zfs {
		rootDiskPath = ZFSPath
	}

	// Machine Config
	machineCfg := MachineConfiguration{
		VcpuCount:  int64(cpuCount),
		MemSizeMib: int64(ramSize),
		Smt:        smtFlag,
	}
	if err := client.PutMachineConfiguration(ctx, machineCfg); err != nil {
		return fmt.Errorf("failed to set machine config: %w", err)
	}

	// Boot Source
	primeryNetwork := getPrimery(network)
	kernelArgStr := fmt.Sprintf("%s ip=%s::%s:%s::%s:off", kernelArgs, *primeryNetwork.IP, primeryNetwork.Gateway, MaskToDotted(primeryNetwork.Mask), primeryNetwork.Name)
	log.Printf("Kernel Args: %s", kernelArgStr)

	// Ensure default args if empty (though original code appended to it)
	finalKernelArgs := defaultString(
		&kernelArgStr,
		"console=ttyS0 reboot=k panic=1 pci=off hostname="+vmID+" overlay_root=vdb init=/sbin/overlay-init ip=172.16.0.2::172.16.0.1:255.255.255.0::eth0:off",
	)

	bootSource := BootSource{
		KernelImagePath: kernelImagePath,
		BootArgs:        finalKernelArgs,
	}
	if err := client.PutBootSource(ctx, bootSource); err != nil {
		return fmt.Errorf("failed to set boot source: %w", err)
	}

	// Drives
	// Root Drive
	rootDrive := Drive{
		DriveID:      "rootfs",
		PathOnHost:   rootDiskPath,
		IsRootDevice: true,
		IsReadOnly:   overlay, // ReadOnly if using overlay
		CacheType:    "Unsafe",
	}
	if err := client.PutDrive(ctx, "rootfs", rootDrive); err != nil {
		return fmt.Errorf("failed to set rootfs drive: %w", err)
	}

	// Additional Disks
	for i, disk := range disks {
		driveID := disk.Name // or fmt.Sprintf("disk%d", i+1) - SDK code used disk.Name as DriveID
		// Warning: SDK code constructed PathOnHost as: diskPath + "disk" + strconv.Itoa(i+1) + ".ext4"
		// And used disk.Name as DriveID.

		d := Drive{
			DriveID:      driveID,
			PathOnHost:   diskPath + "disk" + strconv.Itoa(i+1) + ".ext4",
			IsRootDevice: false,
			IsReadOnly:   disk.IsReadOnly,
			CacheType:    "Unsafe",
		}
		if err := client.PutDrive(ctx, driveID, d); err != nil {
			return fmt.Errorf("failed to set drive %s: %w", driveID, err)
		}
	}

	// Overlay Drive
	if overlay {
		overlayDrive := Drive{
			DriveID:      "overlayfs",
			PathOnHost:   overlayfsPath,
			IsRootDevice: false,
			IsReadOnly:   false,
			CacheType:    "Unsafe",
		}
		if err := client.PutDrive(ctx, "overlayfs", overlayDrive); err != nil {
			return fmt.Errorf("failed to set overlayfs drive: %w", err)
		}
	}

	// Network Interfaces
	nics := CreateNetworkInterface(network)
	for _, nic := range nics {
		if err := client.PutNetworkInterface(ctx, nic.IfaceID, nic); err != nil {
			return fmt.Errorf("failed to set network interface %s: %w", nic.IfaceID, err)
		}
	}

	// MMDS Config (Optional, but good to set if we align with SDK default)
	mmdsConfig := MmdsConfig{
		Version:           "V2",             // Firecracker SDK enables V2
		NetworkInterfaces: []string{"eth0"}, // Usually allowed interfaces
	}
	// SDK set MmdsAddress to 169.254.169.254 and Version to MMDSv2.
	// In manual API, we configure this via /mmds/config.
	_ = client.PutMmdsConfig(ctx, mmdsConfig) // Ignore error if optional or already default

	// Metadata
	metadata := buildMetadata(vmID, network)
	if err := client.PutMmds(ctx, metadata); err != nil {
		log.Printf("Warning: Failed to set metadata: %v", err)
	}

	// Start Instance
	log.Println("Starting Firecracker VM Instance...")
	if err := client.InstanceStart(ctx); err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	// Start a goroutine to wait for the command to finish (it shouldn't unless crashed/shutdown)
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("Firecracker process for %s exited with error: %v", vmID, err)
		} else {
			log.Printf("Firecracker process for %s exited cleanly", vmID)
		}
	}()

	return nil
}

func buildMetadata(vmID string, network []init_app.Network) map[string]interface{} {
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

	return map[string]interface{}{
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
