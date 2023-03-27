package main

import (
	"fmt"
	"regexp"
)

// type Service struct {
// 	Port                      int
// 	Address                   string
// 	TransportProtocol         string // tcp, udp
// 	SessionLayerProtocol      string // tls, ssh
// 	PresentationLayerProtocol string // http
// 	ApplicationLayerProtocol  string // mysql, redis, etcd
// }

type SessionLayerProtocolDiscovery interface {
	SessionLayerDiscover(ipAddr string, port int, transportProtocol string) string
}

type PresentationLayerProtocolDiscovery interface {
	PresentationLayerDiscover(ipAddr string, port int, sessionLayerProtocol string) string
}

type ApplicationLayerProtocolDiscovery interface {
	ApplicationLayerDiscover(ipAddr string, port int, presentationLayerProtocol string) string
}

type PresentationLayerProtocol string

const (
	HTTP PresentationLayerProtocol = "http"
)

type iSessionHandler interface {
	Connect() error
	Destory() error
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	GetHost() string
	GetPort() int
}

type iPresentationDiscoveryResult interface {
	Protocol() PresentationLayerProtocol
	GetIsDetected() bool
	GetProperties() map[string]interface{}
}

type PresentationLayerDiscovery interface {
	Protocol() PresentationLayerProtocol
	Discover(sessionHandler iSessionHandler) (iPresentationDiscoveryResult, error)
}

// For HTTP example implementation of the PresentationLayerDiscovery interface
type HttpDiscovery struct {
}

func (d *HttpDiscovery) Protocol() PresentationLayerProtocol {
	return HTTP
}

type SessionHandler struct {
	IP                string
	Port              string
	TransportProtocol string
}

func (s *SessionHandler) Connect() error {
	// Connect to IP:Port
	return nil
}

type HttpDiscoveryResult struct {
	IsDetected bool
	Properties map[string]interface{}
}

// GetProperties implements iPresentationDiscoveryResult
func (hh *HttpDiscoveryResult) GetProperties() map[string]interface{} {
	return hh.Properties
}

// IsDetected implements iPresentationDiscoveryResult
func (hh *HttpDiscoveryResult) GetIsDetected() bool {
	return hh.IsDetected
}

// Protocol implements iPresentationDiscoveryResult
func (*HttpDiscoveryResult) Protocol() PresentationLayerProtocol {
	return HTTP
}

func (d *HttpDiscovery) Discover(sessionHandler iSessionHandler) (iPresentationDiscoveryResult, error) {
	// Connect to sessionHandler
	err := sessionHandler.Connect()
	if err != nil {
		return nil, err
	}
	defer sessionHandler.Destory()

	// Try to write an HTTP request to sessionHandler
	_, err = sessionHandler.Write([]byte(fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\n\r\n", sessionHandler.GetHost())))
	if err != nil {
		return nil, err
	}

	// Read response from sessionHandler
	headerBuf := make([]byte, 1024)
	_, err = sessionHandler.Read(headerBuf)
	if err != nil {
		return nil, err
	}

	r := &HttpDiscoveryResult{
		IsDetected: false,
		Properties: make(map[string]interface{}),
	}

	// Write regexp to parse HTTP response header and extract version
	re := regexp.MustCompile(`HTTP\/(\d+\.\d+) \d+ .+\r\n`)
	match := re.FindSubmatch(headerBuf)
	if match != nil && len(match) > 1 {
		r.IsDetected = true
		r.Properties["version"] = string(match[1])
		// we should return all important header fields
	} else {
		r.IsDetected = false
	}

	return r, nil
}
