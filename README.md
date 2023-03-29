# KubeScanner

## Port Scanning Api
Input: IP or IP range + port or port range

Output: which ports are open on which IPs

`$ ./PortDiscovery ipaddr ports`

<img width="406" alt="image" src="https://user-images.githubusercontent.com/84588720/227048473-19e6971b-34b6-4d1b-8209-aa4b1943f4c2.png">



## Service Discovery API

### General concept

The service discovery's goal is to map a given host address and port to the following resolution:
* Session layer protocol: TLS, SSH or none
* Presentation layer protocol: HTTP, gRPC or else
* Application layer protocol: MySQL, ElasticSearch, K8s API server, etc.

A given host and port can be identified as "TLS, HTTP, Kubelet", or "TCP, MySQL" as an example.

Since there are a lot of protocols which are dependent on the underlying session layer, the discovery API contains abstractions (interfaces) so there is no need for example to write different code that discovers "Kubernetes API server" in the case of HTTP or HTTPS.

### Session layer protocols

See interface definitions in [types.go](types.go) of:
* `SessionLayerProtocolDiscovery` - this interface is implemented per protocol (TLS, SSH)
* `iSessionLayerDiscoveryResult` - this is the corresponding result object interface
* `iSessionHandler` - session handler interface, it must have an implementation per protocol to enable presentation layer/application layer to work whit this layer

Example implementation in [sl_tls.go](sl_tls.go) which shows how it is implemented for TLS.

### Transport layer protocols

See interface definitions in [types.go](types.go) of:
* `TransportLayerProtocolDiscovery` - this interface is implemented per protocol (HTTP, gRPC)
* `iTransportLayerDiscoveryResult` - this is the corresponding result object interface

Example implementation for HTTP discovery is in [pl_http_discovery.go](pl_http_discovery.go)

## CLI

Input: IP + Port

Output: Service type

***Also checks for anonymous access for etcd server***

`$ ./ServiceDiscovery ipaddr Port`

<img width="416" alt="image" src="https://user-images.githubusercontent.com/84588720/227048649-7d16413a-8d02-4b0d-92fb-857e53b13a99.png">

#
