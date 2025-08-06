package config

import (
	"log"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

/*
- TODO: To think : settings have to be app level and then broken down to the service level
*/
type Config struct {
	OpenAIAPIKey  string `koanf:"openai_api_key" json:"openai_api_key"`
	OpenAIBaseURL string `koanf:"openai_base_url" json:"openai_base_url"`
	OllamaURL     string `koanf:"ollama_url" json:"ollama_url"`
	DBPath        string `koanf:"db_path" json:"db_path"`
	FileStorePath string `koanf:"filestore_path" json:"filestore_path"`
	Port          int    `koanf:"port" json:"port"`
	Host          string `koanf:"host" json:"host"`
}

// getDefaultConfig returns a Config with default values
func getDefaultConfig() Config {
	return Config{
		OpenAIAPIKey:  "",
		OpenAIBaseURL: "https://api.openai.com/v1/chat/completions",
		OllamaURL:     "http://localhost:11434",
		DBPath:        "chatservice.db",
		FileStorePath: "filestore",
		Port:          8080,
		Host:          "localhost",
	}
}

// NewConfig creates and returns a new Config instance with values loaded from
// environment variables, config files, and defaults (in that order of priority)
func NewConfig() *Config {
	k := koanf.New(".")

	// Load default values first
	defaultConfig := getDefaultConfig()
	if err := k.Load(structs.Provider(defaultConfig, "koanf"), nil); err != nil {
		log.Printf("Error loading default config: %v", err)
	}

	// Load from config file if it exists
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config.yaml"
	}
	if _, err := os.Stat(configFile); err == nil {
		if err := k.Load(file.Provider(configFile), yaml.Parser()); err != nil {
			log.Printf("Error loading config file %s: %v", configFile, err)
		}
	}

	// Load from environment variables (highest priority)
	// Environment variables should be prefixed with CHATSERVICE_
	if err := k.Load(env.Provider("CHATSERVICE_", ".", func(s string) string {
		// Convert CHATSERVICE_OPENAI_API_KEY to openai_api_key
		return strings.ToLower(strings.Replace(s, "CHATSERVICE_", "", 1))
	}), nil); err != nil {
		log.Printf("Error loading environment variables: %v", err)
	}

	// Unmarshal into Config struct
	var config Config
	if err := k.Unmarshal("", &config); err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}

	return &config
}
