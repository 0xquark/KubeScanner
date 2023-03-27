package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
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

type ScanConfig struct {
	Targets []ScanTarget
	Ports   []int
	Timeout time.Duration
	TcpOnly bool
	UdpOnly bool
}

func main() {
	config, err := parseArgs()
	if err != nil {
		fmt.Println("Error parsing arguments:", err)
		return
	}

	scanResults := scanTargets(config.Targets, config.TcpOnly, config.UdpOnly, config.Ports, config.Timeout)

	// Print scan results
	printResults(scanResults)
}

func parseArgs() (*ScanConfig, error) {
	var config ScanConfig

	flag.BoolVar(&config.TcpOnly, "tcp", false, "Scan only TCP ports")
	flag.BoolVar(&config.UdpOnly, "udp", false, "Scan only UDP ports")
	flag.Parse()

	if flag.NArg() < 1 {
		return nil, fmt.Errorf("Usage: %s [--tcp|--udp] <host or ip_address or ip_range> [ports...]", os.Args[0])
	}

	targetStr := flag.Arg(0)
	if strings.Contains(targetStr, "-") { // check if target is a range of IP addresses
		ipRange := strings.Split(targetStr, "-")
		startIP := net.ParseIP(ipRange[0])
		endIP := net.ParseIP(ipRange[1])
		if startIP == nil || endIP == nil {
			return nil, fmt.Errorf("Invalid IP address range.")
		}
		for ip := startIP; ip.String() <= endIP.String(); incIP(ip) {
			if ip.To4() == nil {
				fmt.Printf("IPv6 address not supported: %s\n", ip.String())
				continue
			}
			target := ScanTarget{IP: ip}
			config.Targets = append(config.Targets, target)
		}
	} else {
		target := ScanTarget{Host: targetStr}
		if ip := net.ParseIP(target.Host); ip != nil {
			if ip.To4() == nil {
				return nil, fmt.Errorf("IPv6 address not supported.")
			}
			target.IP = ip
		} else {
			// Resolve hostname
			addrs, err := net.LookupHost(target.Host)
			if err != nil {
				return nil, fmt.Errorf("Failed to resolve hostname: %s", target.Host)
			}
			target.IP = net.ParseIP(addrs[0])
			if target.IP.To4() == nil {
				return nil, fmt.Errorf("IPv6 address not supported.")
			}
		}
		config.Targets = append(config.Targets, target)
	}

	if flag.NArg() > 1 {
		for _, portStr := range flag.Args()[1:] {
			port, err := strconv.Atoi(portStr)
			if err != nil {
				return nil, fmt.Errorf("Invalid port number: %s", portStr)
			}
			config.Ports = append(config.Ports, port)
		}
	}

	config.Timeout = 100 * time.Millisecond

	return &config, nil
}

func scanTarget(target ScanTarget, proto string, ports []int, timeout time.Duration, results chan<- ScanResult, wg *sync.WaitGroup) {
	defer wg.Done()
	portsOpen := scanIP(target.IP.String(), proto, ports, timeout)
	if len(portsOpen.TCPPorts) > 0 || len(portsOpen.UDPPorts) > 0 {
		result := ScanResult{
			Host:     target.Host,
			IP:       target.IP,
			TCPPorts: portsOpen.TCPPorts,
			UDPPorts: portsOpen.UDPPorts,
		}
		results <- result
	}
}

