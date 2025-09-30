package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"strconv"

	"sortedstartup/chatservice/dao"
	"sortedstartup/chatservice/events"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	settings "sortedstartup/chatservice/settings"
)

type SettingService struct {
	dao   dao.SettingsDAO
	queue queue.Queue
}

func NewSettingService(queue queue.Queue, daoFactory dao.DAOFactory) *SettingService {
	slog.Info("settings_service:NewSettingService")
	settingsDAO, err := daoFactory.CreateSettingsDAO()
	if err != nil {
		slog.Error("settings_service:NewSettingService, failed to create settings DAO", "error", err)
		return nil
	}
	return &SettingService{dao: settingsDAO, queue: queue}
}

func (s *SettingService) Init() {
	// since right now the Setting is in chatservice so chatservice handles migrations
	isFirstBoot, err := s.IsFirstBoot()
	if err != nil {
		log.Printf("Failed to check if this is first boot: %v", err)
		return
	}

	if isFirstBoot {
		s.SetSetting(context.Background(), settings.DefaultSettings.ToProto())
	}

	s.FirstBootComplete()
}

func (s *SettingService) FirstBootComplete() {
	err := s.dao.SetSettingValue("is_first_boot", "1")
	if err != nil {
		log.Printf("Failed to set is_first_boot setting: %v", err)
	}
}

func (s *SettingService) GetSetting(ctx context.Context) (*pb.Settings, error) {
	slog.Info("settings_service:GetSetting", "settingService", s)
	settingsString, err := s.dao.GetSettingValue("settings")
	if err != nil {
		slog.Error("settings_service:GetSetting", "error", "failed to get settings", "error", err)
		return nil, fmt.Errorf("failed to get settings")
	}

	//json decode the settings
	var settingsObj settings.Settings
	err = json.Unmarshal([]byte(settingsString), &settingsObj)
	if err != nil {
		slog.Error("settings_service:GetSetting", "error", "failed to unmarshal settings", "error", err)
		return nil, fmt.Errorf("failed to get settings")
	}

	return settingsObj.ToProto(), nil
}

func (s *SettingService) SetSetting(ctx context.Context, settingsProto *pb.Settings) error {
	slog.Info("settings_service:SetSetting", "settingService", s)
	settingsJSON, err := json.Marshal(settings.FromProto(settingsProto))
	if err != nil {
		slog.Error("settings_service:SetSetting", "error", "failed to set settings", "error", err)
		return fmt.Errorf("failed to set settings")
	}

	err = s.dao.SetSettingValue("settings", string(settingsJSON))
	if err != nil {
		slog.Error("settings_service:SetSetting", "error", "failed to set settings", "error", err)
		return fmt.Errorf("failed to set settings")
	}

	log.Printf("Publishing event [%s], data:[%s] to reload settings", events.SETTINGS_CHANGED_EVENT, "")
	slog.Info("settings_service:SetSetting", "event", events.SETTINGS_CHANGED_EVENT, "data", "")
	// publish an event, any subscriber now need to reload settings from the database
	s.queue.Publish(context.Background(), events.SETTINGS_CHANGED_EVENT, []byte(""))

	return nil
}

// IsFirstBoot checks if this is the first boot by looking for the 'is_first_boot' setting
// Returns true if the setting doesn't exist or is 0, false otherwise
// Returns an error if there's a database error (except for sql.ErrNoRows)
func (s *SettingService) IsFirstBoot() (bool, error) {
	value, err := s.dao.GetSettingValue("is_first_boot")
	if err != nil {
		// If the setting doesn't exist, consider it first boot
		if err == sql.ErrNoRows {
			return true, nil
		}
		// For other database errors, return the error
		return false, fmt.Errorf("error getting is_first_boot setting: %w", err)
	}

	// Try to parse the value as an integer
	intValue, err := strconv.Atoi(value)
	if err != nil {
		// If we can't parse it, consider it first boot
		log.Printf("Error parsing is_first_boot value '%s': %v", value, err)
		return true, nil
	}

	// Return true if value is 0, false otherwise
	return intValue == 0, nil
}
