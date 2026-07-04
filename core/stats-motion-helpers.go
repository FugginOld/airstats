package main

import (
	"context"
)

// getExtremeValue runs a single-row, single-column query and returns its
// scanned value, or defaultValue if the query errors (e.g. no rows yet).
func getExtremeValue[T int | float64](pg *postgres, query string, defaultValue T) T {
	var returnValue T
	err := pg.db.QueryRow(context.Background(), query).Scan(&returnValue)
	if err != nil {
		return defaultValue
	}
	return returnValue
}

func getHighestAircraftFloor(pg *postgres) int {
	return getExtremeValue(pg, `SELECT barometric_altitude
				FROM highest_aircraft
				ORDER BY barometric_altitude ASC, first_seen ASC
				LIMIT 1`, 0)
}

func getLowestAircraftCeiling(pg *postgres) int {
	return getExtremeValue(pg, `SELECT barometric_altitude
				FROM lowest_aircraft
				ORDER BY barometric_altitude DESC, first_seen ASC
				LIMIT 1`, 999999)
}

func getFastestAircraftFloor(pg *postgres) float64 {
	return getExtremeValue(pg, `SELECT ground_speed
				FROM fastest_aircraft
				ORDER BY ground_speed ASC, first_seen ASC
				LIMIT 1`, 0.0)
}

func getSlowestAircraftCeiling(pg *postgres) float64 {
	return getExtremeValue(pg, `SELECT ground_speed
				FROM slowest_aircraft
				ORDER BY ground_speed DESC, first_seen ASC
				LIMIT 1`, 99999.0)
}
