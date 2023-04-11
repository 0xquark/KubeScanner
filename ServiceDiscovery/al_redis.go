package main

import (
	"errors"
	"strings"
)

type RedisDiscovery struct {
}

type RedisDiscoveryResult struct {
	isDetected bool
	properties map[string]interface{}
}

func (d *RedisDiscoveryResult) Protocol() string {
	return "redis"
}

func (d *RedisDiscovery) Protocol() string {
	return "redis"
}

func (r *RedisDiscoveryResult) GetIsAuthRequired() bool {
	return false
}

func (r *RedisDiscoveryResult) GetIsDetected() bool {
	return r.isDetected
}

func (r *RedisDiscoveryResult) GetProperties() map[string]interface{} {
	return r.properties
}

func (d *RedisDiscovery) Discover(sessionHandler iSessionHandler, presentationLayerDiscoveryResult iPresentationDiscoveryResult) (iApplicationDiscoveryResult, error) {
	// Connect to the Redis server
	err := sessionHandler.Connect()
	if err != nil {
		return nil, err
	}
	defer sessionHandler.Destory()

	// Send the INFO command to Redis
	_, err = sessionHandler.Write([]byte("*1\r\n$4\r\nINFO\r\n"))
	if err != nil {
		return nil, err
	}

	// Read the response from Redis
	headerBuf := make([]byte, 1024)
	_, err = sessionHandler.Read(headerBuf)
	if err != nil {
		return nil, err
	}
	line := string(headerBuf)

	// Check if the response is valid
	if !strings.HasPrefix(line, "+") {
		return nil, errors.New("invalid response from Redis")
	}

	// Parse the response and extract the Redis version
	info := make(map[string]string)
	for {
		lineBuf := make([]byte, 1024)
		n, err := sessionHandler.Read(lineBuf)
		if err != nil {
			return nil, err
		}
		line = string(lineBuf[:n])
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		info[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	version, ok := info["redis_version"]
	if !ok {
		return nil, errors.New("Redis version not found in INFO response")
	}

	// Return the discovery result
	return &RedisDiscoveryResult{
		isDetected: true,
		properties: map[string]interface{}{
			"version": version,
		},
	}, nil
}
