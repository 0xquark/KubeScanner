package main

type enum TransportProtocol {
	"tcp",
	"udp",
	"http",
	"dns",
}

type enum ApplicationProtocol {
	"mysql",
	"etcd",
	"redis",
}

type Service struct {
	Port int
	Address string
	TransportProtocol string // tcp, udp
	SessionLayerProtocol string // tls, ssh
	PresentationLayerProtocol string // http
	ApplicationLayerProtocol string // mysql, redis, etcd
}

type SessionLayerProtocolDiscovery interface {
	SessionLayerDiscover(ipAddr string, port int, transportProtocol string) string
}

type PresentationLayerProtocolDiscovery interface {
	PresentationLayerDiscover(ipAddr string, port int, sessionLayerProtocol string) string
}

type ApplicationLayerProtocolDiscovery interface {
	ApplicationLayerDiscover(ipAddr string, port int, presentationLayerProtocol string) string
}