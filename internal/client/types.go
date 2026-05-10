package client

// Shared types used across multiple endpoints
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Message struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
	Tokens  int    `json:"-"`
}

// /tokenize
type TokenizeRequest struct {
	Content    string `json:"content"`
	WithPieces bool   `json:"with_pieces"`
	Model      string `json:"model"`
}

type tokenizeResponse struct {
	Tokens []int `json:"tokens"`
}

// /v1/chat/completions
type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type ChatCompletionRequest struct {
	Model         string         `json:"model"`
	Messages      []Message      `json:"messages"`
	Stream        bool           `json:"stream"`
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type ChoiceStream struct {
	Delta Message `json:"delta"`
}

type ChatCompletionStreamResponse struct {
	Choices []ChoiceStream `json:"choices"`
	Usage   Usage          `json:"usage"`
}

// /models
type modelsResponse struct {
	Data []modelEntry `json:"data"`
}

type modelEntry struct {
	Id   string         `json:"id"`
	Meta modelMetaEntry `json:"meta"`
}

type modelMetaEntry struct {
	NCTX int `json:"n_ctx"`
}

// ClientHistory - internal conversation state (not an API type)
type ClientHistory struct {
	Messages    []Message
	TotalTokens int
}
