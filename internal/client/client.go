package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func NewClient(baseURL string, model string, systemPrompt string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Model:      model,
		History:    []Message{{Role: "system", Content: systemPrompt}},
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
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
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
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
