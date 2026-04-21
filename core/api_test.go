package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newRouter creates a fresh gin router with all API routes registered against
// the given server, following the same layout as APIServer.Start().
func newRouter(s *APIServer) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("api")
	stats := api.Group("/stats")
	{
		stats.GET("/seen/flights", s.getFlightsSeenMetrics)
		stats.GET("/seen/aircraft", s.getAircraftSeenMetrics)
		stats.GET("/routes/metrics", s.getRouteMetrics)
		stats.GET("/routes/airlines", s.getTopAirlines)
		stats.GET("/routes/routes", s.getTopRoutes)
		stats.GET("/routes/countries-destination", s.getTopDestinationCountries)
		stats.GET("/routes/countries-origin", s.getTopOriginCountries)
		stats.GET("/routes/airports-domestic", s.getTopDomesticAirports)
		stats.GET("/routes/airports-international", s.getTopInternationalAirports)
		stats.GET("/motion/fastest", s.getFastestAircraft)
		stats.GET("/motion/slowest", s.getSlowestAircraft)
		stats.GET("/motion/highest", s.getHighestAircraft)
		stats.GET("/motion/lowest", s.getLowestAircraft)
		stats.GET("/interesting/metrics", s.getInterestingMetrics)
		stats.GET("/interesting/civilian", func(c *gin.Context) { s.getRecentInterestingAircraft(c, "Civ") })
		stats.GET("/types/flights/all", func(c *gin.Context) { s.getTopAircraftTypes(c, "all", "flights") })
		stats.GET("/types/aircraft/all", func(c *gin.Context) { s.getTopAircraftTypes(c, "all", "aircraft") })
		stats.GET("/types/flights/bad", func(c *gin.Context) { s.getTopAircraftTypes(c, "all", "invalid") })
		stats.GET("/above", s.getAboveStats)
	}
	settings := api.Group("/settings")
	{
		settings.GET("", s.getSettings)
		settings.PUT("", s.updateSettings)
	}
	api.GET("/version", s.getVersion)

	return r
}

func doGET(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	r.ServeHTTP(w, req)
	return w
}

