package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// Typed control parameters for StatsService methods — deliberately not raw
// strings, so a typo can't silently reach the SQL built from them.

type SpeedRecord string

const (
	FastestAircraft SpeedRecord = "fastest_aircraft"
	SlowestAircraft SpeedRecord = "slowest_aircraft"
)

func (r SpeedRecord) direction() string {
	if r == FastestAircraft {
		return "DESC"
	}
	return "ASC"
}

// deleteSortOrder is the opposite of direction(): DeleteExcessRows prunes the
// least-extreme rows, which sit at the opposite end of the ordering that
// keeps the most extreme ones.
func (r SpeedRecord) deleteSortOrder() string {
	if r.direction() == "DESC" {
		return "ASC"
	}
	return "DESC"
}

type AltitudeRecord string

const (
	HighestAircraft AltitudeRecord = "highest_aircraft"
	LowestAircraft  AltitudeRecord = "lowest_aircraft"
)

func (r AltitudeRecord) direction() string {
	if r == HighestAircraft {
		return "DESC"
	}
	return "ASC"
}

// deleteSortOrder is the opposite of direction(): DeleteExcessRows prunes the
// least-extreme rows, which sit at the opposite end of the ordering that
// keeps the most extreme ones.
func (r AltitudeRecord) deleteSortOrder() string {
	if r.direction() == "DESC" {
		return "ASC"
	}
	return "DESC"
}

type CountrySide string

const (
	OriginCountry      CountrySide = "origin"
	DestinationCountry CountrySide = "destination"
)

func (s CountrySide) opposite() CountrySide {
	if s == OriginCountry {
		return DestinationCountry
	}
	return OriginCountry
}

type AirportScope string

const (
	DomesticAirports      AirportScope = "domestic"
	InternationalAirports AirportScope = "international"
)

func (a AirportScope) operator() string {
	if a == DomesticAirports {
		return "="
	}
	return "!="
}

type CountBasis string

const (
	CountFlights  CountBasis = "flights"
	CountAircraft CountBasis = "aircraft"
)

type Period string

const (
	PeriodAll   Period = "all"
	PeriodYear  Period = "year"
	PeriodMonth Period = "month"
	PeriodDay   Period = "day"
)

type InterestingGroup string

const (
	Civilian   InterestingGroup = "Civ"
	Police     InterestingGroup = "Pol"
	Military   InterestingGroup = "Mil"
	Government InterestingGroup = "Gov"
)

// StatsService owns the read-side SQL behind /api/stats — the repository
// seam between the HTTP handlers in api.go and Postgres. It never imports
// gin/net-http: methods return (T, error) and the caller maps errors to a
// response.
type StatsService struct {
	pg *postgres
}

func NewStatsService(pg *postgres) *StatsService {
	return &StatsService{pg: pg}
}

func (s *StatsService) GetFlightsSeenMetrics(ctx context.Context, tz string) (FlightsSeenMetrics, error) {
	var metrics FlightsSeenMetrics

	if err := s.pg.db.QueryRow(ctx, "SELECT COUNT(*) FROM aircraft_data").Scan(&metrics.TotalFlights); err != nil {
		log.Error().Err(err).Msg("Failed to count total flights")
	}

	if err := s.pg.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM aircraft_data WHERE DATE(first_seen AT TIME ZONE $1) = CURRENT_DATE", tz,
	).Scan(&metrics.TodayFlights); err != nil {
		log.Error().Err(err).Msg("Failed to count today's flights")
	}

	if err := s.pg.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM aircraft_data WHERE first_seen >= NOW() - INTERVAL '1 hour'",
	).Scan(&metrics.HourFlights); err != nil {
		log.Error().Err(err).Msg("Failed to count past-hour flights")
	}

	return metrics, nil
}

func (s *StatsService) GetAircraftSeenMetrics(ctx context.Context, tz string) (AircraftSeenMetrics, error) {
	var metrics AircraftSeenMetrics

	if err := s.pg.db.QueryRow(ctx, "SELECT COUNT(DISTINCT hex) FROM aircraft_data").Scan(&metrics.TotalAircraft); err != nil {
		log.Error().Err(err).Msg("Failed to count total aircraft")
	}

	if err := s.pg.db.QueryRow(ctx,
		"SELECT COUNT(DISTINCT hex) FROM aircraft_data WHERE DATE(first_seen AT TIME ZONE $1) = CURRENT_DATE", tz,
	).Scan(&metrics.TodayAircraft); err != nil {
		log.Error().Err(err).Msg("Failed to count today's aircraft")
	}

	if err := s.pg.db.QueryRow(ctx,
		"SELECT COUNT(DISTINCT hex) FROM aircraft_data WHERE first_seen >= NOW() - INTERVAL '1 hour'",
	).Scan(&metrics.HourAircraft); err != nil {
		log.Error().Err(err).Msg("Failed to count past-hour aircraft")
	}

	return metrics, nil
}

