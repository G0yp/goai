# todo

- user config (toml or json)
- bubbletea ui
- dynamically fetch max context length from the server
- support for /v1/models
- Keep repl for testing, bubble tea for actual use
- extensions?

# ListModels Implementation

model json response example:

```json
{
  "models": [
    {
      "name": "unsloth/gemma-4-E2B-it-GGUF:Q4_K_M",
      "model": "unsloth/gemma-4-E2B-it-GGUF:Q4_K_M",
      "modified_at": "",
      "size": "",
      "digest": "",
      "type": "model",
      "description": "",
      "tags": [""],
      "capabilities": ["completion", "multimodal"],
      "parameters": "",
      "details": {
        "parent_model": "",
        "format": "gguf",
        "family": "",
        "families": [""],
        "parameter_size": "",
        "quantization_level": ""
      }
    }
  ],
  "object": "list",
  "data": [
    {
      "id": "unsloth/gemma-4-E2B-it-GGUF:Q4_K_M",
      "aliases": ["unsloth/gemma-4-E2B-it-GGUF:Q4_K_M"],
      "tags": [],
      "object": "model",
      "created": 1777491665,
      "owned_by": "llamacpp",
      "meta": {
        "vocab_type": 2,
        "n_vocab": 262144,
        "n_ctx_train": 131072,
        "n_embd": 1536,
        "n_params": 4647450147,
        "size": 3090917516
      }
    }
  ]
}
```