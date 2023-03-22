package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	var ipAddr string
	var ports []int

	// Parse command line arguments
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <ip_address or ip_range> [ports...]\n", os.Args[0])
		return
	} else {
		ipAddr = os.Args[1]
	}

	if len(os.Args) > 2 {
		for _, portStr := range os.Args[2:] {
			port, err := strconv.Atoi(portStr)
			if err != nil {
				fmt.Printf("Invalid port number: %s\n", portStr)
				return
			}
			ports = append(ports, port)
		}
	}

	// Check if the target is a range of IP addresses
	ipRange := strings.Split(ipAddr, "-")
	if len(ipRange) > 1 {
		startIP := net.ParseIP(ipRange[0])
		endIP := net.ParseIP(ipRange[1])
		if startIP == nil || endIP == nil {
			fmt.Println("Invalid IP address range.")
			return
		}
		for ip := startIP; ip.String() <= endIP.String(); incIP(ip) {
			scanIP(ip.String(), ports)
		}
	} else {
		// Single IP address
		scanIP(ipAddr, ports)
	}
}

// Increment IP address
func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		if ip[j] < 255 {
			ip[j]++
			break
		}
		ip[j] = 0
	}
}

// Scan IP address for open ports
func scanIP(ip string, ports []int) {
	if len(ports) == 0 {
		// Scan all ports
		for port := 1; port <= 65535; port++ {
			if isOpen(ip, port) {
				fmt.Printf("%s:%d is open\n", ip, port)
			}
		}
	} else {
		// Scan specified ports
		for _, port := range ports {
			if isOpen(ip, port) {
				fmt.Printf("%s:%d is open\n", ip, port)
			}
		}
	}
}

// Check if port is open on IP address
func isOpen(ip string, port int) bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
