package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newRouter creates a fresh gin router with all API routes registered against
// the given server, mirroring the route layout in APIServer.Start().
func newRouter(s *APIServer) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("api")
	stats := api.Group("/stats")
	{
		stats.GET("/above", s.getAboveStats)

		stats.GET("/seen/flights", s.getFlightsSeenMetrics)
		stats.GET("/seen/aircraft", s.getAircraftSeenMetrics)

		stats.GET("/routes/metrics", s.getRouteMetrics)
		stats.GET("/routes/airlines", s.getTopAirlines)
		stats.GET("/routes/routes", s.getTopRoutes)
		stats.GET("/routes/countries-destination", func(c *gin.Context) { s.getTopCountries(c, DestinationCountry) })
		stats.GET("/routes/countries-origin", func(c *gin.Context) { s.getTopCountries(c, OriginCountry) })
		stats.GET("/routes/airports-domestic", func(c *gin.Context) { s.getTopAirports(c, DomesticAirports) })
		stats.GET("/routes/airports-international", func(c *gin.Context) { s.getTopAirports(c, InternationalAirports) })

		stats.GET("/motion/fastest", func(c *gin.Context) { s.getAircraftBySpeed(c, FastestAircraft) })
		stats.GET("/motion/slowest", func(c *gin.Context) { s.getAircraftBySpeed(c, SlowestAircraft) })
		stats.GET("/motion/highest", func(c *gin.Context) { s.getAircraftByAltitude(c, HighestAircraft) })
		stats.GET("/motion/lowest", func(c *gin.Context) { s.getAircraftByAltitude(c, LowestAircraft) })

		stats.GET("/interesting/metrics", s.getInterestingMetrics)
		stats.GET("/interesting/civilian", func(c *gin.Context) { s.getRecentInterestingAircraft(c, Civilian) })
		stats.GET("/interesting/police", func(c *gin.Context) { s.getRecentInterestingAircraft(c, Police) })
		stats.GET("/interesting/military", func(c *gin.Context) { s.getRecentInterestingAircraft(c, Military) })
		stats.GET("/interesting/government", func(c *gin.Context) { s.getRecentInterestingAircraft(c, Government) })

		stats.GET("/types/flights/all", func(c *gin.Context) { s.getTopAircraftTypes(c, PeriodAll, CountFlights) })
		stats.GET("/types/aircraft/all", func(c *gin.Context) { s.getTopAircraftTypes(c, PeriodAll, CountAircraft) })

		stats.GET("/charts/flights/day", func(c *gin.Context) { s.getChartOverTime(c, PeriodDay, CountFlights) })
		stats.GET("/charts/aircraft/day", func(c *gin.Context) { s.getChartOverTime(c, PeriodDay, CountAircraft) })
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
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("doPUT: json.Marshal: %v", err)
	}
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
		rawURL string
		want   string
	}{
		{"/", "UTC"},
		{"/?tz=America%2FNew_York", "America/New_York"},
		{"/?tz=Invalid%2FZone", "UTC"},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodGet, tc.rawURL, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		got := s.getTimezone(c)
		if got != tc.want {
			t.Errorf("URL=%q getTimezone() = %q, want %q", tc.rawURL, got, tc.want)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Wiring tests: each of these just proves the route calls the right
// StatsService method and maps its result/error to the right HTTP response.
// The SQL/scan logic itself is covered directly in stats_service_test.go.
// ──────────────────────────────────────────────────────────────────────────────

func TestHandlers_Success(t *testing.T) {
	tests := []struct {
		name string
		path string
		db   *mockDB
	}{
		{"FlightsSeenMetrics", "/api/stats/seen/flights", &mockDB{queryRowQueue: []pgx.Row{intRow(1), intRow(1), intRow(1)}}},
		{"AircraftSeenMetrics", "/api/stats/seen/aircraft", &mockDB{queryRowQueue: []pgx.Row{intRow(1), intRow(1), intRow(1)}}},
		{"RouteMetrics", "/api/stats/routes/metrics", &mockDB{queryRowQueue: []pgx.Row{intRow(1), intRow(1), intRow(1)}}},
		{"InterestingMetrics", "/api/stats/interesting/metrics", &mockDB{queryRowQueue: []pgx.Row{intRow(1), intRow(1), intRow(1)}}},
		{"TopAirlines", "/api/stats/routes/airlines", &mockDB{queryQueue: []mockQueryResult{{rows: emptyRows()}}}},
		{"TopRoutes", "/api/stats/routes/routes", &mockDB{queryQueue: []mockQueryResult{{rows: emptyRows()}}}},
		{"TopCountriesDestination", "/api/stats/routes/countries-destination", &mockDB{queryQueue: []mockQueryResult{{rows: emptyRows()}}}},
		{"TopCountriesOrigin", "/api/stats/routes/countries-origin", &mockDB{queryQueue: []mockQueryResult{{rows: emptyRows()}}}},
		{"TopAirportsDomestic", "/api/stats/routes/airports-domestic", &mockDB{queryQueue: []mockQueryResult{{rows: emptyRows()}}}},
		{"TopAirportsInternational", "/api/stats/routes/airports-international", &mockDB{queryQueue: []mockQueryResult{{rows: emptyRows()}}}},
		{"Fastest", "/api/stats/motion/fastest", &mockDB{queryRowQueue: []pgx.Row{noRowsRow()}, queryQueue: []mockQueryResult{{rows: emptyRows()}}}},
		{"Slowest", "/api/stats/motion/slowest", &mockDB{queryRowQueue: []pgx.Row{noRowsRow()}, queryQueue: []mockQueryResult{{rows: emptyRows()}}}},
		{"Highest", "/api/stats/motion/highest", &mockDB{queryRowQueue: []pgx.Row{noRowsRow()}, queryQueue: []mockQueryResult{{rows: emptyRows()}}}},
		{"Lowest", "/api/stats/motion/lowest", &mockDB{queryRowQueue: []pgx.Row{noRowsRow()}, queryQueue: []mockQueryResult{{rows: emptyRows()}}}},
		{"InterestingCivilian", "/api/stats/interesting/civilian", &mockDB{queryRowQueue: []pgx.Row{noRowsRow()}, queryQueue: []mockQueryResult{{rows: emptyRows()}}}},
		{"TypesFlights", "/api/stats/types/flights/all", &mockDB{queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
			*(dest[0].(*string)) = "B738"
			*(dest[1].(*int)) = 1
			*(dest[2].(*float64)) = 1
			return nil
		})}}}},
		{"ChartsFlightsDay", "/api/stats/charts/flights/day", &mockDB{}},
		{"ChartsAircraftDay", "/api/stats/charts/aircraft/day", &mockDB{}},
		{"Settings", "/api/settings", &mockDB{queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
			*(dest[0].(*int)) = 1
			*(dest[1].(*string)) = "k"
			*(dest[2].(*string)) = "v"
			*(dest[3].(*string)) = "d"
			return nil
		})}}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newRouter(newTestServer(tc.db))
			w := doGET(t, r, tc.path)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
			}
		})
	}
}

func TestHandlers_DBError_Returns500(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"TopRoutes", "/api/stats/routes/routes"},
		{"Fastest", "/api/stats/motion/fastest"},
		{"InterestingCivilian", "/api/stats/interesting/civilian"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := &mockDB{
				queryRowQueue: []pgx.Row{noRowsRow()},
				queryQueue:    []mockQueryResult{{err: errors.New("db down")}},
			}
			r := newRouter(newTestServer(db))
			w := doGET(t, r, tc.path)
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want 500", w.Code)
			}
		})
	}
}

func TestGetFlightsSeenMetrics_DBError_StillReturns200(t *testing.T) {
	// Metrics handlers are best-effort: a failed sub-query still yields 200.
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
// /api/settings PUT
// ──────────────────────────────────────────────────────────────────────────────

func TestUpdateSettingsHandler_Success(t *testing.T) {
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
// /api/stats/above — env-driven early-return behavior lives in the handler,
// not StatsService, so it stays covered here.
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
