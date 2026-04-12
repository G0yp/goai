package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/G0yp/goai/internal/client"
)

func main() {
	apiClient := client.NewClient("http://localhost:8080", "unsloth/Qwen3.5-0.8B-GGUF:Q4_0", "You are a helpful assistant. Use no Emojis")
	repl(apiClient)
}

func repl(apiClient *client.Client) {
	const prompt = "Enter input: "
	var stream bool = true

	fmt.Println("Type /help for commands.")

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)

	for scanner.Scan() {
		text := scanner.Text()
		text = strings.TrimSpace(text)

		if action, err := handleSlashCommands(text, &stream); action == true && err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Print(prompt)
			continue
		} else if action == true {
			fmt.Print(prompt)
			continue
		}

		if stream == true {
			fmt.Print("Response: ")
			err := apiClient.SendChatRequestStream(text, os.Stdout)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			fmt.Println()

		} else {
			response, err := apiClient.SendChatRequest(text)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Response: %s\n", response)
			}
		}

		fmt.Print(prompt)
	}
}

func handleSlashCommands(input string, stream *bool) (bool, error) {
	if command, found := strings.CutPrefix(input, "/"); found {
		switch command {
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("/help - Show this help message")
			fmt.Println("/exit - Exit the application")
			fmt.Println("/stream - Toggle streaming mode")
			return true, nil
		case "exit":
			os.Exit(0)
			return true, nil
		case "stream":
			*stream = !*stream
			fmt.Printf("Streaming mode: %v\n", *stream)
			return true, nil

		default:
			return true, fmt.Errorf("unknown command: %s", command)
		}
	} else {
		return false, nil
	}
}
