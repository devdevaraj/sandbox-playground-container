package wait_for_ssh

import (
	"fmt"
	"net"
	"time"
)

func WaitForSSH(host string, port int, timeout, retryInterval time.Duration) error {
	address := fmt.Sprintf("%s:%d", host, port)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, retryInterval)
		if err == nil {
			conn.Close()
			return nil // SSH service is ready
		}
		time.Sleep(retryInterval)
	}
	return fmt.Errorf("SSH service not available on %s:%d after %s", host, port, timeout)
}