func (s *StatsService) GetRouteMetrics(ctx context.Context) (RouteMetrics, error) {
	var metrics RouteMetrics

	if err := s.pg.db.QueryRow(ctx,
		`SELECT COUNT(*)
			FROM aircraft_data a
			INNER JOIN route_data r ON a.flight = r.route_callsign`,
	).Scan(&metrics.TotalRoutes); err != nil {
		log.Error().Err(err).Msg("Failed to count total routes")
	}

	if err := s.pg.db.QueryRow(ctx,
		`SELECT COUNT(*)
			FROM (
				SELECT origin_country_name AS country FROM route_data
				UNION
				SELECT destination_country_name AS country FROM route_data
			) AS unique_countries`,
	).Scan(&metrics.UniqueCountries); err != nil {
		log.Error().Err(err).Msg("Failed to count unique countries")
	}

	if err := s.pg.db.QueryRow(ctx,
		`SELECT COUNT(*)
			FROM (
				SELECT origin_icao_code AS airport FROM route_data
				UNION
				SELECT destination_icao_code AS airport FROM route_data
			) AS unique_airports`,
	).Scan(&metrics.UniqueAirports); err != nil {
		log.Error().Err(err).Msg("Failed to count unique airports")
	}

	return metrics, nil
}

func (s *StatsService) GetInterestingMetrics(ctx context.Context, tz string) (InterestingMetrics, error) {
	var metrics InterestingMetrics

	if err := s.pg.db.QueryRow(ctx, "SELECT COUNT(*) FROM interesting_aircraft_seen").Scan(&metrics.TotalInteresting); err != nil {
		log.Error().Err(err).Msg("Failed to count total interesting aircraft")
	}

	if err := s.pg.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM interesting_aircraft_seen WHERE DATE(seen AT TIME ZONE $1) = CURRENT_DATE", tz,
	).Scan(&metrics.TodayInteresting); err != nil {
		log.Error().Err(err).Msg("Failed to count today's interesting aircraft")
	}

	if err := s.pg.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM interesting_aircraft_seen WHERE seen >= NOW() - INTERVAL '1 hour'",
	).Scan(&metrics.HourInteresting); err != nil {
		log.Error().Err(err).Msg("Failed to count past-hour interesting aircraft")
	}

	return metrics, nil
}

