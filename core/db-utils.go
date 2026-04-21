package main

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

var allowedProcessedColumns = map[string]bool{
	"registration_processed":    true,
	"route_processed":           true,
	"interesting_processed":     true,
	"lowest_aircraft_processed": true,
	"highest_aircraft_processed": true,
	"slowest_aircraft_processed": true,
	"fastest_aircraft_processed": true,
}

var allowedTables = map[string]bool{
	"lowest_aircraft":  true,
	"highest_aircraft": true,
	"slowest_aircraft": true,
	"fastest_aircraft": true,
}

var allowedMetrics = map[string]bool{
	"barometric_altitude": true,
	"ground_speed":        true,
}

var allowedSortOrders = map[string]bool{
	"ASC":  true,
	"DESC": true,
}

func MarkProcessed(pg *postgres, colName string, aircrafts []Aircraft) {

	if !allowedProcessedColumns[colName] {
		log.Error().Msgf("MarkProcessed() - unknown column: %s", colName)
		return
	}

	batch := &pgx.Batch{}

	for _, aircraft := range aircrafts {
		updateStatement := `UPDATE aircraft_data SET ` + colName + ` = true WHERE id = $1`
		batch.Queue(updateStatement, aircraft.Id)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(aircrafts); i++ {
		_, err := br.Exec()
		if err != nil {
			log.Error().Err(err).Msg("MarkProcessed() - Unable to update data")
		}
	}
}

func DeleteExcessRows(pg *postgres, tableName string, metricName string, sortOrder string, maxRows int) {

	if !allowedTables[tableName] {
		log.Error().Msgf("DeleteExcessRows() - unknown table: %s", tableName)
		return
	}
	if !allowedMetrics[metricName] {
		log.Error().Msgf("DeleteExcessRows() - unknown metric: %s", metricName)
		return
	}
	if !allowedSortOrders[sortOrder] {
		log.Error().Msgf("DeleteExcessRows() - unknown sort order: %s", sortOrder)
		return
	}

	queryCount := `SELECT COUNT(*) FROM ` + tableName

	var rowCount int
	err := pg.db.QueryRow(context.Background(), queryCount).Scan(&rowCount)
	if err != nil {
		log.Error().Err(err).Msg("DeleteExcessRows() - Error querying db in DeleteExcessRows()")
		return
	}

	if rowCount > maxRows {

		excessRows := rowCount - maxRows

		if excessRows <= 0 {
			log.Debug().Msgf("DeleteExcessRows() - No excess rows in %s", tableName)
			return
		}

		deleteStatement := `DELETE FROM ` + tableName + `
							WHERE id IN (
								SELECT id
								FROM ` + tableName + `
								ORDER BY ` + metricName + ` ` + sortOrder + ` , first_seen ASC
								LIMIT $1
								)`

		_, err := pg.db.Exec(context.Background(), deleteStatement, excessRows)
		if err != nil {
			log.Error().Err(err).Msgf("DeleteExcessRows() - Failed to delete excess rows in %s", tableName)
		}
	}
}
