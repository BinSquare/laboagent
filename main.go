package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type string
	Data string
}

const (
	SystemPrompt = `
	Provide an ordered array of immediate steps to take control the mouse or keyboard of the computer to complete a goal in json format only.
	Do not include any additional information.
	Include specific instructions like clicking, typing, scrolling, and waiting for the page to load.
	The format of the response should be in JSON, ordered with appropriate wait time for the action to complete.
	The JSON should include one of the following actions: "mouse_move", "mouse_click", "keyboard_type", "keyboard_shortcut", or "scroll".
	Example JSON format:
	[{
		"action": "mouse_move",
		"x": 500,
		"y": 300
	}]
	or 
	[{
		"action": "mouse_move",
		"x": 500,
		"y": 300
	},
	{
		"action": "mouse_click",
	},
	{
		"action": "keyboard_type",
		"text": "Hello world"
	}]
	`
)

var userPrompt string

// Function to get user input from the command line
func getUserInput() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nWhat would you like to do next? Please enter your input: ")
	userInput, _ := reader.ReadString('\n') // Reads until newline character
	return userInput
}

// AI Response structure for actions like mouse, keyboard control// AIResponse structure for actions like mouse, keyboard control
type AIResponse struct {
	Action string   `json:"action"`
	X      int      `json:"x,omitempty"`    // For mouse move
	Y      int      `json:"y,omitempty"`    // For mouse move
	Text   string   `json:"text,omitempty"` // For typing
	Keys   []string `json:"keys,omitempty"` // For keyboard shortcuts
	Time   int      `json:"time,omitempty"` // For wait time in ms
}

// Function to parse and execute mouse/keyboard actions using robotgo
func handleAIResponse(aiResponse string) error {
	// Extract the full JSON array (or object) part from the AI response
	re := regexp.MustCompile(`(?s)\[.*\]`) // This matches anything between square brackets ([])
	jsonStr := re.FindString(aiResponse)

	fmt.Printf("Response: The JSON string to take action on: %s\n", jsonStr)

	if jsonStr == "" {
		return fmt.Errorf("no valid JSON found in AI response")
	}

	// Try to parse the AI response as an array of AIResponse objects
	var responses []AIResponse
	err := json.Unmarshal([]byte(jsonStr), &responses)
	if err != nil {
		return fmt.Errorf("failed to parse AI response: %v", err)
	}

	// Iterate through each action in the JSON array and handle it
	for _, response := range responses {
		err := handleSingleAction(response)
		if err != nil {
			return err
		}
	}

	return nil
}

// Function to handle a single AIResponse action
func handleSingleAction(response AIResponse) error {
	switch response.Action {
	case "mouse_move":
		fmt.Printf("Moving mouse to (%d, %d)\n", response.X, response.Y)
		robotgo.Move(response.X, response.Y) // Use robotgo to move the mouse

	case "mouse_click":
		fmt.Println("Clicking the mouse")
		robotgo.Click() // Use robotgo to click the mouse

	case "keyboard_type":
		fmt.Printf("Typing text: %s\n", response.Text)
		robotgo.TypeStr(response.Text) // Use robotgo to type text

	case "keyboard_shortcut":
		fmt.Printf("Executing keyboard shortcut: %s\n", strings.Join(response.Keys, "+"))
		// Simulate keyboard shortcuts
		for _, key := range response.Keys {
			robotgo.KeyTap(key)
		}

	case "wait":
		fmt.Printf("Waiting for %d milliseconds\n", response.Time)
		time.Sleep(time.Duration(response.Time) * time.Millisecond) // Pause for the specified time

	default:
		fmt.Println("Unknown action from AI response")
	}
	return nil
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

				err, base64Image := imgToBase64(img)
				if err != nil {
					fmt.Println("Error saving image:", err)
					continue
				}

				prompt := fmt.Sprintf("%s %s", SystemPrompt, userPrompt)

				fmt.Printf("Input: %s", base64Image)
				// After capturing the current image, send it to the LLM to be identified.
				llmResponse := promptWithImage(ctx, prompt, "")

				fmt.Printf("Attempting to execute the commands based on llm response..")
				err = handleAIResponse(llmResponse)
				if err != nil {
					fmt.Printf("Unhandled error %s", err)
					continue
				}

				fmt.Printf("Response: Action taken by ai model %s", llmResponse)

				// Prompt the user for the next action after the LLM response
				userPrompt = getUserInput()

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

	// Capture the initial state.
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

	mailbox := make(chan Message)

	// Goroutine to handle any messages to the mailbox
	go agentCtrl(ctx, mailbox, c)

	// Main loop to interact with user
	for {
		// Get the user's input
		userPrompt = getUserInput()


		time.Sleep(time.Second * 5)
		// Create prompt with user's input and system context
		prompt := fmt.Sprintf("%s %s", SystemPrompt, userPrompt)

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

		err, base64Image := imgToBase64(img)
		if err != nil {
			fmt.Println("Error saving image:", err)
			continue
		}

		// Send prompt to the AI model for action interpretation
		res := promptWithImage(ctx, prompt, base64Image)
		if err != nil {
			fmt.Println("Error sending prompt to AI model:", err)
			continue
		}

		err = handleAIResponse(res)
		if err != nil {
			fmt.Printf("Unhandled error %s", err)
			continue
		}

		fmt.Printf("Response: Action taken by ai model %s", res)

		// Based on AI model response, determine the next action
		switch res {
		case "capture_desktop":
			mailbox <- Message{Type: "capture_desktop", Data: ""}
		case "send_message":
			mailbox <- Message{Type: "send_message", Data: "User-requested message"}
		default:
			mailbox <- Message{Type: "unknown_command", Data: ""}
		}
	}
}
