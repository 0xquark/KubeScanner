package main

import (
	"fmt"
	"io"
)

func main() {
	host := "localhost"
	port := 5432

	// Discover session layer protocols
	for _, sessionDiscoveryItem := range SessionDiscoveryList {
		if sessionDiscoveryItem.Reqirement == string(TCP) {
			sessionDiscoveryResult, err := sessionDiscoveryItem.Discovery.SessionLayerDiscover(host, port)
			if err != nil {
				if err != io.EOF {
					fmt.Println("Error while discovering session layer protocol:", err)
				}
				continue
			}

			if sessionDiscoveryResult.GetIsDetected() {
				fmt.Println("Session layer protocol detected:", sessionDiscoveryResult.Protocol())

				// Connect to session handler
				sessionHandler, err := sessionDiscoveryResult.GetSessionHandler()
				if err != nil {
					if err != io.EOF {
						fmt.Println("Error while discovering session layer protocol:", err)
					}
					continue
				}

				// Discover presentation layer protocols
				presentationLayerDetected := false
				for _, presentationDiscoveryItem := range PresentationDiscoveryList {
					if presentationDiscoveryItem.Reqirement == string(TCP) {
						presentationDiscoveryResult, err := presentationDiscoveryItem.Discovery.Discover(sessionHandler)
						if err != nil {
							if err != io.EOF {
								fmt.Println("Error while discovering session layer protocol:", err)
							}
							continue
						}

						if presentationDiscoveryResult.GetIsDetected() {
							presentationLayerDetected = true
							fmt.Println("Presentation layer protocol detected:", presentationDiscoveryResult.Protocol())
							fmt.Println("Properties:", presentationDiscoveryResult.GetProperties())

							// Discover application layer protocols
							for _, applicationDiscoveryItem := range ApplicationDiscoveryList {
								if applicationDiscoveryItem.Reqirement == string(TCP) {
									applicationDiscoveryResult, err := applicationDiscoveryItem.Discovery.Discover(sessionHandler, presentationDiscoveryResult)
									if err != nil {
										if err != io.EOF {
											fmt.Println("Error while discovering session layer protocol:", err)
										}
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

							break // Stop checking presentation layer protocols
						}
					}
				}

				if !presentationLayerDetected {
					fmt.Println("No presentation layer protocol detected")

					// Continue to discover application layer protocols
					for _, applicationDiscoveryItem := range ApplicationDiscoveryList {
						if applicationDiscoveryItem.Reqirement == string(TCP) {
							applicationDiscoveryResult, err := applicationDiscoveryItem.Discovery.Discover(sessionHandler, nil)
							if err != nil {
								if err != io.EOF {
									fmt.Println("Error while discovering session layer protocol:", err)
								}
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
				}

			} else {
				fmt.Println("No session layer protocol detected")
			}

			// If session layer protocol not TCP, continue to the next session layer protocol
			continue
		}

		// If session layer protocol not TCP, continue to the next session layer protocol
		continue
	}
}
