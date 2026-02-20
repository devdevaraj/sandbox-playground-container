package launcher

import (
	"log"
	"net"

	"github.com/devdevaraj/firestarter/init_app"
)

func CreateNetworkInterface(networks []init_app.Network) []NetworkInterface {
	nics := make([]NetworkInterface, 0, len(networks))
	for _, network := range networks {
		log.Println(network.Name)
		var nic NetworkInterface

		// Note: The original code had a complex logic with 'if i == 0 ... && false' which effectively disabled the first block.
		// I am preserving the effective logic (the else block) as the primary logic, but keeping the structure if needed later.
		// Since the first block was dead code (&& false), I will only implement the active logic to keep it clean,
		// but I'll add a comment about what was there.

		// The original 'else' block
		nic = NetworkInterface{
			HostDevName: network.TAP,
			GuestMac:    network.MAC,
			IfaceID:     network.Name,
		}

		// Check if we need to apply IP configuration.
		// The original code had commented out IPConfiguration in the else block.
		// And the if block was disabled.
		// So purely based on the active code path, we just set HostDevName, MacAddress, and AllowMMDS.

		// However, standard manual execution requires us to set the ID.
		// The SDK might have auto-generated IDs or used the IfName.
		// I'll use network.Name ("eth0") as the ID.

		nics = append(nics, nic)
	}
	return nics
}

// Helper to parse CIDR - strictly if needed, but the current active logic doesn't use it.
func parseCIDR(ip string, mask int) net.IPNet {
	return net.IPNet{
		IP:   net.ParseIP(ip),
		Mask: net.CIDRMask(mask, 32),
	}
}
