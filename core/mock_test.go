package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ──────────────────────────── mockRow ────────────────────────────────────────

type mockRow struct {
	scanFn func(dest ...any) error
}

func (r *mockRow) Scan(dest ...any) error { return r.scanFn(dest...) }

// errRow returns a Row whose Scan returns the given error.
func errRow(err error) pgx.Row { return &mockRow{scanFn: func(_ ...any) error { return err }} }

// noRowsRow returns a Row whose Scan returns pgx.ErrNoRows.
func noRowsRow() pgx.Row { return errRow(pgx.ErrNoRows) }

// intRow returns a Row that scans a single int value into dest[0].
func intRow(v int) pgx.Row {
	return &mockRow{scanFn: func(dest ...any) error {
		*(dest[0].(*int)) = v
		return nil
	}}
}

// ──────────────────────────── mockRows ───────────────────────────────────────

// mockRows implements pgx.Rows.  Each element of scanFns handles one row.
type mockRows struct {
	scanFns []func(dest ...any) error
	idx     int
}

// emptyRows returns a pgx.Rows with no data rows.
func emptyRows() pgx.Rows { return &mockRows{} }

// newMockRows creates a Rows with one entry per scan function provided.
func newMockRows(fns ...func(dest ...any) error) pgx.Rows {
	return &mockRows{scanFns: fns}
}

func (r *mockRows) Next() bool                                      { return r.idx < len(r.scanFns) }
func (r *mockRows) Close()                                          {}
func (r *mockRows) Err() error                                      { return nil }
func (r *mockRows) CommandTag() pgconn.CommandTag                   { return pgconn.CommandTag{} }
func (r *mockRows) FieldDescriptions() []pgconn.FieldDescription    { return nil }
func (r *mockRows) RawValues() [][]byte                             { return nil }
func (r *mockRows) Values() ([]any, error)                          { return nil, nil }
func (r *mockRows) Conn() *pgx.Conn                                 { return nil }
func (r *mockRows) Scan(dest ...any) error {
	if r.idx >= len(r.scanFns) {
		return errors.New("mockRows: no more rows")
	}
	fn := r.scanFns[r.idx]
	r.idx++
	return fn(dest...)
}

// ──────────────────────────── mockBatchResults ───────────────────────────────

type mockBatchResults struct{ execErr error }

func (b *mockBatchResults) Exec() (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("UPDATE 1"), b.execErr
}
func (b *mockBatchResults) Query() (pgx.Rows, error) { return emptyRows(), nil }
func (b *mockBatchResults) QueryRow() pgx.Row        { return noRowsRow() }
func (b *mockBatchResults) Close() error             { return nil }

// ──────────────────────────── mockTx ─────────────────────────────────────────

type mockTx struct {
	execFn      func(sql string, args ...any) (pgconn.CommandTag, error)
	commitErr   error
	rollbackErr error
}

func (t *mockTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if t.execFn != nil {
		return t.execFn(sql, args...)
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}
func (t *mockTx) Commit(ctx context.Context) error   { return t.commitErr }
func (t *mockTx) Rollback(ctx context.Context) error { return t.rollbackErr }

// Stub all remaining Tx interface methods that the production code never calls
// through settings/api paths.
func (t *mockTx) Begin(ctx context.Context) (pgx.Tx, error) { return nil, nil }
func (t *mockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (t *mockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return &mockBatchResults{}
}
func (t *mockTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }
func (t *mockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (t *mockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return emptyRows(), nil
}
func (t *mockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return noRowsRow()
}
func (t *mockTx) Conn() *pgx.Conn { return nil }

// ──────────────────────────── mockDB ─────────────────────────────────────────

type mockQueryResult struct {
	rows pgx.Rows
	err  error
}

type mockExecResult struct {
	tag pgconn.CommandTag
	err error
}

// mockDB is a queue-based mock: each call consumes the next pre-programmed
// response from the appropriate queue.  Queues that run dry return safe
// zero-value responses so tests only need to supply entries they care about.
type mockDB struct {
	queryRowQueue []pgx.Row
	queryQueue    []mockQueryResult
	execQueue     []mockExecResult
	beginFn       func() (pgx.Tx, error)
	batchFn       func(*pgx.Batch) pgx.BatchResults

	queryRowIdx int
	queryIdx    int
	execIdx     int
}

func (m *mockDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowIdx >= len(m.queryRowQueue) {
		return noRowsRow()
	}
	r := m.queryRowQueue[m.queryRowIdx]
	m.queryRowIdx++
	return r
}

func (m *mockDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryIdx >= len(m.queryQueue) {
		return emptyRows(), nil
	}
	r := m.queryQueue[m.queryIdx]
	m.queryIdx++
	return r.rows, r.err
}

func (m *mockDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.execIdx >= len(m.execQueue) {
		return pgconn.NewCommandTag("UPDATE 0"), nil
	}
	r := m.execQueue[m.execIdx]
	m.execIdx++
	return r.tag, r.err
}

func (m *mockDB) Begin(ctx context.Context) (pgx.Tx, error) {
	if m.beginFn != nil {
		return m.beginFn()
	}
	return &mockTx{}, nil
}

func (m *mockDB) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	if m.batchFn != nil {
		return m.batchFn(b)
	}
	return &mockBatchResults{}
}

func (m *mockDB) Ping(ctx context.Context) error { return nil }
func (m *mockDB) Close()                         {}

// ──────────────────────────── test helpers ───────────────────────────────────

// newTestServer builds an APIServer backed by the given mock.
func newTestServer(db *mockDB) *APIServer {
	pg := &postgres{db: db}
	return &APIServer{
		pg:       pg,
		port:     "8080",
		settings: NewSettingsService(pg),
	}
}

// newTestPG wraps a mockDB in a postgres struct for service-level tests.
func newTestPG(db *mockDB) *postgres {
	return &postgres{db: db}
}
