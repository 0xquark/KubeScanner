package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// 1. Make sure that the IP addresses are valid and IPv4
// 2. We will need to make port scanning concurrent
// 3. We will need to make the port scanning configurable (UDP/TCP)
// 4. We should support also scanning by hostname
// 5. Make sure that the TCP dial timeout is configurable and set to a 100ms by default

type PortScanner struct {
}

func CreatPortScanner() *PortScanner {
	return &PortScanner{}
}

// the int port mapping should be a struct itself so we can define whether it is a UDP or a TCP prt
func (p *PortScanner) ScanRange(startIP, endIP string, ports []int) map[string][]int {
	var scanResult map[string][]int
	start := net.ParseIP(startIP)
	end := net.ParseIP(endIP)
	if start == nil || end == nil {
		fmt.Println("Invalid IP address range.")
		return scanResult
	}
	for ip := start; ip.String() <= end.String(); incIP(ip) {
		scanResult[ip.String()] = p.Scan(ip.String(), ports)
	}
	return scanResult
}

func (p *PortScanner) Scan(ip string, ports []int) []int {
	var openPorts []int
	if len(ports) == 0 {
		// Scan all ports
		for port := 1; port <= 65535; port++ {
			if p.isOpen(ip, port) {
				openPorts = append(openPorts, port)
			}
		}
	} else {
		// Scan specified ports
		for _, port := range ports {
			if isOpen(ip, port) {
				openPorts = append(openPorts, port)
			}
		}
	}
	return openPorts
}

func (p *PortScanner) isOpen(ip string, port int) bool {
	// Check timeout and set it to something configurable (in a kubernetes cluster this can be around 100ms)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

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
