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
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter input (type exit to end process): ")

	for scanner.Scan() {
		text := scanner.Text()
		text = strings.TrimSpace(text)

		if text == "exit" {
			return
		}

		response, err := apiClient.SendChatRequest(text)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Response: %s\n", response)
		}

		fmt.Print("Enter input (type exit to end process): ")
	}
}

func handleSlashCommands(input string) (bool, error) {
	if command, found := strings.CutPrefix(input, "/"); found {
		switch command {
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("/help - Show this help message")
			fmt.Println("/exit - Exit the application")
			return true, nil
		case "exit":
			os.Exit(0)
			return true, nil
		default:
			return true, fmt.Errorf("unknown command: %s", command)
		}
	} else {
		return false, nil
	}
}
