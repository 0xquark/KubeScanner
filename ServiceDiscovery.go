package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	Port int
	Name string
}

func main() {
	var ipAddr string
	var ports []int
	var services []Service

	// Parse command line arguments
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <ip_address> [ports...]\n", os.Args[0])
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
	} else {
		// Scan all ports
		for port := 1; port <= 65535; port++ {
			ports = append(ports, port)
		}
	}

	// Scan IP address for open ports
	for _, port := range ports {
		if isOpen(ipAddr, port) {
			services = append(services, Service{Port: port})
		}
	}

	// Perform service discovery on open ports
	for i, service := range services {
		services[i].Name = discoverService(ipAddr, service.Port)
	}

	// Print results
	if len(services) == 0 {
		fmt.Println("No open ports found.")
	} else {
		for _, service := range services {
			fmt.Printf("%d/%s\n", service.Port, service.Name)
		}
	}
}

// Perform service discovery on a port
func discoverService(ip string, port int) string {
	if isHTTP(ip, port) {
		return "http"
	}
	if isEtcd(ip, port) {
		return "etcd"
	}
	// Add more discovery functions here
	return "unknown"
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

// Check if a port is serving HTTP
func isHTTP(ip string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), time.Second*5)
	if err != nil {
		return false
	}
	defer conn.Close()

	fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: %s\r\n\r\n", ip)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return false
	}

	return strings.Contains(string(buf[:n]), "HTTP/")
}

// Check if a port is serving ETCD
func isEtcd(ip string, port int) bool {
	// Attempt to connect to the etcd service
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), time.Second*5)
	if err != nil {
		return false
	}
	defer conn.Close()

	// Send a request to the etcd service
	fmt.Fprintf(conn, "GET /version HTTP/1.1\r\nHost: %s\r\n\r\n", ip)

	// Read the response from the etcd service
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return false
	}

	// Check if the response contains the etcd version string
	isEtcd := strings.Contains(string(buf[:n]), "etcdserver")

	if isEtcd {
		// Attempt to access etcd using etcdctl
		cmd := exec.Command("etcdctl", "--endpoints=http://"+ip+":2379", "get", "/", "--prefix", "--keys-only")
		output, err := cmd.CombinedOutput()

		// Check if etcdctl output indicates that anonymous access is available
		isVulnerable := false
		if err != nil {
			if strings.Contains(string(output), "authorization failed") {
				isVulnerable = true
			}
		} else {
			if strings.TrimSpace(string(output)) != "" {
				isVulnerable = true
			}
		}

		if isVulnerable {
			fmt.Printf("Etcd service on %s:%d is vulnerable to anonymous access\n", ip, port)
		}

		return isVulnerable
	}

	return false
}
