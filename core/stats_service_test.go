package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// scanAboveRow fills the 33 columns expected by StatsService.GetAboveStats.
func scanAboveRow(dest ...any) error {
	*(dest[0].(*string)) = "ABCDEF"    // hex
	*(dest[1].(*string)) = "BA123"     // flight
	*(dest[2].(*string)) = "G-ABCD"    // registration
	*(dest[3].(*string)) = "B738"      // type
	*(dest[4].(*float64)) = 45.0       // track
	*(dest[5].(**time.Time)) = nil     // first_seen
	*(dest[6].(**time.Time)) = nil     // last_seen
	*(dest[7].(*float64)) = 51.5       // last_seen_lat
	*(dest[8].(*float64)) = -0.1       // last_seen_lon
	*(dest[9].(*float64)) = 12.3       // last_seen_distance
	*(dest[10].(**float64)) = nil      // destination_distance
	*(dest[11].(**string)) = nil       // reg_type
	*(dest[12].(**string)) = nil       // icao_type
	*(dest[13].(**string)) = nil       // manufacturer
	*(dest[14].(**string)) = nil       // registered_owner_country_name
	*(dest[15].(**string)) = nil       // registered_owner_country_iso
	*(dest[16].(**string)) = nil       // registered_owner_operator_flag
	*(dest[17].(**string)) = nil       // registered_owner
	*(dest[18].(**string)) = nil       // url_photo
	*(dest[19].(**string)) = nil       // url_photo_thumbnail
	*(dest[20].(**string)) = nil       // airline_name
	*(dest[21].(**string)) = nil       // airline_icao
	*(dest[22].(**string)) = nil       // origin_country_name
	*(dest[23].(**string)) = nil       // origin_country_iso_name
	*(dest[24].(**string)) = nil       // origin_iata_code
	*(dest[25].(**string)) = nil       // origin_icao_code
	*(dest[26].(**string)) = nil       // origin_name
	*(dest[27].(**string)) = nil       // destination_country_name
	*(dest[28].(**string)) = nil       // destination_country_iso_name
	*(dest[29].(**string)) = nil       // destination_iata_code
	*(dest[30].(**string)) = nil       // destination_icao_code
	*(dest[31].(**string)) = nil       // destination_name
	*(dest[32].(**float64)) = nil      // route_distance
	return nil
}

// scanSpeedRow fills the 9 columns expected by GetAircraftBySpeed.
func scanSpeedRow(dest ...any) error {
	*(dest[0].(*string)) = "ABCDEF"  // hex
	*(dest[1].(*string)) = "BA123"   // flight
	*(dest[2].(*string)) = "G-ABCD"  // registration
	*(dest[3].(*string)) = "B738"    // type
	*(dest[4].(**time.Time)) = nil   // first_seen
	*(dest[5].(**time.Time)) = nil   // last_seen
	*(dest[6].(*float64)) = 550.5    // ground_speed
	*(dest[7].(*int)) = 450          // ias
	*(dest[8].(*int)) = 480          // tas
	return nil
}

// scanAltRow fills the 8 columns expected by GetAircraftByAltitude.
func scanAltRow(dest ...any) error {
	*(dest[0].(*string)) = "ABCDEF" // hex
	*(dest[1].(*string)) = "BA123"  // flight
	*(dest[2].(*string)) = "G-ABCD" // registration
	*(dest[3].(*string)) = "B738"   // type
	*(dest[4].(**time.Time)) = nil  // first_seen
	*(dest[5].(**time.Time)) = nil  // last_seen
	*(dest[6].(*int)) = 38000       // barometric_altitude
	*(dest[7].(*int)) = 39000       // geometric_altitude
	return nil
}

// scanRouteRow fills the 6 columns expected by GetTopRoutes.
func scanRouteRow(dest ...any) error {
	*(dest[0].(*string)) = "LHR → CDG"
	*(dest[1].(*string)) = "LHR"
	*(dest[2].(*string)) = "London Heathrow"
	*(dest[3].(*string)) = "CDG"
	*(dest[4].(*string)) = "Paris Charles de Gaulle"
	*(dest[5].(*int)) = 48
	return nil
}