func doPUT(t *testing.T, r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// ──────────────────────────────────────────────────────────────────────────────
// /api/version
// ──────────────────────────────────────────────────────────────────────────────

func TestGetVersion(t *testing.T) {
	r := newRouter(newTestServer(&mockDB{}))
	w := doGET(t, r, "/api/version")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"version", "commit", "date"} {
		if _, ok := resp[key]; !ok {
			t.Errorf("response missing key %q", key)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// getTimezone helper
// ──────────────────────────────────────────────────────────────────────────────

func TestGetTimezone(t *testing.T) {
	s := newTestServer(&mockDB{})

	tests := []struct {
		query string
		want  string
	}{
		{"", "UTC"},
		{"?tz=America%2FNew_York", "America/New_York"},
		{"?tz=Invalid%2FZone", "UTC"},
	}

	for _, tc := range tests {
		r := newRouter(s)
		// We use getFlightsSeenMetrics as a carrier; we just want the TZ to be
		// applied without error.  The mock returns 0 for all three counters.
		db := &mockDB{
			queryRowQueue: []pgx.Row{intRow(0), intRow(0), intRow(0)},
		}
		s2 := newTestServer(db)
		r2 := newRouter(s2)
		w := doGET(t, r2, "/api/stats/seen/flights"+tc.query)
		if w.Code != http.StatusOK {
			t.Errorf("query=%q status=%d, want 200", tc.query, w.Code)
		}
		_ = r
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// /api/stats/seen/flights
// ──────────────────────────────────────────────────────────────────────────────

func TestGetFlightsSeenMetrics(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{intRow(1000), intRow(42), intRow(7)},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/seen/flights")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["total_flights"] != float64(1000) {
		t.Errorf("total_flights = %v, want 1000", resp["total_flights"])
	}
	if resp["today_flights"] != float64(42) {
		t.Errorf("today_flights = %v, want 42", resp["today_flights"])
	}
	if resp["hour_flights"] != float64(7) {
		t.Errorf("hour_flights = %v, want 7", resp["hour_flights"])
	}
}

func TestGetFlightsSeenMetrics_DBError_StillReturns200(t *testing.T) {
	// When DB calls fail the handler still returns 200 with partial/empty stats.
	db := &mockDB{
		queryRowQueue: []pgx.Row{errRow(errors.New("conn lost")), errRow(errors.New("conn lost")), errRow(errors.New("conn lost"))},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/seen/flights")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// /api/stats/seen/aircraft
// ──────────────────────────────────────────────────────────────────────────────

func TestGetAircraftSeenMetrics(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{intRow(500), intRow(30), intRow(3)},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/seen/aircraft")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["total_aircraft"] != float64(500) {
		t.Errorf("total_aircraft = %v, want 500", resp["total_aircraft"])
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// /api/stats/routes/metrics
// ──────────────────────────────────────────────────────────────────────────────

func TestGetRouteMetrics(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{intRow(800), intRow(60), intRow(150)},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/routes/metrics")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["total_routes"] != float64(800) {
		t.Errorf("total_routes = %v, want 800", resp["total_routes"])
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// /api/stats/interesting/metrics
// ──────────────────────────────────────────────────────────────────────────────

func TestGetInterestingMetrics(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{intRow(12), intRow(2), intRow(1)},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/interesting/metrics")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["total_interesting"] != float64(12) {
		t.Errorf("total_interesting = %v, want 12", resp["total_interesting"])
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// /api/settings GET
// ──────────────────────────────────────────────────────────────────────────────

func TestGetSettingsHandler_Success(t *testing.T) {
	db := &mockDB{
		queryQueue: []mockQueryResult{
			{rows: newMockRows(
				func(dest ...any) error {
					*(dest[0].(*int)) = 1
					*(dest[1].(*string)) = "route_table_limit"
					*(dest[2].(*string)) = "5"
					*(dest[3].(*string)) = "desc"
					return nil
				},
			)},
		},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/settings")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp []any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("len(resp) = %d, want 1", len(resp))
	}
}

func TestGetSettingsHandler_DBError(t *testing.T) {
	db := &mockDB{
		queryQueue: []mockQueryResult{{err: errors.New("db down")}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/settings")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// /api/settings PUT
// ──────────────────────────────────────────────────────────────────────────────

func TestUpdateSettingsHandler_Success(t *testing.T) {
	// beginFn returns ok tx (default mockTx), then GetAllSettings Query returns
	// one row.
	db := &mockDB{
		queryQueue: []mockQueryResult{
			{rows: newMockRows(
				func(dest ...any) error {
					*(dest[0].(*int)) = 1
					*(dest[1].(*string)) = "route_table_limit"
					*(dest[2].(*string)) = "7"
					*(dest[3].(*string)) = "desc"
					return nil
				},
			)},
		},
	}
	r := newRouter(newTestServer(db))
	w := doPUT(t, r, "/api/settings", map[string]string{"route_table_limit": "7"})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestUpdateSettingsHandler_BadJSON(t *testing.T) {
	r := newRouter(newTestServer(&mockDB{}))
	ww := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(ww, req)

	if ww.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", ww.Code)
	}
}

func TestUpdateSettingsHandler_DBError(t *testing.T) {
	db := &mockDB{
		beginFn: func() (pgx.Tx, error) {
			return nil, errors.New("begin failed")
		},
	}
	r := newRouter(newTestServer(db))
	w := doPUT(t, r, "/api/settings", map[string]string{"k": "v"})

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Motion endpoints (fastest/slowest/highest/lowest)
// ──────────────────────────────────────────────────────────────────────────────

// scanSpeedRow fills the 9 columns expected by getFastestAircraft / getSlowestAircraft.
func scanSpeedRow(dest ...any) error {
	*(dest[0].(*string)) = "ABCDEF"   // hex
	*(dest[1].(*string)) = "BA123"    // flight
	*(dest[2].(*string)) = "G-ABCD"   // registration
	*(dest[3].(*string)) = "B738"     // type
	*(dest[4].(**time.Time)) = nil    // first_seen
	*(dest[5].(**time.Time)) = nil    // last_seen
	*(dest[6].(*float64)) = 550.5    // ground_speed
	*(dest[7].(*int)) = 450           // ias
	*(dest[8].(*int)) = 480           // tas
	return nil
}

// scanAltRow fills the 8 columns expected by getHighestAircraft / getLowestAircraft.
func scanAltRow(dest ...any) error {
	*(dest[0].(*string)) = "ABCDEF"   // hex
	*(dest[1].(*string)) = "BA123"    // flight
	*(dest[2].(*string)) = "G-ABCD"   // registration
	*(dest[3].(*string)) = "B738"     // type
	*(dest[4].(**time.Time)) = nil    // first_seen
	*(dest[5].(**time.Time)) = nil    // last_seen
	*(dest[6].(*int)) = 38000         // barometric_altitude
	*(dest[7].(*int)) = 39000         // geometric_altitude
	return nil
}

func TestGetFastestAircraft(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()}, // for getLimit
		queryQueue:    []mockQueryResult{{rows: newMockRows(scanSpeedRow)}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/motion/fastest")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp []any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("len(resp) = %d, want 1", len(resp))
	}
}

func TestGetFastestAircraft_DBError(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue:    []mockQueryResult{{err: errors.New("db down")}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/motion/fastest")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}

func TestGetSlowestAircraft(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue:    []mockQueryResult{{rows: newMockRows(scanSpeedRow)}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/motion/slowest")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestGetHighestAircraft(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue:    []mockQueryResult{{rows: newMockRows(scanAltRow)}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/motion/highest")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp []any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("len(resp) = %d, want 1", len(resp))
	}
}

func TestGetLowestAircraft(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue:    []mockQueryResult{{rows: newMockRows(scanAltRow)}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/motion/lowest")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Route table endpoints
// ──────────────────────────────────────────────────────────────────────────────

func scanRouteRow(dest ...any) error {
	*(dest[0].(*string)) = "LHR → CDG"
	*(dest[1].(*string)) = "LHR"
	*(dest[2].(*string)) = "London Heathrow"
	*(dest[3].(*string)) = "CDG"
	*(dest[4].(*string)) = "Paris Charles de Gaulle"
	*(dest[5].(*int)) = 48
	return nil
}

func TestGetTopRoutes(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue:    []mockQueryResult{{rows: newMockRows(scanRouteRow)}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/routes/routes")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp []any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("len(resp) = %d, want 1", len(resp))
	}
}

func TestGetTopRoutes_DBError(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue:    []mockQueryResult{{err: errors.New("db down")}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/routes/routes")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}

func TestGetTopAirlines(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
			*(dest[0].(*string)) = "British Airways"
			*(dest[1].(*string)) = "BAW"
			*(dest[2].(*string)) = "BA"
			*(dest[3].(*int)) = 200
			return nil
		})}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/routes/airlines")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestGetTopDestinationCountries(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
			*(dest[0].(*string)) = "France"
			*(dest[1].(*string)) = "FR"
			*(dest[2].(*int)) = 150
			return nil
		})}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/routes/countries-destination")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestGetTopOriginCountries(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
			*(dest[0].(*string)) = "United Kingdom"
			*(dest[1].(*string)) = "GB"
			*(dest[2].(*int)) = 120
			return nil
		})}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/routes/countries-origin")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestGetTopDomesticAirports(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
			*(dest[0].(*string)) = "LHR"
			*(dest[1].(*string)) = "London Heathrow"
			*(dest[2].(*string)) = "United Kingdom"
			*(dest[3].(*int)) = 500
			return nil
		})}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/routes/airports-domestic")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestGetTopInternationalAirports(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
			*(dest[0].(*string)) = "CDG"
			*(dest[1].(*string)) = "Paris Charles de Gaulle"
			*(dest[2].(*string)) = "France"
			*(dest[3].(*int)) = 300
			return nil
		})}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/routes/airports-international")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Aircraft types
// ──────────────────────────────────────────────────────────────────────────────

func TestGetTopAircraftTypes_Flights(t *testing.T) {
	db := &mockDB{
		queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
			*(dest[0].(*string)) = "B738"
			*(dest[1].(*int)) = 55
			*(dest[2].(*float64)) = 22.0
			return nil
		})}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/types/flights/all")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestGetTopAircraftTypes_Aircraft(t *testing.T) {
	db := &mockDB{
		queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
			*(dest[0].(*string)) = "A320"
			*(dest[1].(*int)) = 40
			*(dest[2].(*float64)) = 18.0
			return nil
		})}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/types/aircraft/all")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestGetTopAircraftTypes_InvalidParam(t *testing.T) {
	r := newRouter(newTestServer(&mockDB{}))
	w := doGET(t, r, "/api/stats/types/flights/bad")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// /api/stats/above
