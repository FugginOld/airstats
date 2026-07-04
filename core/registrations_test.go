package main

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestUnprocessedRegistrations_Success(t *testing.T) {
	db := &mockDB{
		queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
			*(dest[0].(*int)) = 1
			*(dest[1].(*string)) = "ABCDEF"
			return nil
		})}},
	}

	aircrafts := unprocessedRegistrations(newTestPG(db))
	if len(aircrafts) != 1 || aircrafts[0].Hex != "ABCDEF" {
		t.Errorf("aircrafts = %+v", aircrafts)
	}
}

func TestUnprocessedRegistrations_DBError(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{err: errors.New("db down")}}}

	if aircrafts := unprocessedRegistrations(newTestPG(db)); aircrafts != nil {
		t.Errorf("aircrafts = %+v, want nil", aircrafts)
	}
}

func TestCheckRegistrationExists_SplitsExistingAndNew(t *testing.T) {
	db := &mockDB{
		queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
			*(dest[0].(*int)) = 1
			*(dest[1].(*string)) = "ABCDEF" // mode_s of the one existing registration
			return nil
		})}},
	}

	toProcess := []Aircraft{
		{Hex: "ABCDEF"}, // already has a registration row
		{Hex: "123456"}, // does not
	}

	existing, new := checkRegistrationExists(newTestPG(db), toProcess)

	if len(existing) != 1 || existing[0].Hex != "ABCDEF" {
		t.Errorf("existing = %+v, want [{Hex:ABCDEF}]", existing)
	}
	if len(new) != 1 || new[0].Hex != "123456" {
		t.Errorf("new = %+v, want [{Hex:123456}]", new)
	}
}

func TestCheckRegistrationExists_DBError(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{err: errors.New("db down")}}}

	existing, new := checkRegistrationExists(newTestPG(db), []Aircraft{{Hex: "ABCDEF"}})
	if existing != nil || new != nil {
		t.Errorf("existing=%+v new=%+v, want nil, nil", existing, new)
	}
}

func TestInsertRegistrations_LowercasesModeS(t *testing.T) {
	var capturedArgs []any
	db := &mockDB{
		batchFn: func(b *pgx.Batch) pgx.BatchResults {
			for _, q := range b.QueuedQueries {
				capturedArgs = q.Arguments
			}
			return &mockBatchResults{}
		},
	}

	reg := RegistrationInfo{}
	reg.Response.Aircraft.ModeS = "ABCDEF"
	reg.Response.Aircraft.Registration = "G-TEST"

	insertRegistrations(newTestPG(db), []RegistrationInfo{reg})

	if len(capturedArgs) != 11 {
		t.Fatalf("len(args) = %d, want 11", len(capturedArgs))
	}
	if capturedArgs[3] != "abcdef" {
		t.Errorf("mode_s arg = %v, want lowercased \"abcdef\"", capturedArgs[3])
	}
}