// scanRecentInterestingRow fills the 14 columns expected by
// StatsService.GetRecentInterestingAircraft.
func scanRecentInterestingRow(dest ...any) error {
	*(dest[0].(*string)) = "ABCDEF"      // icao
	*(dest[1].(*string)) = "G-ABCD"      // registration
	*(dest[2].(*string)) = "Some Op"     // operator
	*(dest[3].(*string)) = "B738"        // type
	*(dest[4].(*string)) = "B738"        // icao_type
	*(dest[5].(*string)) = "Civ"         // group
	*(dest[6].(*string)) = "cat"         // category
	*(dest[7].(*string)) = "t1"          // tag1
	*(dest[8].(*string)) = "t2"          // tag2
	*(dest[9].(*string)) = "t3"          // tag3
	*(dest[10].(*string)) = "ABCDEF"     // hex
	*(dest[11].(*string)) = "BA123"      // flight
	*(dest[12].(**time.Time)) = nil      // seen
	*(dest[13].(*float64)) = 123.0       // seen_epoch
	return nil
}

func TestStatsService_GetFlightsSeenMetrics(t *testing.T) {
	db := &mockDB{queryRowQueue: []pgx.Row{intRow(1000), intRow(42), intRow(7)}}
	svc := NewStatsService(newTestPG(db))

	metrics, err := svc.GetFlightsSeenMetrics(context.Background(), "UTC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.TotalFlights != 1000 || metrics.TodayFlights != 42 || metrics.HourFlights != 7 {
		t.Errorf("metrics = %+v, want {1000 42 7}", metrics)
	}
}

func TestStatsService_GetFlightsSeenMetrics_PartialFailure(t *testing.T) {
	// Best-effort contract: a failed sub-query zero-values that field, but the
	// call as a whole still succeeds (mirrors the pre-existing handler test
	// TestGetFlightsSeenMetrics_DBError_StillReturns200 in api_test.go).
	db := &mockDB{queryRowQueue: []pgx.Row{errRow(errors.New("conn lost")), intRow(42), errRow(errors.New("conn lost"))}}
	svc := NewStatsService(newTestPG(db))

	metrics, err := svc.GetFlightsSeenMetrics(context.Background(), "UTC")
	if err != nil {
		t.Fatalf("expected no error (best-effort), got %v", err)
	}
	if metrics.TotalFlights != 0 || metrics.TodayFlights != 42 || metrics.HourFlights != 0 {
		t.Errorf("metrics = %+v, want {0 42 0}", metrics)
	}
}

func TestStatsService_GetAircraftSeenMetrics(t *testing.T) {
	db := &mockDB{queryRowQueue: []pgx.Row{intRow(500), intRow(30), intRow(3)}}
	svc := NewStatsService(newTestPG(db))

	metrics, err := svc.GetAircraftSeenMetrics(context.Background(), "UTC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.TotalAircraft != 500 {
		t.Errorf("TotalAircraft = %d, want 500", metrics.TotalAircraft)
	}
}

func TestStatsService_GetRouteMetrics(t *testing.T) {
	db := &mockDB{queryRowQueue: []pgx.Row{intRow(800), intRow(60), intRow(150)}}
	svc := NewStatsService(newTestPG(db))

	metrics, err := svc.GetRouteMetrics(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.TotalRoutes != 800 || metrics.UniqueCountries != 60 || metrics.UniqueAirports != 150 {
		t.Errorf("metrics = %+v", metrics)
	}
}

func TestStatsService_GetInterestingMetrics(t *testing.T) {
	db := &mockDB{queryRowQueue: []pgx.Row{intRow(12), intRow(2), intRow(1)}}
	svc := NewStatsService(newTestPG(db))

	metrics, err := svc.GetInterestingMetrics(context.Background(), "UTC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.TotalInteresting != 12 {
		t.Errorf("TotalInteresting = %d, want 12", metrics.TotalInteresting)
	}
}

func TestStatsService_GetAboveStats(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{rows: newMockRows(scanAboveRow)}}}
	svc := NewStatsService(newTestPG(db))

	aircraft, err := svc.GetAboveStats(context.Background(), 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(aircraft) != 1 || aircraft[0].Hex != "ABCDEF" {
		t.Errorf("aircraft = %+v", aircraft)
	}
}

func TestStatsService_GetAboveStats_DBError(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{err: errors.New("db down")}}}
	svc := NewStatsService(newTestPG(db))

	if _, err := svc.GetAboveStats(context.Background(), 50); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestStatsService_GetRecentInterestingAircraft(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{rows: newMockRows(scanRecentInterestingRow)}}}
	svc := NewStatsService(newTestPG(db))

	aircraft, err := svc.GetRecentInterestingAircraft(context.Background(), Civilian, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(aircraft) != 1 || aircraft[0].Icao != "ABCDEF" {
		t.Errorf("aircraft = %+v", aircraft)
	}
}

