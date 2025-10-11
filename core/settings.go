package main

import (
	"context"
	"fmt"
)

type UserSetting struct {
	ID           int    `json:"id"`
	SettingKey   string `json:"setting_key"`
	SettingValue string `json:"setting_value"`
	Description  string `json:"description"`
}

type SettingsService struct {
	pg *postgres
}

func NewSettingsService(pg *postgres) *SettingsService {
	return &SettingsService{pg: pg}
}

func (s *SettingsService) GetAllSettings() ([]UserSetting, error) {
	query := `
		SELECT id, setting_key, setting_value, description
		FROM user_settings
		ORDER BY setting_key`

	rows, err := s.pg.db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve user settings: %w", err)
	}
	defer rows.Close()

	var settings []UserSetting
	for rows.Next() {
		var setting UserSetting
		err := rows.Scan(&setting.ID, &setting.SettingKey, &setting.SettingValue, &setting.Description)
		if err != nil {
			return nil, fmt.Errorf("Failed to read user settings: %w", err)
		}
		settings = append(settings, setting)
	}

	return settings, nil
}

func (s *SettingsService) GetSetting(key string) (*UserSetting, error) {
	query := `
		SELECT id, setting_key, setting_value, description
		FROM user_settings
		WHERE setting_key = $1`

	var setting UserSetting
	err := s.pg.db.QueryRow(context.Background(), query, key).Scan(
		&setting.ID, &setting.SettingKey, &setting.SettingValue, &setting.Description)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve user setting %s: %w", key, err)
	}

	return &setting, nil
}

func (s *SettingsService) UpdateSetting(key string, value string) error {
	query := `
		UPDATE user_settings
		SET setting_value = $2
		WHERE setting_key = $1`

	result, err := s.pg.db.Exec(context.Background(), query, key, value)
	if err != nil {
		return fmt.Errorf("Failed to update setting %s: %w", key, err)
	}

	if result.RowsAffected() == 1 {
		fmt.Printf("%s updated to %s\n", key, value)
	}

	return nil
}

func (s *SettingsService) UpdateSettings(settings map[string]string) error {

	tx, _ := s.pg.db.Begin(context.Background())
	defer tx.Rollback(context.Background())

	query := `
		UPDATE user_settings
		SET setting_value = $2
		WHERE setting_key = $1`

	for key, value := range settings {
		_, err := tx.Exec(context.Background(), query, key, value)
		if err != nil {
			return fmt.Errorf("Failed to update setting %s: %w", key, err)
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("Failed to commit settings: %w", err)
	}

	return nil
}
