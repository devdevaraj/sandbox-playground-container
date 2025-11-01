package launcher

import (
	"net"

	"github.com/devdevaraj/firestarter/init_app"
	"github.com/firecracker-microvm/firecracker-go-sdk"
)

// func CreateNetworkInterface(networks []init_app.Network) []firecracker.NetworkInterface {
// 	nics := make([]firecracker.NetworkInterface, 0, len(networks))
// 	for _, network := range networks {
// 		nic := firecracker.NetworkInterface{
// 			StaticConfiguration: &firecracker.StaticNetworkConfiguration{
// 				HostDevName: network.TAP,
// 				MacAddress:  network.MAC,
// 			},
// 			CNIConfiguration: nil,
// 		}
// 		nics = append(nics, nic)
// 	}
// 	return nics
// }

func CreateNetworkInterface(networks []init_app.Network) []firecracker.NetworkInterface {
	nics := make([]firecracker.NetworkInterface, 0, len(networks))
	for i, network := range networks {
		var nic firecracker.NetworkInterface
		if i == 0 && network.IP != nil && *network.IP != "" {
			nic = firecracker.NetworkInterface{
				StaticConfiguration: &firecracker.StaticNetworkConfiguration{
					HostDevName: network.TAP,
					MacAddress:  network.MAC,
					IPConfiguration: &firecracker.IPConfiguration{
						IPAddr: net.IPNet{
							IP:   net.ParseIP(*network.IP),
							Mask: net.CIDRMask(network.Mask, 32),
						},
						Gateway:     net.ParseIP(network.Gateway),
						Nameservers: []string{network.Nameservers.NS1, network.Nameservers.NS2},
						IfName:      network.Name,
					},
				},
			}
		} else {
			nic = firecracker.NetworkInterface{
				StaticConfiguration: &firecracker.StaticNetworkConfiguration{
					HostDevName: network.TAP,
					MacAddress:  network.MAC,
				},
				CNIConfiguration: nil,
			}
		}
		nics = append(nics, nic)
	}
	return nics
}
