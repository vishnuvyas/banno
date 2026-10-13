package internal

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type bannoConfig struct {
	OpenAIAPIKey string `toml:"openai_api_key"`
}

type BannoConfig interface {
	GetOpenAIAPIKey() string
}

func (c *bannoConfig) GetOpenAIAPIKey() string {
	return c.OpenAIAPIKey
}

func NewConfig(filename string) (BannoConfig, error) {
	if filename == "" {
		filename = "config.toml"
	}

	var bc bannoConfig
	_, err := toml.DecodeFile(filename, &bc)
	if err != nil {
		return nil, fmt.Errorf("unable to open config file: %s (due to %v)", filename, err)
	}

	return &bc, nil
}
