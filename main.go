package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure all resources are released when main exits

	err, img := captureDesktop()
	if err != nil {
		fmt.Println("LaboAgent exit", err)
		return
	}

	err = imgToPNG(img)
	if err != nil {
		fmt.Println("LaboAgent exit", err)
		return
	}

	err, c := connectWebsocket()
	if err != nil {
		fmt.Println("LaboAgent exit", err)
		return
	}
	defer closeWebsocket(c)

	// Create a ticker for periodic tasks
	ticker := time.NewTicker(3 * time.Second) // Adjust the interval as needed
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Context has been cancelled, exit the loop
			fmt.Println("Context cancelled, exiting...")
			return
		case <-ticker.C:
			// Periodic tasks
			err, img := captureDesktop()
			if err != nil {
				fmt.Println("Error capturing desktop:", err)
				continue
			}

			err = imgToPNG(img)
			if err != nil {
				fmt.Println("Error saving image:", err)
				continue
			}

			// Send a message to the server
			message := "Screenshot taken and saved locally."
			sendMessageWebsocket(message, c)

			// Receive a response from the server
			_, response, err := c.ReadMessage()
			if err != nil {
				fmt.Printf("Read error: %v\n", err)
				// Optionally handle disconnection here
				// For example, you might want to attempt reconnection
				continue
			}
			fmt.Printf("Received from server: %s\n", response)
		}
	}
}
