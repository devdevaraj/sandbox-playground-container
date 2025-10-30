package launcher

import (
	"net"

	"github.com/devdevaraj/firestarter/init_app"
	"github.com/firecracker-microvm/firecracker-go-sdk"
)

func CreateNetworkInterface(network init_app.Network) firecracker.NetworkInterface {
	// fmt.Printf("%s", tapName+ipAddr+gateway+macAddr)
	return firecracker.NetworkInterface{
		StaticConfiguration: &firecracker.StaticNetworkConfiguration{
			HostDevName: network.TAP,
			MacAddress:  network.MAC,
			IPConfiguration: &firecracker.IPConfiguration{
				IPAddr: net.IPNet{
					IP:   net.ParseIP(network.IP),
					Mask: net.CIDRMask(network.Mask, 32),
				},
				Gateway:     net.ParseIP(network.Gateway),
				Nameservers: []string{network.Nameservers.NS1, network.Nameservers.NS2},
				IfName:      network.Name,
			},
		},
	}
}