func (s *StatsService) GetAboveStats(ctx context.Context, radius int) ([]AboveAircraft, error) {
	query := `
		SELECT
			ad.hex,
			ad.flight,
			ad.r,
			ad.t,
			ad.track,
			ad.first_seen,
			ad.last_seen,
			ad.last_seen_lat,
			ad.last_seen_lon,
			ad.last_seen_distance,
			ad.destination_distance,
			-- Registration data
			reg.type,
			reg.icao_type,
			reg.manufacturer,
			reg.registered_owner_country_name,
			reg.registered_owner_country_iso_name,
			reg.registered_owner_operator_flag_code,
			reg.registered_owner,
			reg.url_photo,
			reg.url_photo_thumbnail,
			-- Route data
			rt.airline_name,
			rt.airline_icao,
			rt.origin_country_name,
			rt.origin_country_iso_name,
			rt.origin_iata_code,
			rt.origin_icao_code,
			rt.origin_name,
			rt.destination_country_name,
			rt.destination_country_iso_name,
			rt.destination_iata_code,
			rt.destination_icao_code,
			rt.destination_name,
			rt.route_distance
		FROM aircraft_data ad
		LEFT JOIN registration_data reg ON ad.hex = reg.mode_s
		LEFT JOIN route_data rt ON ad.flight = rt.route_callsign
		WHERE ad.last_seen >= NOW() - INTERVAL '60 seconds'
			AND ad.last_seen_distance <= $1
		ORDER BY ad.last_seen_distance ASC
		LIMIT 5;`

	rows, err := s.pg.db.Query(ctx, query, radius)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query above stats")
		return nil, fmt.Errorf("failed to query above stats: %w", err)
	}
	defer rows.Close()

	aircraft := []AboveAircraft{}
	for rows.Next() {
		var a AboveAircraft
		err := rows.Scan(
			&a.Hex, &a.Flight, &a.Registration, &a.Type, &a.Track,
			&a.FirstSeen, &a.LastSeen, &a.LastSeenLat, &a.LastSeenLon, &a.LastSeenDistance, &a.DestinationDistance,

			&a.RegType, &a.IcaoType, &a.Manufacturer, &a.RegisteredOwnerCountryName, &a.RegisteredOwnerCountryISO,
			&a.RegisteredOwnerOperatorFlag, &a.RegisteredOwner, &a.URLPhoto, &a.URLPhotoThumbnail,

			&a.AirlineName, &a.AirlineICAO, &a.OriginCountryName, &a.OriginCountryISOName, &a.OriginIATACode,
			&a.OriginICAOCode, &a.OriginName, &a.DestinationCountryName, &a.DestinationCountryISOName,
			&a.DestinationIATACode, &a.DestinationICAOCode, &a.DestinationName, &a.RouteDistance)
		if err != nil {
			log.Error().Err(err).Msg("GetAboveStats: row scan failed")
			continue
		}
		aircraft = append(aircraft, a)
	}

	return aircraft, nil
}

func (s *StatsService) GetRecentInterestingAircraft(ctx context.Context, group InterestingGroup, limit int) ([]RecentInterestingAircraft, error) {
	query := `
		WITH latest_unique_reg AS (
			SELECT DISTINCT ON (registration) icao, registration,
			operator, type, icao_type, "group",
			category, tag1, tag2, tag3,
					hex, flight, seen, seen_epoch
			FROM interesting_aircraft_seen
			WHERE "group" = $1
			ORDER BY registration, seen DESC
		)
		SELECT *
		FROM latest_unique_reg
		ORDER BY seen DESC
		LIMIT $2`

	rows, err := s.pg.db.Query(ctx, query, string(group), limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query recent interesting aircraft")
		return nil, fmt.Errorf("failed to query recent interesting aircraft: %w", err)
	}
	defer rows.Close()

	aircraft := []RecentInterestingAircraft{}
	for rows.Next() {
		var a RecentInterestingAircraft
		err := rows.Scan(&a.Icao, &a.Registration, &a.Operator, &a.Type, &a.IcaoType,
			&a.Group, &a.Category, &a.Tag1, &a.Tag2, &a.Tag3,
			&a.Hex, &a.Flight, &a.Seen, &a.SeenEpoch)
		if err != nil {
			continue
		}
		aircraft = append(aircraft, a)
	}

	return aircraft, nil
}

func (s *StatsService) GetAircraftBySpeed(ctx context.Context, record SpeedRecord, limit int) ([]AircraftSpeedRecord, error) {
	query := fmt.Sprintf(`
		SELECT hex, flight, registration, type, first_seen, last_seen,
			   ground_speed, indicated_air_speed, true_air_speed
		FROM %s
		ORDER BY ground_speed %s
		LIMIT $1`, string(record), record.direction())

	rows, err := s.pg.db.Query(ctx, query, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query aircraft by speed")
		return nil, fmt.Errorf("failed to query aircraft by speed: %w", err)
	}
	defer rows.Close()

	aircraft := []AircraftSpeedRecord{}
	for rows.Next() {
		var a AircraftSpeedRecord
		err := rows.Scan(&a.Hex, &a.Flight, &a.Registration, &a.Type, &a.FirstSeen,
			&a.LastSeen, &a.GroundSpeed, &a.IndicatedAirSpeed, &a.TrueAirSpeed)
		if err != nil {
			continue
		}
		aircraft = append(aircraft, a)
	}

	return aircraft, nil
}

