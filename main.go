package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string
	Data string
}

func agentCtrl(ctx context.Context, mailbox chan Message, c *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			// Exit when context is done
			fmt.Println("Context cancelled, exiting WebSocket listener...")
			return
		case msg := <-mailbox:
			// Process messages sent via the channel
			switch msg.Type {
			case "capture_desktop":
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
			case "send_message":
				// Send message to WebSocket
				sendMessageWebsocket(msg.Data, c)

				// Optionally read a response
				_, response, err := c.ReadMessage()
				if err != nil {
					fmt.Printf("Read error: %v\n", err)
					// Optionally handle disconnection and reconnection logic here
					continue
				}
				fmt.Printf("Received from server: %s\n", response)
			}
		}
	}

}

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

	mailbox := make(chan Message)

	// Goroutine to handle any messages to the mailbox
	go agentCtrl(ctx, mailbox, c)

	// Main loop
	for {
		select {
		case <-ctx.Done():
			// Context has been cancelled, exit the loop
			fmt.Println("Context cancelled, exiting main loop...")
			return
		case <-ticker.C:
			// Periodic task: capture a screenshot and send a message
			mailbox <- Message{Type: "capture_desktop", Data: ""}
			mailbox <- Message{Type: "send_message", Data: "Screenshot taken and saved locally."}
		}
	}
}