func TestStatsService_GetRecentInterestingAircraft_DBError(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{err: errors.New("db down")}}}
	svc := NewStatsService(newTestPG(db))

	if _, err := svc.GetRecentInterestingAircraft(context.Background(), Police, 5); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestStatsService_GetAircraftBySpeed(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{rows: newMockRows(scanSpeedRow)}}}
	svc := NewStatsService(newTestPG(db))

	aircraft, err := svc.GetAircraftBySpeed(context.Background(), FastestAircraft, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(aircraft) != 1 || aircraft[0].GroundSpeed != 550.5 {
		t.Errorf("aircraft = %+v", aircraft)
	}
}

func TestStatsService_GetAircraftByAltitude(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{rows: newMockRows(scanAltRow)}}}
	svc := NewStatsService(newTestPG(db))

	aircraft, err := svc.GetAircraftByAltitude(context.Background(), HighestAircraft, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(aircraft) != 1 || aircraft[0].BarometricAltitude != 38000 {
		t.Errorf("aircraft = %+v", aircraft)
	}
}

func TestStatsService_GetTopAircraftTypes(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
		*(dest[0].(*string)) = "B738"
		*(dest[1].(*int)) = 55
		*(dest[2].(*float64)) = 22.0
		return nil
	})}}}
	svc := NewStatsService(newTestPG(db))

	types, err := svc.GetTopAircraftTypes(context.Background(), PeriodAll, CountFlights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(types) != 1 || types[0].AircraftType != "B738" {
		t.Errorf("types = %+v", types)
	}
}

func TestStatsService_GetTopRoutes(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{rows: newMockRows(scanRouteRow)}}}
	svc := NewStatsService(newTestPG(db))

	routes, err := svc.GetTopRoutes(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(routes) != 1 || routes[0].FlightCount != 48 {
		t.Errorf("routes = %+v", routes)
	}
}

func TestStatsService_GetTopCountries(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
		*(dest[0].(*string)) = "France"
		*(dest[1].(*string)) = "FR"
		*(dest[2].(*int)) = 150
		return nil
	})}}}
	svc := NewStatsService(newTestPG(db))

	countries, err := svc.GetTopCountries(context.Background(), DestinationCountry, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(countries) != 1 || countries[0].CountryISO != "FR" {
		t.Errorf("countries = %+v", countries)
	}
}

func TestStatsService_GetTopAirlines(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
		*(dest[0].(*string)) = "British Airways"
		*(dest[1].(*string)) = "BAW"
		*(dest[2].(*string)) = "BA"
		*(dest[3].(*int)) = 200
		return nil
	})}}}
	svc := NewStatsService(newTestPG(db))

	airlines, err := svc.GetTopAirlines(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(airlines) != 1 || airlines[0].AirlineICAO != "BAW" {
		t.Errorf("airlines = %+v", airlines)
	}
}

func TestStatsService_GetTopAirports(t *testing.T) {
	db := &mockDB{queryQueue: []mockQueryResult{{rows: newMockRows(func(dest ...any) error {
		*(dest[0].(*string)) = "LHR"
		*(dest[1].(*string)) = "London Heathrow"
		*(dest[2].(*string)) = "United Kingdom"
		*(dest[3].(*int)) = 500
		return nil
	})}}}
	svc := NewStatsService(newTestPG(db))

	airports, err := svc.GetTopAirports(context.Background(), DomesticAirports, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(airports) != 1 || airports[0].AirportCode != "LHR" {
		t.Errorf("airports = %+v", airports)
	}
}

func TestStatsService_GetChartOverTime(t *testing.T) {
	svc := NewStatsService(newTestPG(&mockDB{}))

	chart, err := svc.GetChartOverTime(context.Background(), "UTC", PeriodDay, CountFlights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chart.Series[0].ID != "flights_day" {
		t.Errorf("seriesID = %q, want flights_day", chart.Series[0].ID)
	}
}

func TestStatsService_GetChartOverTime_Aircraft(t *testing.T) {
	svc := NewStatsService(newTestPG(&mockDB{}))

	chart, err := svc.GetChartOverTime(context.Background(), "UTC", PeriodYear, CountAircraft)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chart.Series[0].ID != "aircraft_year" {
		t.Errorf("seriesID = %q, want aircraft_year", chart.Series[0].ID)
	}
}

func TestStatsService_GetChartOverTime_BeginError(t *testing.T) {
	db := &mockDB{beginFn: func() (pgx.Tx, error) { return nil, errors.New("begin failed") }}
	svc := NewStatsService(newTestPG(db))

	if _, err := svc.GetChartOverTime(context.Background(), "UTC", PeriodYear, CountAircraft); err == nil {
		t.Fatal("expected error, got nil")
	}
}