func (s *StatsService) GetAircraftByAltitude(ctx context.Context, record AltitudeRecord, limit int) ([]AircraftAltitudeRecord, error) {
	query := fmt.Sprintf(`
		SELECT hex, flight, registration, type, first_seen, last_seen,
			   barometric_altitude, geometric_altitude
		FROM %s
		ORDER BY barometric_altitude %s
		LIMIT $1`, string(record), record.direction())

	rows, err := s.pg.db.Query(ctx, query, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query aircraft by altitude")
		return nil, fmt.Errorf("failed to query aircraft by altitude: %w", err)
	}
	defer rows.Close()

	aircraft := []AircraftAltitudeRecord{}
	for rows.Next() {
		var a AircraftAltitudeRecord
		err := rows.Scan(&a.Hex, &a.Flight, &a.Registration, &a.Type, &a.FirstSeen,
			&a.LastSeen, &a.BarometricAltitude, &a.GeometricAltitude)
		if err != nil {
			continue
		}
		aircraft = append(aircraft, a)
	}

	return aircraft, nil
}

func (s *StatsService) GetTopAircraftTypes(ctx context.Context, period Period, basis CountBasis) ([]AircraftTypeCount, error) {
	var timeFilter string
	switch period {
	case PeriodYear:
		timeFilter = `age(now(), first_seen) <= INTERVAL '1 year' AND`
	case PeriodMonth:
		timeFilter = `age(now(), first_seen) <= INTERVAL '1 month' AND`
	case PeriodDay:
		timeFilter = `age(now(), first_seen) <= INTERVAL '1 day' AND`
	default:
		timeFilter = ""
	}
	innerFilter := `WHERE ` + timeFilter + ` t IS NOT NULL AND t != ''`

	var innerQuery string
	switch basis {
	case CountAircraft:
		innerQuery = `(SELECT t, hex FROM aircraft_data ` + innerFilter + `GROUP BY t, hex)`
	case CountFlights:
		innerQuery = `aircraft_data ` + innerFilter
	default:
		return nil, fmt.Errorf("unknown count basis %q", basis)
	}

	query := `SELECT
					t,
					count,
					ROUND(count * 100.0 / SUM(count) OVER(), 0) as percentage
				FROM (
					SELECT t, Count(t) as count
					FROM ` + innerQuery + `
					GROUP BY t ORDER BY count DESC
				) top_15
				ORDER BY count DESC LIMIT 15`

	rows, err := s.pg.db.Query(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query top aircraft types")
		return nil, fmt.Errorf("failed to query top aircraft types: %w", err)
	}
	defer rows.Close()

	types := []AircraftTypeCount{}
	for rows.Next() {
		var t AircraftTypeCount
		if err := rows.Scan(&t.AircraftType, &t.Count, &t.Percentage); err != nil {
			continue
		}
		types = append(types, t)
	}

	return types, nil
}

func (s *StatsService) GetTopRoutes(ctx context.Context, limit int) ([]TopRoute, error) {
	query := `
		SELECT
			CONCAT(rd.origin_iata_code, ' → ', rd.destination_iata_code) as route,
			rd.origin_iata_code,
			rd.origin_name,
			rd.destination_iata_code,
			rd.destination_name,
			COUNT(*) as flight_count
		FROM aircraft_data ad
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
			AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
			AND rd.origin_iata_code != rd.destination_iata_code
		GROUP BY rd.origin_iata_code, rd.origin_name, rd.destination_iata_code, rd.destination_name
		ORDER BY flight_count DESC
		LIMIT $1`

	rows, err := s.pg.db.Query(ctx, query, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query top routes")
		return nil, fmt.Errorf("failed to query top routes: %w", err)
	}
	defer rows.Close()

	routes := []TopRoute{}
	for rows.Next() {
		var r TopRoute
		if err := rows.Scan(&r.Route, &r.OriginIATACode, &r.OriginName, &r.DestinationIATACode, &r.DestinationName, &r.FlightCount); err != nil {
			continue
		}
		routes = append(routes, r)
	}

	return routes, nil
}

// GetTopCountries backs both the "destination" and "origin" top-countries
// views — side and its opposite are always hardcoded from route
// registration, never user input.
func (s *StatsService) GetTopCountries(ctx context.Context, side CountrySide, limit int) ([]TopCountry, error) {
	opposite := side.opposite()

	query := fmt.Sprintf(`
		SELECT
			rd.%[1]s_country_name,
			rd.%[1]s_country_iso_name,
			COUNT(*) as flight_count
		FROM aircraft_data ad
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.%[1]s_country_iso_name IS NOT NULL AND rd.%[1]s_country_iso_name != ''
			AND rd.%[2]s_country_iso_name != rd.%[1]s_country_iso_name
		GROUP BY rd.%[1]s_country_name, %[1]s_country_iso_name
		ORDER BY flight_count DESC
		LIMIT $1`, string(side), string(opposite))

	rows, err := s.pg.db.Query(ctx, query, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query top countries")
		return nil, fmt.Errorf("failed to query top countries: %w", err)
	}
	defer rows.Close()

	countries := []TopCountry{}
	for rows.Next() {
		var c TopCountry
		if err := rows.Scan(&c.CountryName, &c.CountryISO, &c.FlightCount); err != nil {
			continue
		}
		countries = append(countries, c)
	}

	return countries, nil
}

