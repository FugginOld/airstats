package main

import (
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestDeleteExcessRows_UnknownTableRejected(t *testing.T) {
	called := false
	db := &mockDB{queryRowQueue: []pgx.Row{&mockRow{scanFn: func(dest ...any) error {
		called = true
		*(dest[0].(*int)) = 999
		return nil
	}}}}

	DeleteExcessRows(newTestPG(db), "not_a_real_table", "ground_speed", "ASC", 50)
	if called {
		t.Error("DeleteExcessRows queried the DB for an unknown table, want rejected before any DB call")
	}
}

func TestDeleteExcessRows_UnknownMetricRejected(t *testing.T) {
	called := false
	db := &mockDB{queryRowQueue: []pgx.Row{&mockRow{scanFn: func(dest ...any) error {
		called = true
		*(dest[0].(*int)) = 999
		return nil
	}}}}

	DeleteExcessRows(newTestPG(db), "fastest_aircraft", "not_a_real_metric", "ASC", 50)
	if called {
		t.Error("DeleteExcessRows queried the DB for an unknown metric, want rejected before any DB call")
	}
}

func TestDeleteExcessRows_UnknownSortOrderRejected(t *testing.T) {
	called := false
	db := &mockDB{queryRowQueue: []pgx.Row{&mockRow{scanFn: func(dest ...any) error {
		called = true
		*(dest[0].(*int)) = 999
		return nil
	}}}}

	DeleteExcessRows(newTestPG(db), "fastest_aircraft", "ground_speed", "SIDEWAYS", 50)
	if called {
		t.Error("DeleteExcessRows queried the DB for an unknown sort order, want rejected before any DB call")
	}
}

func TestDeleteExcessRows_DeletesWhenOverLimit(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{intRow(60)}, // row count, > maxRows(50)
		execQueue:     []mockExecResult{{}},
	}

	// No SQL/args capture available on mockDB.Exec; this just proves the
	// over-limit path runs the delete without erroring.
	DeleteExcessRows(newTestPG(db), "fastest_aircraft", "ground_speed", "ASC", 50)
}

func TestMarkProcessed_UnknownColumnRejected(t *testing.T) {
	called := false
	db := &mockDB{batchFn: func(b *pgx.Batch) pgx.BatchResults {
		called = true
		return &mockBatchResults{}
	}}

	MarkProcessed(newTestPG(db), "not_a_real_column", []Aircraft{{Id: 1}})
	if called {
		t.Error("MarkProcessed sent a batch for an unknown column, want rejected before any DB call")
	}
}

func TestMarkProcessed_UpdatesEachAircraft(t *testing.T) {
	var batchCount int
	db := &mockDB{batchFn: func(b *pgx.Batch) pgx.BatchResults {
		batchCount = len(b.QueuedQueries)
		return &mockBatchResults{}
	}}

	MarkProcessed(newTestPG(db), "fastest_aircraft_processed", []Aircraft{{Id: 1}, {Id: 2}, {Id: 3}})
	if batchCount != 3 {
		t.Errorf("batchCount = %d, want 3", batchCount)
	}
}
