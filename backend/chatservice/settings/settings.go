package config

import (
	"context"
	"fmt"
	"log"
	"sortedstartup/chatservice/dao"
	"sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	"sync"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"

	events "sortedstartup/chatservice/events"
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
	queue    queue.Queue
	dao      dao.SettingsDAO
}

const SQLITE_DB_URL = "db.sqlite"

func NewSettingsManager(queue queue.Queue) *SettingsManager {
	dao := dao.NewSQLiteSettingsDAO(SQLITE_DB_URL)

	cm := &SettingsManager{
		settings: &Settings{},
		parser:   json.Parser(),
		queue:    queue,
		dao:      dao,
	}

	cm.StartSettingsChangedSubscriber()
	return cm
}

func (cm *SettingsManager) LoadSettingsFromProto(settings *proto.Settings) error {

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

func (s *SettingsManager) StartSettingsChangedSubscriber() {
	go func() {
		sub, err := s.queue.Subscribe(context.Background(), events.SETTINGS_CHANGED_EVENT)
		if err != nil {
			fmt.Printf("Failed %v\n", err)
			return
		}
		for msg := range sub {
			log.Printf("Received message [%s]: %s\n", events.SETTINGS_CHANGED_EVENT, string(msg.Data))
			// reload settings from the database
			log.Println("Reloading settings from the database")
			s.LoadSettingsFromDB()

		}
	}()
}

func (s *SettingsManager) LoadSettingsFromDB() error {

	settings, err := s.dao.GetSettings()
	if err != nil {
		return err
	}

	return s.LoadSettingsFromProto(settings)
}