func (s *StatsService) GetTopAirlines(ctx context.Context, limit int) ([]TopAirline, error) {
	query := `
		SELECT
			rd.airline_name,
			rd.airline_icao,
			rd.airline_iata,
			COUNT(*) as flight_count
		FROM aircraft_data ad
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.airline_name IS NOT NULL AND rd.airline_name != ''
			AND rd.origin_iata_code != rd.destination_iata_code
			AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
			AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
		GROUP BY rd.airline_name, rd.airline_icao, rd.airline_iata
		ORDER BY flight_count DESC
		LIMIT $1`

	rows, err := s.pg.db.Query(ctx, query, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query top airlines")
		return nil, fmt.Errorf("failed to query top airlines: %w", err)
	}
	defer rows.Close()

	airlines := []TopAirline{}
	for rows.Next() {
		var a TopAirline
		if err := rows.Scan(&a.AirlineName, &a.AirlineICAO, &a.AirlineIATA, &a.FlightCount); err != nil {
			continue
		}
		airlines = append(airlines, a)
	}

	return airlines, nil
}

// GetTopAirports backs both the domestic and international top-airports
// views — scope is always hardcoded from route registration, never user
// input.
func (s *StatsService) GetTopAirports(ctx context.Context, scope AirportScope, limit int) ([]TopAirport, error) {
	domesticCountryISO := os.Getenv("DOMESTIC_COUNTRY_ISO")

	query := fmt.Sprintf(`
		SELECT
			airport_code,
			airport_name,
			airport_country,
			SUM(flight_count) as flight_count
		FROM (
			SELECT
				rd.origin_iata_code as airport_code,
				rd.origin_name as airport_name,
				rd.origin_country_name as airport_country,
				COUNT(*) as flight_count
			FROM aircraft_data ad
			INNER JOIN route_data rd ON ad.flight = rd.route_callsign
			WHERE rd.origin_country_iso_name %[1]s $1
				AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
				AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
				AND rd.origin_iata_code != rd.destination_iata_code
			GROUP BY rd.origin_iata_code, rd.origin_name, rd.origin_country_name
			UNION ALL
			SELECT
				rd.destination_iata_code as airport_code,
				rd.destination_name as airport_name,
				rd.destination_country_name as airport_country,
				COUNT(*) as flight_count
			FROM aircraft_data ad
			INNER JOIN route_data rd ON ad.flight = rd.route_callsign
			WHERE rd.destination_country_iso_name %[1]s $1
				AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
				AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
				AND rd.origin_iata_code != rd.destination_iata_code
			GROUP BY rd.destination_iata_code, rd.destination_name, rd.destination_country_name
		) combined_airports
		GROUP BY airport_code, airport_name, airport_country
		ORDER BY flight_count DESC
		LIMIT $2`, scope.operator())

	rows, err := s.pg.db.Query(ctx, query, domesticCountryISO, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query top airports")
		return nil, fmt.Errorf("failed to query top airports: %w", err)
	}
	defer rows.Close()

	airports := []TopAirport{}
	for rows.Next() {
		var a TopAirport
		if err := rows.Scan(&a.AirportCode, &a.AirportName, &a.AirportCountry, &a.FlightCount); err != nil {
			continue
		}
		airports = append(airports, a)
	}

	return airports, nil
}

