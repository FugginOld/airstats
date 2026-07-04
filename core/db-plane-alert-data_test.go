package main

import (
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestCheckForUpdates_TableMissing(t *testing.T) {
	db := &mockDB{queryRowQueue: []pgx.Row{
		&mockRow{scanFn: func(dest ...any) error {
			*(dest[0].(*bool)) = false
			return nil
		}},
	}}

	if _, _, err := checkForUpdates(newTestPG(db), false); err == nil {
		t.Fatal("expected error when interesting_aircraft table doesn't exist")
	}
}

func TestCheckForUpdates_EmptyTableNeedsUpdate(t *testing.T) {
	db := &mockDB{queryRowQueue: []pgx.Row{
		&mockRow{scanFn: func(dest ...any) error { *(dest[0].(*bool)) = true; return nil }}, // table exists
		intRow(0), // row count == 0
	}}

	needsUpdating, _, err := checkForUpdates(newTestPG(db), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !needsUpdating {
		t.Error("needsUpdating = false, want true for an empty table")
	}
}

func TestCheckForUpdates_CustomURLNonEmptySkips(t *testing.T) {
	db := &mockDB{queryRowQueue: []pgx.Row{
		&mockRow{scanFn: func(dest ...any) error { *(dest[0].(*bool)) = true; return nil }}, // table exists
		intRow(5), // row count > 0
	}}

	// isCustom=true with existing data must skip without ever reaching the
	// network-calling commit-hash comparison branch.
	needsUpdating, _, err := checkForUpdates(newTestPG(db), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if needsUpdating {
		t.Error("needsUpdating = true, want false for a custom URL with existing data")
	}
}
