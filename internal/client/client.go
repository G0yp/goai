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
	History    []Message
	HTTPClient *http.Client
}

func NewClient(baseURL string, model string, systemPrompt string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Model:      model,
		History:    []Message{{Role: "system", Content: systemPrompt}},
		HTTPClient: &http.Client{},
	}
}

func (c *Client) SendChatRequest(prompt string) (string, error) {
	const maxMessages = 20
	if len(c.History) >= maxMessages {
		c.History = append(c.History[:1], c.History[2:]...)
	}

	reqBody := ChatCompletionRequest{
		Model:    c.Model,
		Messages: append(c.History, Message{Role: "user", Content: prompt}),
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

	c.History = append(c.History, Message{Role: "user", Content: prompt})

	response := completionResp.Choices[0].Message.Content

	c.History = append(c.History, Message{Role: "assistant", Content: response})

	return response, nil
}

func (c *Client) SendChatRequestStream(prompt string, out io.Writer) error {
	const maxMessages = 20
	if len(c.History) >= maxMessages {
		c.History = append(c.History[:1], c.History[2:]...)
	}

	reqBody := ChatCompletionRequest{
		Model:    c.Model,
		Messages: append(c.History, Message{Role: "user", Content: prompt}),
		Stream:   true,
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

	c.History = append(c.History, Message{Role: "user", Content: prompt})
	if role == "" {
		role = "assistant"
	}
	c.History = append(c.History, Message{Role: role, Content: fullContent.String()})

	return nil
}
