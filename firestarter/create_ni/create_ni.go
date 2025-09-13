package create_ni

import (
	"net"

	"github.com/firecracker-microvm/firecracker-go-sdk"
)

func CreateNetworkInterface(tapName, ipAddr, gateway, macAddr string) firecracker.NetworkInterface {
	return firecracker.NetworkInterface{
		StaticConfiguration: &firecracker.StaticNetworkConfiguration{
			HostDevName: tapName,
			MacAddress:  macAddr,
			IPConfiguration: &firecracker.IPConfiguration{
				IPAddr: net.IPNet{
					IP:   net.ParseIP(ipAddr),
					Mask: net.CIDRMask(24, 32),
				},
				Gateway:     net.ParseIP(gateway),
				Nameservers: []string{"8.8.8.8", "1.1.1.1"},
				IfName:      "eth0",
			},
		},
	}
}
