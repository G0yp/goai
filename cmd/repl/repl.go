package repl

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/G0yp/goai/internal/client"
)

func Repl(apiClient *client.Client) {
	const prompt = "Enter input: "
	var stream bool = true

	fmt.Println("Type /help for commands.")

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)

	for scanner.Scan() {
		text := scanner.Text()
		text = strings.TrimSpace(text)

		if action, err := handleSlashCommands(text, &stream, apiClient); err != nil {
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

func handleSlashCommands(input string, stream *bool, c *client.Client) (bool, error) {
	if command, found := strings.CutPrefix(input, "/"); found {
		command := strings.Fields(command)
		switch command[0] {
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("/help - Show this help message")
			fmt.Println("/exit - Exit the application")
			fmt.Println("/stream - Toggle streaming mode")
			fmt.Println("/tokens - Shows current token usage")
			fmt.Println("/system - Sets a new system prompt")

			return true, nil
		case "exit":
			os.Exit(0)
			return true, nil
		case "stream":
			*stream = !*stream
			fmt.Printf("Streaming mode: %v\n", *stream)
			return true, nil
		case "tokens":
			fmt.Printf("Current token usage: %v\n", c.History.TotalTokens)
			return true, nil
		case "system":
			c.History.Messages[0].Content = strings.Join(command[1:], " ")
			fmt.Println("Set new system prompt")
			return true, nil

		default:
			return true, fmt.Errorf("unknown command: %s", command)
		}
	} else {
		return false, nil
	}
}
