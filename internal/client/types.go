package client

import "net/http"

// for general api requests
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// for non streamed responses
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
}

// For streaming responses

type ChoiceStream struct {
	Message Message `json:"delta"`
}

type ChatCompletionStreamResponse struct {
	Choices []ChoiceStream `json:"choices"`
}

type Client struct {
	BaseURL    string
	Model      string
	History    []Message
	HTTPClient *http.Client
}
