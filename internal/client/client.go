package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
		Stream:   false,
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

func (c *Client) SendChatRequestStream(prompt string) error {
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

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
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

	completionResp := ChatCompletionStreamResponse{}
	completionResp.Choices = []ChoiceStream{{}}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
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

		// explicit handling for when role and content come in either the same or seperate chunks
		if chunk.Choices[0].Message.Role != "" && chunk.Choices[0].Message.Content != "" {
			completionResp.Choices[0].Message.Role += chunk.Choices[0].Message.Role
			completionResp.Choices[0].Message.Content += chunk.Choices[0].Message.Content
		} else if chunk.Choices[0].Message.Role != "" {
			completionResp.Choices[0].Message.Role += chunk.Choices[0].Message.Role
		} else {
			completionResp.Choices[0].Message.Content += chunk.Choices[0].Message.Content
		}

		fmt.Print(chunk.Choices[0].Message.Content)
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}

	c.History = append(c.History, Message{Role: "user", Content: prompt})
	if completionResp.Choices[0].Message.Role == "" {
		completionResp.Choices[0].Message.Role = "assistant"
	} else {
		c.History = append(c.History, completionResp.Choices[0].Message)
	}

	return nil
}
