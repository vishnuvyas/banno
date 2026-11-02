// Package clients - contain clients and common interfaces for LLM interface
package clients

type (
	ToolFunction struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters"`
	}

	Tool struct {
		Type     string       `json:"type"`
		Function ToolFunction `json:"function"`
	}

	ToolCall struct {
		Function struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			Arguments   map[string]any `json:"arguments"`
		} `json:"function"`
	}

	Message struct {
		Role      string     `json:"role"`
		Content   string     `json:"content"`
		Thinking  string     `json:"thinking"`
		ToolCalls []ToolCall `json:"tool_calls"`
		Images    []string   `json:"images"`
	}

	ChatRequest struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
		Tools    []Tool    `json:"tools"`
		Think    bool      `json:"think"`
	}

	ChatResponse struct{}
)
