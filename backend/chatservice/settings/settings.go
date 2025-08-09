package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sortedstartup/chatservice/dao"
	"sortedstartup/chatservice/events"
	"sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	"sync"
)

var SQLITE_DB_URL = "db.sqlite"

/*
- TODO: To think : settings have to be app level and then broken down to the service level
*/
type Settings struct {
	OpenAIAPIKey string `koanf:"openai_api_key" json:"openai_api_key"`
	OpenAIAPIURL string `koanf:"openai_api_url" json:"openai_api_url"`
	OllamaURL    string `koanf:"ollama_url" json:"ollama_url"`
}

var DefaultSettings = &Settings{
	OpenAIAPIKey: "",
	OpenAIAPIURL: "https://api.openai.com/v1/chat/completions",
	OllamaURL:    "",
}

func (s *Settings) ToProto() *proto.Settings {
	return &proto.Settings{
		OPENAI_API_KEY: s.OpenAIAPIKey,
		OPENAI_API_URL: s.OpenAIAPIURL,
		OLLAMA_URL:     s.OllamaURL,
	}
}

func FromProto(protoSettings *proto.Settings) *Settings {
	return &Settings{
		OpenAIAPIKey: protoSettings.OPENAI_API_KEY,
		OpenAIAPIURL: protoSettings.OPENAI_API_URL,
		OllamaURL:    protoSettings.OLLAMA_URL,
	}
}

// Application should use settings from here, not directly from the database
// This monitors the database for changes and reloads the settings
type SettingsManager struct {
	settings *Settings
	mu       sync.RWMutex
	queue    queue.Queue
	dao      dao.SettingsDAO
}

func NewSettingsManager(queue queue.Queue) *SettingsManager {
	dao := dao.NewSQLiteSettingsDAO(SQLITE_DB_URL)

	cm := &SettingsManager{
		settings: &Settings{},
		queue:    queue,
		dao:      dao,
	}

	cm.StartSettingsChangedSubscriber()
	return cm
}

func (cm *SettingsManager) LoadSettingsFromProto(protoSettings *proto.Settings) error {

	cm.settings = &Settings{
		OpenAIAPIKey: protoSettings.OPENAI_API_KEY,
		OpenAIAPIURL: protoSettings.OPENAI_API_URL,
		OllamaURL:    protoSettings.OLLAMA_URL,
	}

	cm.LoadSettings(cm.settings)
	return nil
}

func (cm *SettingsManager) LoadSettings(settings_ *Settings) error {

	// The lock prevents race conditions when loading settings from the database
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Create new config struct
	// clone the settings_, Later this will be replaced by koanf
	newSettings := *settings_

	// Replace the config atomically
	cm.settings = &newSettings
	return nil
}

func (cm *SettingsManager) GetSettings() *Settings {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.settings
}

func (s *SettingsManager) StartSettingsChangedSubscriber() {
	go func() {
		sub, err := s.queue.Subscribe(context.Background(), events.SETTINGS_CHANGED_EVENT)
		if err != nil {
			fmt.Printf("Failed %v\n", err)
			return
		}
		for msg := range sub {
			log.Printf("Received message [%s], data:[%s]\n", events.SETTINGS_CHANGED_EVENT, string(msg.Data))
			// reload settings from the database
			log.Println("Reloading settings from the database")
			s.LoadSettingsFromDB()

		}
	}()
}

func (s *SettingsManager) LoadSettingsFromDB() error {

	settingsString, err := s.dao.GetSettingValue("settings")
	if err != nil {
		return err
	}

	//json decode the settings
	var settings Settings
	err = json.Unmarshal([]byte(settingsString), &settings)
	if err != nil {
		return err
	}

	return s.LoadSettings(&settings)
}
