package main

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func scanUnprocessedInterestingRow(dest ...any) error {
	*(dest[0].(*int)) = 1                  // id
	*(dest[1].(*string)) = "ABCDEF"        // hex
	*(dest[2].(*string)) = "BA123"         // flight
	*(dest[3].(*string)) = "G-TEST"        // r
	*(dest[4].(*string)) = "B738"          // t
	*(dest[5].(*int)) = 35000              // alt_baro
	*(dest[6].(*int)) = 35100              // alt_geom
	*(dest[7].(*float64)) = 450            // gs
	*(dest[8].(*int)) = 440                // ias
	*(dest[9].(*int)) = 460                // tas
	*(dest[10].(*float64)) = 90            // track
	*(dest[11].(*int)) = 0                 // baro_rate
	*(dest[12].(*float64)) = 51.5          // lat
	*(dest[13].(*float64)) = -0.1          // lon
	*(dest[14].(*int)) = 0                 // alert
	*(dest[15].(*int)) = 0                 // db_flags
	*(dest[16].(*time.Time)) = time.Now()  // first_seen
	*(dest[17].(*float64)) = 123.0         // first_seen_epoch
	return nil
}

func TestUnprocessedInteresting_Success(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{rows: newMockRows(scanUnprocessedInterestingRow)}}}

	aircrafts := unprocessedInteresting(newTestPG(db))
	if len(aircrafts) != 1 || aircrafts[0].Hex != "ABCDEF" {
		t.Errorf("aircrafts = %+v", aircrafts)
	}
}

func TestUnprocessedInteresting_DBError(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{err: errors.New("db down")}}}

	if aircrafts := unprocessedInteresting(newTestPG(db)); aircrafts != nil {
		t.Errorf("aircrafts = %+v, want nil", aircrafts)
	}
}

func TestUpdateInterestingSeen_NoUnprocessed_NoOp(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{rows: emptyRows()}}}
	updateInterestingSeen(newTestPG(db)) // must not panic or attempt further DB calls
}

func TestUpdateInterestingSeen_MergesAircraftFieldsOntoMatch(t *testing.T) {
	var insertedArgs []any

	db := &mockDB{
		queryQueue: []mockQueryResult{
			{rows: newMockRows(scanUnprocessedInterestingRow)}, // unprocessedInteresting
			{rows: newMockRows(func(dest ...any) error { // interesting_aircraft lookup
				*(dest[0].(*string)) = "ABCDEF" // icao — matches hex, uppercased
				*(dest[1].(*sql.NullString)) = sql.NullString{String: "reg", Valid: true}
				*(dest[2].(*sql.NullString)) = sql.NullString{String: "operator", Valid: true}
				*(dest[3].(*sql.NullString)) = sql.NullString{String: "type", Valid: true}
				*(dest[4].(*sql.NullString)) = sql.NullString{String: "icaotype", Valid: true}
				*(dest[5].(*sql.NullString)) = sql.NullString{String: "group", Valid: true}
				*(dest[6].(*sql.NullString)) = sql.NullString{String: "t1", Valid: true}
				*(dest[7].(*sql.NullString)) = sql.NullString{String: "t2", Valid: true}
				*(dest[8].(*sql.NullString)) = sql.NullString{String: "t3", Valid: true}
				*(dest[9].(*sql.NullString)) = sql.NullString{String: "cat", Valid: true}
				return nil
			})},
		},
		batchFn: func(b *pgx.Batch) pgx.BatchResults {
			for _, q := range b.QueuedQueries {
				if len(q.Arguments) == 27 {
					insertedArgs = q.Arguments
				}
			}
			return &mockBatchResults{}
		},
	}

	updateInterestingSeen(newTestPG(db))

	if insertedArgs == nil {
		t.Fatal("expected an interesting_aircraft_seen insert, got none")
	}
	if insertedArgs[10] != "ABCDEF" {
		t.Errorf("hex arg = %v, want ABCDEF (merged from matched aircraft)", insertedArgs[10])
	}
	if insertedArgs[11] != "BA123" {
		t.Errorf("flight arg = %v, want BA123 (merged from matched aircraft)", insertedArgs[11])
	}
}
