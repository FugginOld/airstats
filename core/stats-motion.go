package main

import (
	"context"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

func updateMeasurementStatistics(pg *postgres) {

	aircrafts := getAircraftsForMeasurementStatistics(pg)

	updateAircraftByAltitude(pg, aircrafts, LowestAircraft)
	updateAircraftBySpeed(pg, aircrafts, FastestAircraft)
	updateAircraftByAltitude(pg, aircrafts, HighestAircraft)
	updateAircraftBySpeed(pg, aircrafts, SlowestAircraft)

}

// updateAircraftByAltitude backs both the highest_aircraft and lowest_aircraft
// record tables — only the processed flag, threshold, and sort direction
// differ between the two.
func updateAircraftByAltitude(pg *postgres, aircrafts []Aircraft, record AltitudeRecord) {
	tableName := string(record)
	processedMetricName := tableName + "_processed"
	metricName := "barometric_altitude"
	descending := record.direction() == "DESC"

	var isProcessed func(Aircraft) bool
	var threshold int
	var filterNearZero bool

	switch record {
	case HighestAircraft:
		isProcessed = func(a Aircraft) bool { return a.HighestProcessed }
		threshold = getHighestAircraftFloor(pg)
		filterNearZero = false
	case LowestAircraft:
		isProcessed = func(a Aircraft) bool { return a.LowestProcessed }
		threshold = getLowestAircraftCeiling(pg)
		filterNearZero = true
	}

	var aircraftToProcess []Aircraft
	for _, aircraft := range aircrafts {
		if !isProcessed(aircraft) {
			aircraftToProcess = append(aircraftToProcess, aircraft)
		}
	}

	if len(aircraftToProcess) == 0 {
		return
	}

	sort.Slice(aircraftToProcess, func(i, j int) bool {
		if descending {
			return aircraftToProcess[i].AltBaro > aircraftToProcess[j].AltBaro
		}
		return aircraftToProcess[i].AltBaro < aircraftToProcess[j].AltBaro
	})

	var aircraftsToInsert []Aircraft
	for _, aircraft := range aircraftToProcess {
		if filterNearZero && aircraft.AltBaro < 1 {
			continue
		}
		beyondThreshold := aircraft.AltBaro < threshold
		if descending {
			beyondThreshold = aircraft.AltBaro > threshold
		}
		if beyondThreshold {
			aircraftsToInsert = append(aircraftsToInsert, aircraft)
		} else {
			break
		}
	}

	batch := &pgx.Batch{}

	for _, aircraft := range aircraftsToInsert {
		insertStatement := `
			INSERT INTO ` + tableName + ` (
				hex,
				flight,
				registration,
				type,
				first_seen,
				last_seen,
				barometric_altitude,
				geometric_altitude)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (hex, first_seen)
			DO UPDATE SET
				barometric_altitude = EXCLUDED.barometric_altitude,
				geometric_altitude = EXCLUDED.geometric_altitude,
				last_seen = EXCLUDED.last_seen`

		batch.Queue(
			insertStatement,
			aircraft.Hex,
			aircraft.Flight,
			aircraft.R,
			aircraft.T,
			aircraft.FirstSeen,
			aircraft.LastSeen,
			aircraft.AltBaro,
			aircraft.AltGeom)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(aircraftsToInsert); i++ {
		_, err := br.Exec()
		if err != nil {
			log.Error().Err(err).Msgf("updateAircraftByAltitude(%s) - Unable to insert data", tableName)
		}
	}

	DeleteExcessRows(pg, tableName, metricName, record.deleteSortOrder(), 50)

	MarkProcessed(pg, processedMetricName, aircraftToProcess)
}

// updateAircraftBySpeed backs both the fastest_aircraft and slowest_aircraft
// record tables — only the processed flag, threshold, and sort direction
// differ between the two.
func updateAircraftBySpeed(pg *postgres, aircrafts []Aircraft, record SpeedRecord) {
	tableName := string(record)
	processedMetricName := tableName + "_processed"
	metricName := "ground_speed"
	descending := record.direction() == "DESC"

	var isProcessed func(Aircraft) bool
	var threshold float64
	var filterNearZero bool

	switch record {
	case FastestAircraft:
		isProcessed = func(a Aircraft) bool { return a.FastestProcessed }
		threshold = getFastestAircraftFloor(pg)
		filterNearZero = false
	case SlowestAircraft:
		isProcessed = func(a Aircraft) bool { return a.SlowestProcessed }
		threshold = getSlowestAircraftCeiling(pg)
		filterNearZero = true
	}

	var aircraftToProcess []Aircraft
	for _, aircraft := range aircrafts {
		if !isProcessed(aircraft) {
			aircraftToProcess = append(aircraftToProcess, aircraft)
		}
	}

	if len(aircraftToProcess) == 0 {
		return
	}

	sort.Slice(aircraftToProcess, func(i, j int) bool {
		if descending {
			return aircraftToProcess[i].Gs > aircraftToProcess[j].Gs
		}
		return aircraftToProcess[i].Gs < aircraftToProcess[j].Gs
	})

	var aircraftsToInsert []Aircraft
	for _, aircraft := range aircraftToProcess {
		if filterNearZero && aircraft.Gs < 1 {
			continue
		}
		beyondThreshold := aircraft.Gs < threshold
		if descending {
			beyondThreshold = aircraft.Gs > threshold
		}
		if beyondThreshold {
			aircraftsToInsert = append(aircraftsToInsert, aircraft)
		} else {
			break
		}
	}

	batch := &pgx.Batch{}

	for _, aircraft := range aircraftsToInsert {
		insertStatement := `
			INSERT INTO ` + tableName + ` (
				hex,
				flight,
				registration,
				type,
				first_seen,
				last_seen,
				ground_speed,
				indicated_air_speed,
				true_air_speed)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (hex, first_seen)
			DO UPDATE SET
				ground_speed = EXCLUDED.ground_speed,
				indicated_air_speed = EXCLUDED.indicated_air_speed,
				true_air_speed = EXCLUDED.true_air_speed,
				last_seen = EXCLUDED.last_seen`

		batch.Queue(
			insertStatement,
			aircraft.Hex,
			aircraft.Flight,
			aircraft.R,
			aircraft.T,
			aircraft.FirstSeen,
			aircraft.LastSeen,
			aircraft.Gs,
			aircraft.Ias,
			aircraft.Tas)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(aircraftsToInsert); i++ {
		_, err := br.Exec()
		if err != nil {
			log.Error().Err(err).Msgf("updateAircraftBySpeed(%s) - Unable to insert data", tableName)
		}
	}

	DeleteExcessRows(pg, tableName, metricName, record.deleteSortOrder(), 50)

	MarkProcessed(pg, processedMetricName, aircraftToProcess)
}

func getAircraftsForMeasurementStatistics(pg *postgres) []Aircraft {

	query := `SELECT id, hex, flight, r, t, first_seen, last_seen, alt_baro, alt_geom, gs, ias, tas,
				lowest_aircraft_processed, highest_aircraft_processed, fastest_aircraft_processed, slowest_aircraft_processed
				FROM aircraft_data
				WHERE lowest_aircraft_processed = false OR
					highest_aircraft_processed = false OR
					fastest_aircraft_processed = false OR
					slowest_aircraft_processed = false`

	rows, err := pg.db.Query(context.Background(), query)
	if err != nil {
		log.Error().Err(err).Msg("getAircraftsForMeasurementStatistics() - Error querying db")
		return nil
	}
	defer rows.Close()

	var aircrafts []Aircraft

	for rows.Next() {

		var aircraft Aircraft

		err := rows.Scan(
			&aircraft.Id,
			&aircraft.Hex,
			&aircraft.Flight,
			&aircraft.R,
			&aircraft.T,
			&aircraft.FirstSeen,
			&aircraft.LastSeen,
			&aircraft.AltBaro,
			&aircraft.AltGeom,
			&aircraft.Gs,
			&aircraft.Ias,
			&aircraft.Tas,
			&aircraft.LowestProcessed,
			&aircraft.HighestProcessed,
			&aircraft.FastestProcessed,
			&aircraft.SlowestProcessed)

		if err != nil {
			log.Error().Err(err).Msg("getAircraftsForMeasurementStatistics() - Error scanning rows")
			return nil
		}
		aircrafts = append(aircrafts, aircraft)
	}

	log.Debug().Msgf("Aircrafts that have not have statistics processed: %d", len(aircrafts))
	return aircrafts
}
