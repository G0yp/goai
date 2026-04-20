package client

// for general api requests
type ChatCompletionRequest struct {
	Model         string          `json:"model"`
	Messages      []Message       `json:"messages"`
	Stream        bool            `json:"stream"`
	StreamOptions map[string]bool `json:"stream_options,omitempty"`
}

// for non-streamed responses
type Message struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
}

// for streaming responses

type ChoiceStream struct {
	Delta Message `json:"delta"`
}

type ChatCompletionStreamResponse struct {
	Choices []ChoiceStream `json:"choices"`
}
