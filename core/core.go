package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sevlyar/go-daemon"
)

func main() {

	checkFlags()

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true})
	log.Info().Msg("hello world")

	// Load .env file
	if err := godotenv.Load("../.env"); err != nil {
		if err := godotenv.Load(); err != nil {
			log.Warn().Msg("No .env file found, using environment variables")
		}
	}

	// LogLevel = strings.ToUpper(os.Getenv("LOG_LEVEL"))
	// if LogLevel == "" {
	// 	LogLevel = "INFO"
	// }

	// If running outside of docker, run as a daemon
	if os.Getenv("DOCKER_ENV") != "true" {
		execPath, _ := os.Executable()
		execDir := filepath.Dir(execPath)

		cntxt := &daemon.Context{
			PidFileName: filepath.Join(execDir, "skystats.pid"),
			PidFilePerm: 0644,
			LogFileName: filepath.Join(execDir, "skystats.log"),
			LogFilePerm: 0640,
			WorkDir:     execDir,
			Umask:       027,
		}

		d, err := cntxt.Reborn()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to launch daemon")
		}
		if d != nil {
			return
		}
		defer cntxt.Release()

		log.Info().Msg("Skystats: Running in daemon mode")
	}

	// Welcome to skystats
	if banner, err := os.ReadFile("../docs/logo/skystats_ascii.txt"); err == nil {
		log.Print("\n" + string(banner))
	}

	url := GetConnectionUrl()

	log.Info().Msg("Connecting to Postgres database")

	pg, err := NewPG(context.Background(), url)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Postgres database")
		os.Exit(1)
	}

	// Setup db
	log.Info().Msg("Checking to see if any database initialisation / migrations are needed")
	if err := RunDatabaseMigrations(); err != nil {
		log.Printf("Error initialising or migrating the database: %v", err)
		os.Exit(1)
	}

	log.Info().Msg("Updating database with plane-alert-db data")
	if err := UpsertPlaneAlertDb(pg); err != nil {
		log.Error().Msgf("Error updating interesting aircraft data: %v", err)
		os.Exit(1)
	}

	// Start API server in a separate goroutine
	log.Info().Msg("Starting API server")
	go func() {
		apiServer := NewAPIServer(pg)
		apiServer.Start()
	}()

	updateAircraftDataTicker := time.NewTicker(2 * time.Second)
	updateStatisticsTicker := time.NewTicker(120 * time.Second)
	updateRegistrationsTicker := time.NewTicker(30 * time.Second)
	updateRoutesTicker := time.NewTicker(300 * time.Second)
	updateInterestingSeenTicker := time.NewTicker(120 * time.Second)

	defer func() {
		log.Info().Msg("Closing database connection")
		updateAircraftDataTicker.Stop()
		updateStatisticsTicker.Stop()
		updateRegistrationsTicker.Stop()
		updateRoutesTicker.Stop()
		updateInterestingSeenTicker.Stop()
		pg.Close()
	}()

	for {
		select {
		case <-updateAircraftDataTicker.C:
			log.Debug().Msgf("Update Aircraft: %s", time.Now().Format("2006-01-02 15:04:05"))
			updateAircraftDatabase(pg)
		case <-updateStatisticsTicker.C:
			log.Debug().Msgf("Update Statistics: %s", time.Now().Format("2006-01-02 15:04:05"))
			updateMeasurementStatistics(pg)
		case <-updateRegistrationsTicker.C:
			log.Debug().Msgf("Update Registrations: %s", time.Now().Format("2006-01-02 15:04:05"))
			updateRegistrations(pg)
		case <-updateRoutesTicker.C:
			log.Debug().Msgf("Update Routes: %s", time.Now().Format("2006-01-02 15:04:05"))
			updateRoutes(pg)
		case <-updateInterestingSeenTicker.C:
			log.Debug().Msgf("Update Interesting Seen: %s", time.Now().Format("2006-01-02 15:04:05"))
			updateInterestingSeen(pg)
		}
	}

}

func checkFlags() {
	flag.Parse()
	if showVersion {
		showVersionExit()
	}
}
