package gorgonzola

import (
	"net"
)

// Holds the local address
var localAddress string

func init() {
	localAddress = guessLocalAddress()
}

// Guess the local IP
func guessLocalAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic("No network-interfaces found!")
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	// no IP found, this cannot work!
	panic("No IP found!")
}

// Map the localhost (IPv4, IPv6) address to the exposed IP
func remap(address string) string {
	if address == "127.0.0.1" || address == "::1" {
		return localAddress
	}
	return address
}

// Clear the port information from ip:port
func stripAddress(remoteAddress string) (string, error) {
	ip, _, err := net.SplitHostPort(remoteAddress)
	return remap(ip), err
}
