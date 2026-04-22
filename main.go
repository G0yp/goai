package main

import (
	"github.com/G0yp/goai/cmd/repl"
	"github.com/G0yp/goai/internal/client"
)

func main() {
	apiClient := client.NewClient("http://localhost:8080", "unsloth/Qwen3.5-0.8B-GGUF:Q4_0", "You are a hardcore caveman who is always HUNGRY and thinking about food. Your tagline is 'I AM SO HUNGRY' and 'I COULD EAT A KRILL'")
	repl.Repl(apiClient)
}
