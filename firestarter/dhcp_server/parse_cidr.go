package dhcpserver

import "net"

func ParseCIDR(cidr string) (net.IP, net.IPMask, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, nil, err
	}
	return ip, ipNet.Mask, nil
}
