package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type KubeApiServerDiscoveryResult struct {
	isDetected bool
	properties map[string]interface{}
}

func (r *KubeApiServerDiscoveryResult) Protocol() string {
	return "kube-apiserver"
}

func (r *KubeApiServerDiscoveryResult) GetIsDetected() bool {
	return r.isDetected
}

func (r *KubeApiServerDiscoveryResult) GetProperties() map[string]interface{} {
	return r.properties
}

func (r *KubeApiServerDiscoveryResult) GetIsAuthRequired() bool {
	// This is just an example. You would need to implement this method based on your requirements.
	return false
}

type KubeApiServerDiscovery struct {
}

func (d *KubeApiServerDiscovery) Protocol() string {
	return "kube-apiserver"
}

func (d *KubeApiServerDiscovery) Discover(sessionHandler iSessionHandler, presentationLayerDiscoveryResult iPresentationDiscoveryResult) (iApplicationDiscoveryResult, error) {
	// Perform a GET request to the /version endpoint of the HTTP server
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/version", sessionHandler.GetHost(), sessionHandler.GetPort()))
	if err != nil {
		return nil, fmt.Errorf("failed to GET /version: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body and check if it contains the Kubernetes server header
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	if !strings.Contains(resp.Header.Get("Server"), "Kubernetes") {
		return nil, nil
	}

	// Return a discovery result indicating that a kube-apiserver instance was detected
	return &KubeApiServerDiscoveryResult{
		isDetected: true,
		properties: map[string]interface{}{
			"response": string(body),
		},
	}, nil
}
