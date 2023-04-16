package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
)

type TlsSessionDiscovery struct {
}

type TlsSessionDiscoveryResult struct {
	isTls bool
	host  string
	port  int
}

type TlsSessionHandler struct {
	host string
	port int
	conn *tls.Conn
}

func (d *TlsSessionDiscovery) Protocol() TransportProtocol {
	return TCP
}

func (d *TlsSessionDiscovery) SessionLayerDiscover(hostAddr string, port int) (iSessionLayerDiscoveryResult, error) {
	// Parse the proxy address
	proxyUrl, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		return nil, err
	}

	// Create a TLS config with InsecureSkipVerify set and the proxy address
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            nil,
	}
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyUrl),
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Get(fmt.Sprintf("https://%s:%d", hostAddr, port))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &TlsSessionDiscoveryResult{isTls: true, host: hostAddr, port: port}, nil
}

func (d *TlsSessionDiscoveryResult) Protocol() SessionLayerProtocol {
	return TLS
}

func (d *TlsSessionDiscoveryResult) GetIsDetected() bool {
	return d.isTls
}

func (d *TlsSessionDiscoveryResult) GetProperties() map[string]interface{} {
	return nil
}

func (d *TlsSessionDiscoveryResult) GetSessionHandler() (iSessionHandler, error) {
	return &TlsSessionHandler{host: d.host, port: d.port}, nil
}

func (d *TlsSessionHandler) Connect() error {
	// Parse the proxy address
	proxyUrl, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		return err
	}

	// Create a TLS config with InsecureSkipVerify set and the proxy address
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            nil,
	}
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyUrl),
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Get(fmt.Sprintf("https://%s:%d", d.host, d.port))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", d.host, d.port), tlsConfig)
	if err != nil {
		return err
	}
	d.conn = conn
	return nil
}

func (d *TlsSessionHandler) Destory() error {
	return d.conn.Close()
}

func (d *TlsSessionHandler) Write(data []byte) (int, error) {
	return d.conn.Write(data)
}

func (d *TlsSessionHandler) Read(data []byte) (int, error) {
	return d.conn.Read(data)
}

func (d *TlsSessionHandler) GetHost() string {
	return d.host
}

func (d *TlsSessionHandler) GetPort() int {
	return d.port
}
