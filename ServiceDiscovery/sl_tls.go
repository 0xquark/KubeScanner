package main

import (
	"crypto/tls"
	"fmt"
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
	// Create a TLS config with InsecureSkipVerify set
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", hostAddr, port), tlsConfig)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

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

	// Create a TLS config with InsecureSkipVerify set
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

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
