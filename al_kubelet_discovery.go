package main

type KubeletDiscoveryResult struct {
}

type KubeletDiscovery struct {
}

func (d *KubeletDiscovery) Protocol() string {
	return "kubelet"
}

func (d *KubeletDiscovery) Discover(sessionHandler iSessionHandler, presenationLayerDiscoveryResult iPresentationDiscoveryResult) (iApplicationDiscoveryResult, error) {
	return nil, nil
}
