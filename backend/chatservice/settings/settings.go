package config

import (
	"sortedstartup/chatservice/proto"
	"sync"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

/*
- TODO: To think : settings have to be app level and then broken down to the service level
*/
type Settings struct {
	OpenAIAPIKey string `koanf:"openai_api_key" json:"openai_api_key"`
	OpenAIAPIURL string `koanf:"openai_api_url" json:"openai_api_url"`
	OllamaURL    string `koanf:"ollama_url" json:"ollama_url"`
}

type SettingsManager struct {
	settings *Settings
	mu       sync.RWMutex
	parser   koanf.Parser
}

func NewSettingsManager() *SettingsManager {
	cm := &SettingsManager{
		parser: json.Parser(),
	}
	return cm
}

func (cm *SettingsManager) LoadSettingsFromProto(settings *proto.Settings) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.settings = &Settings{
		OpenAIAPIKey: settings.OPENAI_API_KEY,
		OpenAIAPIURL: settings.OPENAI_API_URL,
		OllamaURL:    settings.OLLAMA_URL,
	}

	cm.LoadSettings(cm.settings)
	return nil
}

func (cm *SettingsManager) LoadSettings(settings *Settings) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Create new koanf instance
	k := koanf.New(".")

	if err := k.Load(structs.Provider(settings, "koanf"), nil); err != nil {
		return err
	}

	// Create new config struct
	var newSettings Settings

	// Unmarshal into the struct
	if err := k.Unmarshal("", &newSettings); err != nil {
		return err
	}

	// Replace the config atomically
	cm.settings = &newSettings
	return nil
}

func (cm *SettingsManager) GetSettings() *Settings {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.settings
}