// ──────────────────────────────────────────────────────────────────────────────

func TestGetAboveStats_MissingRadius(t *testing.T) {
	t.Setenv("ABOVE_RADIUS", "")
	r := newRouter(newTestServer(&mockDB{}))
	w := doGET(t, r, "/api/stats/above")
	// Handler returns early without writing a response body; gin defaults to 200.
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (empty body)", w.Code)
	}
}

func TestGetAboveStats_InvalidRadius(t *testing.T) {
	t.Setenv("ABOVE_RADIUS", "not-a-number")
	r := newRouter(newTestServer(&mockDB{}))
	w := doGET(t, r, "/api/stats/above")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestGetAboveStats_DBError(t *testing.T) {
	t.Setenv("ABOVE_RADIUS", "50")
	db := &mockDB{
		queryQueue: []mockQueryResult{{err: errors.New("db down")}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/above")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}

func TestGetAboveStats_EmptyResult(t *testing.T) {
	t.Setenv("ABOVE_RADIUS", "50")
	db := &mockDB{
		queryQueue: []mockQueryResult{{rows: emptyRows()}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/above")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp []any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("len(resp) = %d, want 0", len(resp))
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// /api/stats/interesting/civilian
// ──────────────────────────────────────────────────────────────────────────────

func TestGetRecentInterestingAircraft_EmptyResult(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue:    []mockQueryResult{{rows: emptyRows()}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/interesting/civilian")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestGetRecentInterestingAircraft_DBError(t *testing.T) {
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue:    []mockQueryResult{{err: errors.New("db down")}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/interesting/civilian")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// getLimit helper (tested through a real handler)
// ──────────────────────────────────────────────────────────────────────────────

func TestGetLimit_FromDB(t *testing.T) {
	// Return a valid setting row so getLimit uses the DB value (10) instead of 5.
	db := &mockDB{
		queryRowQueue: []pgx.Row{
			&mockRow{scanFn: func(dest ...any) error {
				*(dest[0].(*int)) = 99
				*(dest[1].(*string)) = "record_holder_table_limit"
				*(dest[2].(*string)) = "10"
				*(dest[3].(*string)) = "desc"
				return nil
			}},
		},
		queryQueue: []mockQueryResult{{rows: emptyRows()}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/motion/fastest")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestGetLimit_Fallback(t *testing.T) {
	// When DB returns no row, getLimit should fall back to 5.
	db := &mockDB{
		queryRowQueue: []pgx.Row{noRowsRow()},
		queryQueue:    []mockQueryResult{{rows: emptyRows()}},
	}
	r := newRouter(newTestServer(db))
	w := doGET(t, r, "/api/stats/motion/fastest")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}
