package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

// Start server communications
func connectWebsocket() (error, *websocket.Conn) {
	// Connect to the WebSocket server
	url := "ws://localhost:8080/ws" // Adjust the path if necessary
	fmt.Printf("Connecting to WebSocket server at %s\n", url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Printf("Dial error: %v\n", err)
		return err, nil
	}
	return nil, c
}

func closeWebsocket(c *websocket.Conn) {
	// Send a close message before closing the connection
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		fmt.Printf("Error during WebSocket close handshake: %v\n", err)
		return
	}
	c.Close()
}

func sendMessageWebsocket(message string, c *websocket.Conn) {
	fmt.Printf("Sending message: %s\n", message)
	err := c.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		fmt.Printf("Write error: %v\n", err)
		return
	}
}