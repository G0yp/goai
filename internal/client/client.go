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
	MaxContext int
	History    ClientHistory
	HTTPClient *http.Client
}

func NewClient(baseURL string, model string, systemPrompt string) (*Client, error) {
	client := &Client{
		BaseURL:    baseURL,
		Model:      model,
		HTTPClient: &http.Client{},
	}

	// handles blank model entry - probably better way to do this instead of the for loop
	if model == "" {
		models, err := client.GetAvailableModels()
		if err != nil {
			return nil, err
		}
		for model, ctx := range models {
			client.Model = model
			client.MaxContext = ctx
			break
		}
	}

	tokens, err := client.tokenize(systemPrompt)
	if err != nil {
		return nil, err
	}
	client.History = ClientHistory{
		Messages:    []Message{{Role: "system", Content: systemPrompt, Tokens: tokens}},
		TotalTokens: tokens,
	}

	return client, nil

}

func (c *Client) trimHistory() {
	const maxTokens = 4000
	for c.History.TotalTokens > maxTokens && len(c.History.Messages) > 1 {
		c.History.TotalTokens -= c.History.Messages[1].Tokens
		c.History.Messages = append(c.History.Messages[:1], c.History.Messages[2:]...)
	}
}

func (c *Client) tokenize(message string) (int, error) {
	reqBody := TokenizeRequest{
		Content:    message,
		WithPieces: false,
		Model:      c.Model,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/tokenize", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, nil
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API error (%d): %s", resp.StatusCode, body)
	}

	tokenResp := tokenizeResponse{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return 0, err
	}

	tokens := len(tokenResp.Tokens)

	return tokens, err
}

func (c *Client) UpdateSystemPrompt(systemPrompt string) error {
	tokens, err := c.tokenize(systemPrompt)
	if err != nil {
		return err
	}

	// remove previous system promot tokens from total
	c.History.TotalTokens -= c.History.Messages[0].Tokens

	// update system prompt and token count
	c.History.Messages[0].Content = systemPrompt
	c.History.Messages[0].Tokens = tokens
	c.History.TotalTokens += tokens

	return nil
}

func (c *Client) GetAvailableModels() (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/models", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var modelsResp modelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, err
	}

	models := make(map[string]int, len(modelsResp.Data))
	for _, entry := range modelsResp.Data {
		models[entry.Id] = entry.Meta.NCTX
	}

	return models, nil
}

func (c *Client) SwitchModels(model string, ctx int) error {
	// /props endpoint to fetch the n_ctx variable for max context and return it.
	// will have to update NewClient() to set MaxContext
	// /models/load and /models/unload let me control which models are currently running for the SwitchModel function.
	// HOT OFF THE PRESS, /models RETURN THE n_ctx length, /props reportedly hung itself after the news.

	c.Model = model
	c.MaxContext = ctx
	return nil
}

func (c *Client) ClearHistory() error {
	return nil
}

func (c *Client) SendChatRequest(prompt string) (string, error) {
	c.trimHistory()

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
	c.trimHistory()

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
	var tokens Usage
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

		if len(chunk.Choices) == 0 && chunk.Usage.TotalTokens > 0 {
			tokens = chunk.Usage
			continue
		} else if len(chunk.Choices) == 0 {
			continue
		}

		if chunk.Choices[0].Delta.Content != "" {
			fullContent.WriteString(chunk.Choices[0].Delta.Content)
			fmt.Fprint(out, chunk.Choices[0].Delta.Content)
		}
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}

	promptMessage := Message{Role: "user", Content: prompt, Tokens: tokens.PromptTokens}
	responseMessage := Message{Role: "assistant", Content: fullContent.String(), Tokens: tokens.CompletionTokens}

	c.History.Messages = append(c.History.Messages, promptMessage)
	c.History.Messages = append(c.History.Messages, responseMessage)
	c.History.TotalTokens += tokens.TotalTokens

	return nil
}
