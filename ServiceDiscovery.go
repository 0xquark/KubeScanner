package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
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

	if isMinikube(ip, port) {
		return "minikube"
	}

	if isInsecureAPI(ip, port) {
		return "insecure_api"
	}

	if isKubernetesAPI(ip, port) {
		return "kubernetes_api"
	}

	if isKubeletHTTPS(ip, port) {
		return "kubelet"
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

func isMinikube(ip string, port int) bool {
	// Attempt to connect to the kube-apiserver
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), time.Second*5)
	if err != nil {
		return false
	}
	defer conn.Close()

	// Send a request to the kube-apiserver
	fmt.Fprintf(conn, "GET /version HTTP/1.1\r\nHost: %s\r\n\r\n", ip)

	// Read the response from the kube-apiserver
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return false
	}

	// Check if the response contains the minikube version string
	isMinikube := strings.Contains(string(buf[:n]), "minikube")

	return isMinikube
}

// Check for insecure API Port
func isInsecureAPI(ip string, port int) bool {
	url := fmt.Sprintf("https://%s:%d", ip, port)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	// Disable TLS verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Check if the response contains "Unauthorized"
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return strings.Contains(string(body), "Unauthorized")
}

// Check if a port is serving Kubernetes API
func isKubernetesAPI(ip string, port int) bool {
	// Attempt to connect to the Kubernetes API service
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), time.Second*5)
	if err != nil {
		return false
	}
	defer conn.Close()

	// Send a request to the Kubernetes API service
	fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: %s\r\n\r\n", ip)

	// Read the response from the Kubernetes API service
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return false
	}

	// Check if the response contains the Kubernetes API version string
	isKubernetesAPI := strings.Contains(string(buf[:n]), "kubernetes")

	return isKubernetesAPI
}

// To check if Kubelet is running on the port, we can make a request to the /healthz endpoint of the Kubelet API. If the response status code is 200, then Kubelet is running on the port.
// To check if the HTTPS API allows full mode access, we can make a request to the /pods endpoint of the Kubernetes API using the curl command. If the response contains a list of running pods, then the API allows full mode access.
func isKubeletHTTPS(ip string, port int) bool {
	// Attempt to connect to the Kubelet HTTPS API
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), time.Second*5)
	if err != nil {
		return false
	}
	defer conn.Close()

	// Send a request to the Kubelet HTTPS API
	fmt.Fprintf(conn, "GET /healthz HTTP/1.1\r\nHost: %s\r\n\r\n", ip)

	// Read the response from the Kubelet HTTPS API
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return false
	}

	// Check if the response indicates that the Kubelet is running
	isKubelet := strings.Contains(string(buf[:n]), "ok")

	if isKubelet {
		// Attempt to access the Kubernetes API using the Kubelet's pod IP address
		podIP := strings.Split(ip, ":")[0]
		cmd := exec.Command("kubectl", "--insecure-skip-tls-verify", "--server=https://"+podIP+":10250", "get", "pods", "--all-namespaces")
		output, err := cmd.CombinedOutput()

		// Check if kubectl output indicates that unauthenticated access is available
		isVulnerable := false
		if err != nil {
			if strings.Contains(string(output), "Unauthorized") || strings.Contains(string(output), "authentication required") {
				isVulnerable = true
			}
		} else {
			if strings.TrimSpace(string(output)) != "" {
				isVulnerable = true
			}
		}

		// Check if unauthenticated access is available for pod status and node state
		resp1, err1 := http.Get(fmt.Sprintf("http://%s:%d/api/v1/nodes", ip, 10255))
		resp2, err2 := http.Get(fmt.Sprintf("http://%s:%d/api/v1/pods", ip, 10255))
		if err1 == nil && err2 == nil {
			defer resp1.Body.Close()
			defer resp2.Body.Close()
			body1, _ := ioutil.ReadAll(resp1.Body)
			body2, _ := ioutil.ReadAll(resp2.Body)
			isVulnerable = isVulnerable || (strings.TrimSpace(string(body1)) != "" && strings.TrimSpace(string(body2)) != "")
		}

		if isVulnerable {
			fmt.Printf("Kubelet service on %s:%d is vulnerable to unauthenticated access\n", ip, port)
		}

		return isVulnerable
	}

	return false
}

/* A function to check if a port is serving HTTP and to determine if it's a kube-apiserver which is serving minikube or etcd or other services
func isHTTP(ip string, port int) string {
	// Check if it's an HTTP service
	if strings.HasPrefix(fmt.Sprintf("%d", port), "8") || strings.HasPrefix(fmt.Sprintf("%d", port), "80") || strings.HasPrefix(fmt.Sprintf("%d", port), "443") {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), time.Second*5)
		if err != nil {
			return "unknown"
		}
		defer conn.Close()

		fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: %s\r\n\r\n", ip)
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			return "unknown"
		}

		// Check if it's a kube-apiserver
		if strings.Contains(string(buf[:n]), "kube-apiserver") {
			// Check if it's serving minikube
			if isMinikube(ip, port) {
				return "minikube"
			}
			if isInsecureAPI(ip, port){
				return "insecure-api"
			}
			return "kube-apiserver"
		}

		// Check if it's etcd
		if isEtcd(ip, port) {
			return "etcd"
		}

		return "http"
	}
	return "unknown"
}
*/
