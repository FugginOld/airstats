package main

import (
	"testing"

	"github.com/jackc/pgx/v5"
)

func floatRow(v float64) pgx.Row {
	return &mockRow{scanFn: func(dest ...any) error {
		*(dest[0].(*float64)) = v
		return nil
	}}
}

// TestUpdateAircraftBySpeed_ColumnOrderFixed is a regression test for the
// indicated_air_speed/true_air_speed column swap that existed in the
// original updateFastestAircraft/updateSlowestAircraft: the insert must send
// Ias into indicated_air_speed and Tas into true_air_speed, not the reverse.
func TestUpdateAircraftBySpeed_ColumnOrderFixed(t *testing.T) {
	var capturedSQL string
	var capturedArgs []any

	db := &mockDB{
		queryRowQueue: []pgx.Row{
			floatRow(100), // getFastestAircraftFloor threshold
			noRowsRow(),   // DeleteExcessRows row count -> bails out early
		},
		batchFn: func(b *pgx.Batch) pgx.BatchResults {
			for _, q := range b.QueuedQueries {
				if capturedSQL == "" {
					capturedSQL = q.SQL
					capturedArgs = q.Arguments
				}
			}
			return &mockBatchResults{}
		},
	}

	aircraft := Aircraft{
		Hex: "ABCDEF", Flight: "BA123", R: "G-ABCD", T: "B738",
		Gs: 500, Ias: 111, Tas: 222,
		FastestProcessed: false,
	}

	updateAircraftBySpeed(newTestPG(db), []Aircraft{aircraft}, FastestAircraft)

	if capturedSQL == "" {
		t.Fatal("expected an INSERT to be queued, got none")
	}
	if len(capturedArgs) != 9 {
		t.Fatalf("len(args) = %d, want 9", len(capturedArgs))
	}
	if capturedArgs[7] != 111 {
		t.Errorf("indicated_air_speed arg = %v, want Ias (111)", capturedArgs[7])
	}
	if capturedArgs[8] != 222 {
		t.Errorf("true_air_speed arg = %v, want Tas (222)", capturedArgs[8])
	}
}

// TestUpdateAircraftBySpeed_FiltersNearZeroForSlowestOnly pins the
// near-zero-reading guard to the record it actually applies to: slowest
// (minimum-seeking) filters noise near zero, fastest (maximum-seeking) does
// not need to.
func TestUpdateAircraftBySpeed_FiltersNearZeroForSlowestOnly(t *testing.T) {
	countInsertBatches := func(b *pgx.Batch, count *int) pgx.BatchResults {
		for _, q := range b.QueuedQueries {
			if len(q.Arguments) == 9 {
				*count++
			}
		}
		return &mockBatchResults{}
	}

	var slowestCount int
	slowestDB := &mockDB{
		queryRowQueue: []pgx.Row{floatRow(999), noRowsRow()},
		batchFn:       func(b *pgx.Batch) pgx.BatchResults { return countInsertBatches(b, &slowestCount) },
	}
	nearZeroSlow := Aircraft{Hex: "A", Gs: 0.5, SlowestProcessed: false}
	updateAircraftBySpeed(newTestPG(slowestDB), []Aircraft{nearZeroSlow}, SlowestAircraft)
	if slowestCount != 0 {
		t.Errorf("SlowestAircraft: near-zero Gs was inserted (%d times), want filtered out", slowestCount)
	}

	var fastestCount int
	fastestDB := &mockDB{
		queryRowQueue: []pgx.Row{floatRow(-999), noRowsRow()},
		batchFn:       func(b *pgx.Batch) pgx.BatchResults { return countInsertBatches(b, &fastestCount) },
	}
	nearZeroFast := Aircraft{Hex: "A", Gs: 0.5, FastestProcessed: false}
	updateAircraftBySpeed(newTestPG(fastestDB), []Aircraft{nearZeroFast}, FastestAircraft)
	if fastestCount != 1 {
		t.Errorf("FastestAircraft: near-zero Gs was filtered (inserted %d times), want inserted", fastestCount)
	}
}

// TestUpdateAircraftByAltitude_NoUnprocessedAircraft exercises the
// early-return path: no unprocessed aircraft means no DB calls at all.
func TestUpdateAircraftByAltitude_NoUnprocessedAircraft(t *testing.T) {
	db := &mockDB{}
	aircraft := Aircraft{Hex: "A", HighestProcessed: true}
	updateAircraftByAltitude(newTestPG(db), []Aircraft{aircraft}, HighestAircraft)
}
