package clients

import (
	openai "github.com/openai/openai-go/v3" // imported as openai
)

type OpenAIClient struct {
	Key   string
	Model openai.Model
}