// GetChartOverTime backs both the flights-over-time and aircraft-over-time
// chart endpoints — basis selects COUNT(*) vs COUNT(DISTINCT hex). Every
// period runs inside a transaction with the caller's tz applied via SET
// LOCAL TIME ZONE, so hourly ("day") buckets follow the requested timezone
// exactly like the monthly/yearly buckets already did.
func (s *StatsService) GetChartOverTime(ctx context.Context, tz string, period Period, basis CountBasis) (*ChartResponse, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}

	countExpr := "COUNT(*)"
	labelWord, seriesPrefix := "Flights", "flights"
	if basis == CountAircraft {
		countExpr = "COUNT(DISTINCT hex)"
		labelWord, seriesPrefix = "Aircraft", "aircraft"
	}

	var query, seriesID, label, periodUnit string
	switch period {
	case PeriodYear:
		seriesID = seriesPrefix + "_year"
		label = labelWord + " Past Year"
		periodUnit = "month"
		query = fmt.Sprintf(`WITH months AS (
				SELECT generate_series(
					DATE_TRUNC('month', CURRENT_DATE - INTERVAL '12 months'),
					DATE_TRUNC('month', CURRENT_DATE),
					'1 month'
				)::date AS month
				),
				counts AS (
				SELECT
					DATE_TRUNC('month', first_seen)::date AS month,
					%s AS count
				FROM aircraft_data
				WHERE first_seen >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '12 months')
					AND first_seen < DATE_TRUNC('month', CURRENT_DATE) + INTERVAL '1 month'
				GROUP BY 1
				)
				SELECT
				m.month::timestamptz,
				COALESCE(c.count, 0) AS count
				FROM months m
				LEFT JOIN counts c USING (month)
				ORDER BY m.month;`, countExpr)
	case PeriodMonth:
		seriesID = seriesPrefix + "_month"
		label = labelWord + " Past Month"
		periodUnit = "day"
		query = fmt.Sprintf(`WITH days AS (
				SELECT generate_series(
					CURRENT_DATE - INTERVAL '1 month',
					CURRENT_DATE,
					'1 day'
				)::date AS day
				),
				counts AS (
				SELECT
					DATE(first_seen) AS day,
					%s AS count
				FROM aircraft_data
				WHERE first_seen >= CURRENT_DATE - INTERVAL '1 month'
					AND first_seen < CURRENT_DATE + INTERVAL '1 day'
				GROUP BY 1
				)
				SELECT
					d.day::timestamptz,
					COALESCE(c.count, 0) AS count
				FROM days d
				LEFT JOIN counts c USING (day)
				ORDER BY d.day;`, countExpr)
	case PeriodDay:
		seriesID = seriesPrefix + "_day"
		label = labelWord + " Past 24 Hours"
		periodUnit = "hour"
		query = fmt.Sprintf(`WITH end_hour AS (
				SELECT date_trunc('hour', CURRENT_TIMESTAMP) AS h,
				       CURRENT_TIMESTAMP AS now
				)
				SELECT
				gs AS hour,
				COALESCE(c.count, 0) AS count
				FROM generate_series(
					(SELECT h FROM end_hour) - interval '23 hours',
					(SELECT h FROM end_hour),
					interval '1 hour'
					) AS gs
				LEFT JOIN (
				SELECT date_trunc('hour', first_seen) AS hour, %s AS count
				FROM aircraft_data, end_hour
				WHERE first_seen >= (SELECT h FROM end_hour) - interval '23 hours'
					AND first_seen <= (SELECT now FROM end_hour)
				GROUP BY 1
				) c ON c.hour = gs
				ORDER BY gs;`, countExpr)
	default:
		return nil, fmt.Errorf("unknown chart period %q", period)
	}

	tx, err := s.pg.db.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction for chart query")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, fmt.Sprintf("SET LOCAL TIME ZONE '%s'", tz)); err != nil {
		log.Error().Err(err).Msg("Failed to set local time zone for chart query")
		return nil, fmt.Errorf("failed to set local time zone: %w", err)
	}

	rows, err := tx.Query(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query chart data")
		return nil, fmt.Errorf("failed to query chart data: %w", err)
	}
	defer rows.Close()

	points := []ChartPoint{}
	for rows.Next() {
		var timeVal time.Time
		var count int
		if err := rows.Scan(&timeVal, &count); err != nil {
			continue
		}
		points = append(points, ChartPoint{X: timeVal.In(loc), Y: float64(count)})
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to commit chart query transaction")
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &ChartResponse{
		Series: []ChartSeries{{ID: seriesID, Label: label, Unit: "count", Points: points}},
		X:      ChartXAxisMeta{Type: "time", Unit: periodUnit},
		Meta:   ChartMeta{GeneratedAt: time.Now()},
	}, nil
}