func scanTargets(targets []ScanTarget, tcpOnly bool, udpOnly bool, ports []int, timeout time.Duration) []ScanResult {
	var wg sync.WaitGroup
	results := make(chan ScanResult, len(targets))

	for _, target := range targets {
		wg.Add(1)
		go func(target ScanTarget) {
			defer wg.Done()
			switch {
			case !tcpOnly && !udpOnly:
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
			case tcpOnly:
				tcpPortsOpen := scanIP(target.IP.String(), "tcp", ports, timeout)
				if len(tcpPortsOpen.TCPPorts) > 0 {
					result := ScanResult{
						Host:     target.Host,
						IP:       target.IP,
						TCPPorts: tcpPortsOpen.TCPPorts,
					}
					results <- result
				}
			case udpOnly:
				udpPortsOpen := scanIP(target.IP.String(), "udp", ports, timeout)
				if len(udpPortsOpen.UDPPorts) > 0 {
					result := ScanResult{
						Host:     target.Host,
						IP:       target.IP,
						UDPPorts: udpPortsOpen.UDPPorts,
					}
					results <- result
				}
			}
		}(target)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var scanResults []ScanResult
	for result := range results {
		scanResults = append(scanResults, result)
	}

	return scanResults
}

func printResults(results []ScanResult) {
	for _, result := range results {
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

// Define the maximum port number
const maxPort = 65535

// scanIP scans the specified IP address for open TCP and UDP ports.
// If no ports are specified, it scans all ports. Returns a ScanResult
// struct containing the IP address and open TCP and UDP ports.
func scanIP(ip string, proto string, ports []int, timeout time.Duration) ScanResult {
	// Create a channel to collect open ports and a done channel for synchronization
	openPorts := make(chan int)
	done := make(chan struct{})
	defer close(done)

	// Parse the IP address
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		// Invalid IP address
		fmt.Printf("%s is not a valid IP address\n", ip)
		return ScanResult{}
	}

	// Start a goroutine for each specified port
	var wg sync.WaitGroup
	for _, port := range ports {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			if isOpen(ip, port, proto, timeout) {
				fmt.Printf("%s:%d/%s is open\n", ip, port, proto)
				openPorts <- port
			}
		}(port)
	}

	// If no ports are specified, scan all ports
	if len(ports) == 0 {
		for port := 1; port <= maxPort; port++ {
			wg.Add(1)
			go func(port int) {
				defer wg.Done()
				if isOpen(ip, port, proto, timeout) {
					fmt.Printf("%s:%d/%s is open\n", ip, port, proto)
					openPorts <- port
				}
			}(port)
		}
	}

	// Start a goroutine to wait for all other goroutines to finish
	go func() {
		wg.Wait()
		close(openPorts)
	}()

	// Collect the open ports from the channel
	openTCPPorts := []int{}
	openUDPPorts := []int{}
	for port := range openPorts {
		switch proto {
		case "tcp":
			openTCPPorts = append(openTCPPorts, port)
		case "udp":
			openUDPPorts = append(openUDPPorts, port)
		}
	}

	// Create and return a ScanResult struct
	result := ScanResult{
		IP:       ipAddr,
		TCPPorts: openTCPPorts,
		UDPPorts: openUDPPorts,
	}
	return result
}

// isOpen checks if the specified TCP or UDP port is open on the specified IP address.
// Returns true if the port is open, false otherwise.
func isOpen(ip string, port int, proto string, timeout time.Duration) bool {
	conn, err := net.DialTimeout(proto, fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

/*

Should i use interfaces and structs more often like this ?

type PortChecker interface {
	IsOpen(ip string, port int, timeout time.Duration) bool
}

type TCPPortChecker struct{}

func (c *TCPPortChecker) IsOpen(ip string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

type UDPPortChecker struct{}

func (c *UDPPortChecker) IsOpen(ip string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

type Scanner interface {
	Scan(ip string, ports []int, timeout time.Duration) ScanResult
}

type TCPSanner struct{}

func (s *TCPSanner) Scan(ip string, ports []int, timeout time.Duration) ScanResult {
	openPorts := []int{}

	if net.ParseIP(ip) == nil {
		// Invalid IP address
		fmt.Printf("%s is not a valid IP address\n", ip)
		return ScanResult{}
	}

	var wg sync.WaitGroup

	if len(ports) == 0 {
		// Scan all ports
		for port := 1; port <= 65535; port++ {
			wg.Add(1)
			go func(port int) {
				defer wg.Done()
				if isOpen(ip, port, "tcp", timeout) {
					fmt.Printf("%s:%d/tcp is open\n", ip, port)
					openPorts = append(openPorts, port)
				}
			}(port)
		}
	} else {
		// Scan specified ports
		for _, port := range ports {
			wg.Add(1)
			go func(port int) {
				defer wg.Done()
				if isOpen(ip, port, "tcp", timeout) {
					fmt.Printf("%s:%d/tcp is open\n", ip, port)
					openPorts = append(openPorts, port)
				}
			}(port)
		}
	}

	wg.Wait()

	result := ScanResult{
		IP:       net.ParseIP(ip),
		TCPPorts: openPorts,
	}

	return result
}

type UDPSanner struct{}

func (s *UDPSanner) Scan(ip string, ports []int, timeout time.Duration) ScanResult {
	openPorts := []int{}

	if net.ParseIP(ip) == nil {
		// Invalid IP address
		fmt.Printf("%s is not a valid IP address\n", ip)
		return ScanResult{}
	}

	var wg sync.WaitGroup

	if len(ports) == 0 {
		// Scan all ports
		for port := 1; port <= 65535; port++ {
			wg.Add(1)
			go func(port int) {
				defer wg.Done()
				if isOpen(ip, port, "udp", timeout) {
					fmt.Printf("%s:%d/udp is open\n", ip, port)
					openPorts = append(openPorts, port)
				}
			}(port)
		}
	} else {

		// Scan specified ports
		for _, port := range ports {
			wg.Add(1)
			go func(port int) {
				defer wg.Done()
				if isOpen(ip, port, "udp", timeout) {
					fmt.Printf("%s:%d/udp is open\n", ip, port)
					openPorts = append(openPorts, port)
				}
			}(port)
		}
	}
*/
