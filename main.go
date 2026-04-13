package main

import (
	"github.com/G0yp/goai/cmd/repl"
	"github.com/G0yp/goai/internal/client"
)

func main() {
	apiClient := client.NewClient("http://localhost:8080", "unsloth/Qwen3.5-0.8B-GGUF:Q4_0", "You are a helpful assistant. Use no Emojis")
	repl.Repl(apiClient)
}
