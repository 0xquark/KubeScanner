package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type ScanTarget struct {
	Host string
	IP   net.IP
}

type ScanResult struct {
	Host     string
	IP       net.IP
	TCPPorts []int
	UDPPorts []int
}

func main() {
	var targets []ScanTarget
	var ports []int
	var timeout time.Duration = 100 * time.Millisecond

	// Parse command line arguments
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <host or ip_address or ip_range> [ports...]\n", os.Args[0])
		return
	} else {
		target := ScanTarget{Host: os.Args[1]}
		if ip := net.ParseIP(target.Host); ip != nil {
			if ip.To4() == nil {
				fmt.Println("IPv6 address not supported.")
				return
			}
			target.IP = ip
		} else {
			// Resolve hostname
			addrs, err := net.LookupHost(target.Host)
			if err != nil {
				fmt.Printf("Failed to resolve hostname: %s\n", target.Host)
				return
			}
			target.IP = net.ParseIP(addrs[0])
			if target.IP.To4() == nil {
				fmt.Println("IPv6 address not supported.")
				return
			}
		}
		targets = append(targets, target)
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
	ipRange := strings.Split(targets[0].Host, "-")
	if len(ipRange) > 1 {
		startIP := net.ParseIP(ipRange[0])
		endIP := net.ParseIP(ipRange[1])
		if startIP == nil || endIP == nil {
			fmt.Println("Invalid IP address range.")
			return
		}
		for ip := startIP; ip.String() <= endIP.String(); incIP(ip) {
			if ip.To4() == nil {
				fmt.Printf("IPv6 address not supported: %s\n", ip.String())
				continue
			}
			target := ScanTarget{IP: ip}
			targets = append(targets, target)
		}
	}

	// Scan targets
	results := make(chan ScanResult)
	for _, target := range targets {
		go func(target ScanTarget) {
			tcpPortsOpen := scanIP(target.IP.String(), "tcp", ports, timeout)
			udpPortsOpen := scanIP(target.IP.String(), "udp", ports, timeout)
			if len(tcpPortsOpen.TCPPorts) > 0 || len(udpPortsOpen.UDPPorts) > 0 {
				result := ScanResult{
					Host:     target.Host,
					IP:       target.IP,
					TCPPorts: tcpPortsOpen.TCPPorts,
					UDPPorts: udpPortsOpen.UDPPorts,
				}
				results <- result
			}
		}(target)
	}

	// Print results
	for i := 0; i < len(targets); i++ {
		result := <-results
		if len(result.TCPPorts) > 0 || len(result.UDPPorts) > 0 {
			fmt.Printf("%s (%s) has the following ports open:\n", result.Host, result.IP.String())
			if len(result.TCPPorts) > 0 {
				fmt.Printf("TCP: %v\n", result.TCPPorts)
			}
			if len(result.UDPPorts) > 0 {
				fmt.Printf("UDP: %v\n", result.UDPPorts)
			}
		} else {
			fmt.Printf("%s (%s) has an empty list of open ports.\n", result.Host, result.IP.String())
		}

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

// Scan IP address for open ports and return a slice of open ports
// Scan IP address for open ports and return a slice of open ports
func scanIP(ip string, proto string, ports []int, timeout time.Duration) ScanResult {
	openTCPPorts := []int{}
	openUDPPorts := []int{}

	if net.ParseIP(ip) == nil {
		// Invalid IP address
		fmt.Printf("%s is not a valid IP address\n", ip)
		return ScanResult{}
	}

	if len(ports) == 0 {
		// Scan all ports
		for port := 1; port <= 65535; port++ {
			go func(port int) {
				if isOpen(ip, port, proto, timeout) {
					fmt.Printf("%s:%d/%s is open\n", ip, port, proto)
					if proto == "tcp" {
						openTCPPorts = append(openTCPPorts, port)
					} else {
						openUDPPorts = append(openUDPPorts, port)
					}
				}
			}(port)
		}
	} else {
		// Scan specified ports
		for _, port := range ports {
			go func(port int) {
				if isOpen(ip, port, proto, timeout) {
					fmt.Printf("%s:%d/%s is open\n", ip, port, proto)
					if proto == "tcp" {
						openTCPPorts = append(openTCPPorts, port)
					} else {
						openUDPPorts = append(openUDPPorts, port)
					}
				}
			}(port)
		}
	}

	// Wait for all goroutines to complete
	time.Sleep(timeout)

	result := ScanResult{
		IP:       net.ParseIP(ip),
		TCPPorts: openTCPPorts,
		UDPPorts: openUDPPorts,
	}

	return result
}

func isOpen(ip string, port int, proto string, timeout time.Duration) bool {
	conn, err := net.DialTimeout(proto, fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
