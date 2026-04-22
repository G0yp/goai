package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	Model      string
	History    ClientHistory
	HTTPClient *http.Client
}

func NewClient(baseURL string, model string, systemPrompt string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Model:      model,
		History:    ClientHistory{Messages: []Message{{Role: "system", Content: systemPrompt}}, TotalTokens: 0},
		HTTPClient: &http.Client{},
	}
}

func (c *Client) TrimHistory() {
	const maxTokens = 4000
	for c.History.TotalTokens > maxTokens && len(c.History.Messages) > 1 {
		c.History.TotalTokens -= c.History.Messages[1].Tokens
		c.History.Messages = append(c.History.Messages[:1], c.History.Messages[2:]...)
	}
}

func (c *Client) SendChatRequest(prompt string) (string, error) {
	c.TrimHistory()

	reqBody := ChatCompletionRequest{
		Model:    c.Model,
		Messages: append(c.History.Messages, Message{Role: "user", Content: prompt}),
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	completionResp := ChatCompletionResponse{}
	if err := json.Unmarshal(body, &completionResp); err != nil {
		return "", err
	}

	if len(completionResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	tokens := completionResp.Usage
	response := completionResp.Choices[0].Message.Content

	promptMessage := Message{Role: "user", Content: prompt, Tokens: tokens.PromptTokens}
	responseMessage := Message{Role: "assistant", Content: response, Tokens: tokens.CompletionTokens}

	c.History.Messages = append(c.History.Messages, promptMessage)
	c.History.Messages = append(c.History.Messages, responseMessage)
	c.History.TotalTokens += tokens.TotalTokens

	return response, nil
}

func (c *Client) SendChatRequestStream(prompt string, out io.Writer) error {
	c.TrimHistory()

	reqBody := ChatCompletionRequest{
		Model:         c.Model,
		Messages:      append(c.History.Messages, Message{Role: "user", Content: prompt}),
		Stream:        true,
		StreamOptions: &StreamOptions{IncludeUsage: true},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// Creats a context with a timeout to prevent hanging if the server doesn't respond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	idleTimeout := 30 * time.Second
	idleTimer := time.AfterFunc(idleTimeout, func() {
		cancel()
	})

	defer idleTimer.Stop()

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// builder lets me create a string without concationation during the stream
	var fullContent strings.Builder
	var role string

	const kilobyte = 1024
	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*kilobyte) // 64 kilobytes
	scanner.Buffer(buf, 1024*kilobyte)  // 1 megabyte
	for scanner.Scan() {
		idleTimer.Reset(idleTimeout)
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			break
		}

		var chunk ChatCompletionStreamResponse
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return err
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		if chunk.Choices[0].Delta.Role != "" {
			role = chunk.Choices[0].Delta.Role
		}
		if chunk.Choices[0].Delta.Content != "" {
			fullContent.WriteString(chunk.Choices[0].Delta.Content)
			fmt.Fprint(out, chunk.Choices[0].Delta.Content)
		}
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}

	c.History.Messages = append(c.History.Messages, Message{Role: "user", Content: prompt})
	if role == "" {
		role = "assistant"
	}
	c.History.Messages = append(c.History.Messages, Message{Role: role, Content: fullContent.String()})

	return nil
}
