// Package clients - contain clients and common interfaces for LLM interface
package clients

type (
	ToolFunction struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters"`
	}

	Tool struct {
		Type     string       `json:"name"`
		Function ToolFunction `json:"tool_function"`
	}

	ToolCall struct {
		Function struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
	}
)
