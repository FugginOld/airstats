package main

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ──────────────────────────────────────────────────────────────────────────────
// GetAllSettings
// ──────────────────────────────────────────────────────────────────────────────

func TestGetAllSettings_Success(t *testing.T) {
	db := &mockDB{
		queryQueue: []mockQueryResult{
			{rows: newMockRows(
				func(dest ...any) error {
					*(dest[0].(*int)) = 1
					*(dest[1].(*string)) = "route_table_limit"
					*(dest[2].(*string)) = "10"
					*(dest[3].(*string)) = "Limit for route table"
					return nil
				},
				func(dest ...any) error {
					*(dest[0].(*int)) = 2
					*(dest[1].(*string)) = "record_holder_table_limit"
					*(dest[2].(*string)) = "5"
					*(dest[3].(*string)) = "Limit for record holder table"
					return nil
				},
			)},
		},
	}
	svc := NewSettingsService(newTestPG(db))

	settings, err := svc.GetAllSettings()
	if err != nil {
		t.Fatalf("GetAllSettings() unexpected error: %v", err)
	}
	if len(settings) != 2 {
		t.Fatalf("len(settings) = %d, want 2", len(settings))
	}
	if settings[0].SettingKey != "route_table_limit" || settings[0].SettingValue != "10" {
		t.Errorf("settings[0] = %+v, unexpected values", settings[0])
	}
}

func TestGetAllSettings_QueryError(t *testing.T) {
	db := &mockDB{
		queryQueue: []mockQueryResult{
			{rows: nil, err: errors.New("db unavailable")},
		},
	}
	svc := NewSettingsService(newTestPG(db))

	if _, err := svc.GetAllSettings(); err == nil {
		t.Fatal("GetAllSettings() expected error, got nil")
	}
}

func TestGetAllSettings_ScanError(t *testing.T) {
	db := &mockDB{
		queryQueue: []mockQueryResult{
			{rows: newMockRows(
				func(dest ...any) error { return errors.New("scan error") },
			)},
		},
	}
	svc := NewSettingsService(newTestPG(db))

	if _, err := svc.GetAllSettings(); err == nil {
		t.Fatal("GetAllSettings() expected scan error, got nil")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// GetSetting
// ──────────────────────────────────────────────────────────────────────────────

func TestGetSetting_Found(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{
			&mockRow{scanFn: func(dest ...any) error {
				*(dest[0].(*int)) = 3
				*(dest[1].(*string)) = "route_table_limit"
				*(dest[2].(*string)) = "8"
				*(dest[3].(*string)) = "Limit for route table"
				return nil
			}},
		},
	}
	svc := NewSettingsService(newTestPG(db))

	setting, err := svc.GetSetting("route_table_limit")
	if err != nil {
		t.Fatalf("GetSetting() unexpected error: %v", err)
	}
	if setting.SettingValue != "8" {
		t.Errorf("SettingValue = %q, want %q", setting.SettingValue, "8")
	}
}

func TestGetSetting_NotFound(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
	}
	svc := NewSettingsService(newTestPG(db))

	if _, err := svc.GetSetting("missing_key"); err == nil {
		t.Fatal("GetSetting() expected error for missing key, got nil")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// UpdateSetting
// ──────────────────────────────────────────────────────────────────────────────

func TestUpdateSetting_Success(t *testing.T) {
	db := &mockDB{
		execQueue: []mockExecResult{
			{tag: pgconn.NewCommandTag("UPDATE 1")},
		},
	}
	svc := NewSettingsService(newTestPG(db))

	if err := svc.UpdateSetting("route_table_limit", "15"); err != nil {
		t.Fatalf("UpdateSetting() unexpected error: %v", err)
	}
}

func TestUpdateSetting_ExecError(t *testing.T) {
	db := &mockDB{
		execQueue: []mockExecResult{
			{err: errors.New("exec failed")},
		},
	}
	svc := NewSettingsService(newTestPG(db))

	if err := svc.UpdateSetting("route_table_limit", "15"); err == nil {
		t.Fatal("UpdateSetting() expected error, got nil")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// UpdateSettings
// ──────────────────────────────────────────────────────────────────────────────

func TestUpdateSettings_Success(t *testing.T) {
	db := &mockDB{}
	svc := NewSettingsService(newTestPG(db))

	if err := svc.UpdateSettings(map[string]string{"route_table_limit": "5"}); err != nil {
		t.Fatalf("UpdateSettings() unexpected error: %v", err)
	}
}

func TestUpdateSettings_BeginError(t *testing.T) {
	db := &mockDB{
		beginFn: func() (pgx.Tx, error) {
			return nil, errors.New("cannot begin transaction")
		},
	}
	svc := NewSettingsService(newTestPG(db))

	if err := svc.UpdateSettings(map[string]string{"k": "v"}); err == nil {
		t.Fatal("UpdateSettings() expected begin error, got nil")
	}
}

func TestUpdateSettings_ExecError(t *testing.T) {
	db := &mockDB{
		beginFn: func() (pgx.Tx, error) {
			return &mockTx{
				execFn: func(_ string, _ ...any) (pgconn.CommandTag, error) {
					return pgconn.CommandTag{}, errors.New("exec failed in tx")
				},
			}, nil
		},
	}
	svc := NewSettingsService(newTestPG(db))

	if err := svc.UpdateSettings(map[string]string{"k": "v"}); err == nil {
		t.Fatal("UpdateSettings() expected exec error, got nil")
	}
}

func TestUpdateSettings_CommitError(t *testing.T) {
	db := &mockDB{
		beginFn: func() (pgx.Tx, error) {
			return &mockTx{commitErr: errors.New("commit failed")}, nil
		},
	}
	svc := NewSettingsService(newTestPG(db))

	if err := svc.UpdateSettings(map[string]string{"k": "v"}); err == nil {
		t.Fatal("UpdateSettings() expected commit error, got nil")
	}
}
