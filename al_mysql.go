package main

type MysqlDiscoveryResult struct {
}

type MysqlDiscovery struct {
}

func (d *MysqlDiscovery) Protocol() string {
	return "mysql"
}

func (d *MysqlDiscovery) Discover(sessionHandler iSessionHandler, presenationLayerDiscoveryResult iPresentationDiscoveryResult) (iApplicationDiscoveryResult, error) {
	// Implement mysql server connect

	return nil, nil
}
