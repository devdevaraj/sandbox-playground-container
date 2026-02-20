package launcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

// FirecrackerClient handles communication with the Firecracker socket
type FirecrackerClient struct {
	httpClient *http.Client
	socketPath string
}

// NewFirecrackerClient creates a new client for the given socket path
func NewFirecrackerClient(socketPath string) *FirecrackerClient {
	return &FirecrackerClient{
		socketPath: socketPath,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", socketPath)
				},
			},
		},
	}
}

// MachineConfiguration defines the VM resources
type MachineConfiguration struct {
	VcpuCount  int64 `json:"vcpu_count"`
	MemSizeMib int64 `json:"mem_size_mib"`
	Smt        bool  `json:"smt"`
}

// BootSource defines the kernel and boot arguments
type BootSource struct {
	KernelImagePath string `json:"kernel_image_path"`
	BootArgs        string `json:"boot_args"`
}

// Drive defines a block device
type Drive struct {
	DriveID      string `json:"drive_id"`
	PathOnHost   string `json:"path_on_host"`
	IsRootDevice bool   `json:"is_root_device"`
	IsReadOnly   bool   `json:"is_read_only"`
	CacheType    string `json:"cache_type,omitempty"`
}

// NetworkInterface defines a network interface
type NetworkInterface struct {
	IfaceID     string `json:"iface_id"`
	GuestMac    string `json:"guest_mac,omitempty"`
	HostDevName string `json:"host_dev_name"`
}

// MmdsConfig defines MMDS configuration
type MmdsConfig struct {
	Version           string   `json:"version"`
	NetworkInterfaces []string `json:"network_interfaces,omitempty"`
}

// InstanceAction defines an action to perform on the instance
type InstanceAction struct {
	ActionType string `json:"action_type"`
}

// makeRequest acts as a helper to make HTTP requests to the unix socket
func (c *FirecrackerClient) makeRequest(ctx context.Context, method, endpoint string, payload interface{}) error {
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			return fmt.Errorf("failed to encode payload: %w", err)
		}
	}

	// The URL host doesn't matter for unix socket, but scheme must be http
	url := "http://localhost" + endpoint

	req, err := http.NewRequestWithContext(ctx, method, url, &body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody bytes.Buffer
		_, _ = errBody.ReadFrom(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, errBody.String())
	}

	return nil
}

func (c *FirecrackerClient) PutMachineConfiguration(ctx context.Context, config MachineConfiguration) error {
	return c.makeRequest(ctx, http.MethodPut, "/machine-config", config)
}

func (c *FirecrackerClient) PutBootSource(ctx context.Context, config BootSource) error {
	return c.makeRequest(ctx, http.MethodPut, "/boot-source", config)
}

func (c *FirecrackerClient) PutDrive(ctx context.Context, driveID string, config Drive) error {
	return c.makeRequest(ctx, http.MethodPut, fmt.Sprintf("/drives/%s", driveID), config)
}

func (c *FirecrackerClient) PutNetworkInterface(ctx context.Context, ifaceID string, config NetworkInterface) error {
	return c.makeRequest(ctx, http.MethodPut, fmt.Sprintf("/network-interfaces/%s", ifaceID), config)
}

func (c *FirecrackerClient) PutMmdsConfig(ctx context.Context, config MmdsConfig) error {
	return c.makeRequest(ctx, http.MethodPut, "/mmds/config", config)
}

func (c *FirecrackerClient) PutMmds(ctx context.Context, metadata interface{}) error {
	return c.makeRequest(ctx, http.MethodPut, "/mmds", metadata)
}

func (c *FirecrackerClient) InstanceStart(ctx context.Context) error {
	action := InstanceAction{
		ActionType: "InstanceStart",
	}
	return c.makeRequest(ctx, http.MethodPut, "/actions", action)
}
