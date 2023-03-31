package main

import (
	"fmt"
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
	// Use HttpDiscovery implementation to send an HTTP request to the session handler
	httpDiscovery := &HttpDiscovery{}
	plResult, err := httpDiscovery.Discover(sessionHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to discover kube-apiserver: %v", err)
	}

	// Check if the HTTP response contains the Kubernetes server header
	if !strings.Contains(plResult.GetProperties()["header"].(string), "Kubernetes") {
		return nil, nil
	}

	// Return a discovery result indicating that a kube-apiserver instance was detected
	return &KubeApiServerDiscoveryResult{
		isDetected: true,
		properties: map[string]interface{}{
			"header": plResult.GetProperties()["header"],
		},
	}, nil
}
