package internal

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type bannoConfig struct {
	OpenAIAPIKey  string `toml:"openai_api_key"`
	DropboxAPIKey string `toml:"dropbox_api_key"`
}

type BannoConfig interface {
	GetOpenAIAPIKey() string
	GetDropboxAPIKey() string
}

func (c *bannoConfig) GetOpenAIAPIKey() string {
	return c.OpenAIAPIKey
}

func (c *bannoConfig) GetDropboxAPIKey() string {
	return c.DropboxAPIKey
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
