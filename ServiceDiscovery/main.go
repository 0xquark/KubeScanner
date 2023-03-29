package main

import (
	"fmt"
)

func main() {
	host := "8.8.8.8"
	port := 443

	// Discover session layer protocols
	for _, sessionDiscoveryItem := range SessionDiscoveryList {
		if sessionDiscoveryItem.Reqirement == string(TCP) {
			sessionDiscoveryResult, err := sessionDiscoveryItem.Discovery.SessionLayerDiscover(host, port)
			if err != nil {
				fmt.Println("Error while discovering session layer protocol:", err)
				continue
			}

			if sessionDiscoveryResult.GetIsDetected() {
				fmt.Println("Session layer protocol detected:", sessionDiscoveryResult.Protocol())

				// Connect to session handler
				sessionHandler, err := sessionDiscoveryResult.GetSessionHandler()
				if err != nil {
					fmt.Println("Error while creating session handler:", err)
					continue
				}

				// Discover presentation layer protocols
				for _, presentationDiscoveryItem := range PresentationDiscoveryList {
					if presentationDiscoveryItem.Reqirement == string(TCP) {
						presentationDiscoveryResult, err := presentationDiscoveryItem.Discovery.Discover(sessionHandler)
						if err != nil {
							fmt.Println("Error while discovering presentation layer protocol:", err)
							continue
						}

						if presentationDiscoveryResult.GetIsDetected() {
							fmt.Println("Presentation layer protocol detected:", presentationDiscoveryResult.Protocol())
							fmt.Println("Properties:", presentationDiscoveryResult.GetProperties())

							// Discover application layer protocols
							for _, applicationDiscoveryItem := range ApplicationDiscoveryList {
								if applicationDiscoveryItem.Reqirement == string(TCP) {
									applicationDiscoveryResult, err := applicationDiscoveryItem.Discovery.Discover(sessionHandler, presentationDiscoveryResult)
									if err != nil {
										fmt.Println("Error while discovering application layer protocol:", err)
										continue
									}

									if applicationDiscoveryResult.GetIsDetected() {
										fmt.Println("Application layer protocol detected:", applicationDiscoveryResult.Protocol())
										fmt.Println("Properties:", applicationDiscoveryResult.GetProperties())
									} else {
										fmt.Println("No application layer protocol detected")
									}
								}
							}
						} else {
							fmt.Println("No presentation layer protocol detected")
						}
					}
				}

			} else {
				fmt.Println("No session layer protocol detected")
			}
		}
	}
}
