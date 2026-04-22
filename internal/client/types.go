package client

// for general api requests
type ChatCompletionRequest struct {
	Model         string         `json:"model"`
	Messages      []Message      `json:"messages"`
	Stream        bool           `json:"stream"`
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ClientHistory struct {
	Messages    []Message
	TotalTokens int
}

// for non-streamed responses
type Message struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
	Tokens  int    `json:"-"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// for streaming responses

type ChoiceStream struct {
	Delta Message `json:"delta"`
}

type ChatCompletionStreamResponse struct {
	Choices []ChoiceStream `json:"choices"`
}
