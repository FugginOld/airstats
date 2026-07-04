package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type APIServer struct {
	pg       *postgres
	port     string
	settings *SettingsService
	stats    *StatsService
}

func NewAPIServer(pg *postgres) *APIServer {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	return &APIServer{
		pg:       pg,
		port:     port,
		settings: NewSettingsService(pg),
		stats:    NewStatsService(pg),
	}
}

func (s *APIServer) getTimezone(c *gin.Context) string {

	tz := c.Query("tz")
	if tz == "" {
		return "UTC"
	}

	_, err := time.LoadLocation(tz)
	if err != nil {
		return "UTC"
	}

	return tz
}

func (s *APIServer) Start() {

	// Start gin API server in release or debug mode based on LOG_LEVEL
	logLevel := os.Getenv("LOG_LEVEL")
	var r *gin.Engine
	if logLevel == "DEBUG" || logLevel == "TRACE" {
		r = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
		r.Use(gin.Recovery())
	}

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	})

	// API routes
	api := r.Group("api")
	{
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
			stats.GET("/types/flights/year", func(c *gin.Context) { s.getTopAircraftTypes(c, PeriodYear, CountFlights) })
			stats.GET("/types/flights/month", func(c *gin.Context) { s.getTopAircraftTypes(c, PeriodMonth, CountFlights) })
			stats.GET("/types/flights/day", func(c *gin.Context) { s.getTopAircraftTypes(c, PeriodDay, CountFlights) })

			stats.GET("/types/aircraft/all", func(c *gin.Context) { s.getTopAircraftTypes(c, PeriodAll, CountAircraft) })
			stats.GET("/types/aircraft/year", func(c *gin.Context) { s.getTopAircraftTypes(c, PeriodYear, CountAircraft) })
			stats.GET("/types/aircraft/month", func(c *gin.Context) { s.getTopAircraftTypes(c, PeriodMonth, CountAircraft) })
			stats.GET("/types/aircraft/day", func(c *gin.Context) { s.getTopAircraftTypes(c, PeriodDay, CountAircraft) })

			stats.GET("/charts/flights/year", func(c *gin.Context) { s.getChartOverTime(c, PeriodYear, CountFlights) })
			stats.GET("/charts/flights/month", func(c *gin.Context) { s.getChartOverTime(c, PeriodMonth, CountFlights) })
			stats.GET("/charts/flights/day", func(c *gin.Context) { s.getChartOverTime(c, PeriodDay, CountFlights) })

			stats.GET("/charts/aircraft/year", func(c *gin.Context) { s.getChartOverTime(c, PeriodYear, CountAircraft) })
			stats.GET("/charts/aircraft/month", func(c *gin.Context) { s.getChartOverTime(c, PeriodMonth, CountAircraft) })
			stats.GET("/charts/aircraft/day", func(c *gin.Context) { s.getChartOverTime(c, PeriodDay, CountAircraft) })

		}

		settings := api.Group("/settings")
		{
			settings.GET("", s.getSettings)
			settings.PUT("", s.updateSettings)
		}

		api.GET("/version", s.getVersion)
	}

	// Serve static files
	r.Static("/static", "../web")
	r.StaticFile("/", "../web/index.html")

	r.Run("0.0.0.0:" + s.port)

}

func (s *APIServer) getFlightsSeenMetrics(c *gin.Context) {
	metrics, err := s.stats.GetFlightsSeenMetrics(c.Request.Context(), s.getTimezone(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (s *APIServer) getAircraftSeenMetrics(c *gin.Context) {
	metrics, err := s.stats.GetAircraftSeenMetrics(c.Request.Context(), s.getTimezone(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (s *APIServer) getRouteMetrics(c *gin.Context) {
	metrics, err := s.stats.GetRouteMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (s *APIServer) getInterestingMetrics(c *gin.Context) {
	metrics, err := s.stats.GetInterestingMetrics(c.Request.Context(), s.getTimezone(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (s *APIServer) getAboveStats(c *gin.Context) {

	radiusValue := os.Getenv("ABOVE_RADIUS")
	radius, err := strconv.Atoi(radiusValue)
	if err != nil || radius <= 0 {
		log.Error().Err(err).Msg("Error parsing ABOVE_RADIUS environment variable")
		return
	}

	aircraft, err := s.stats.GetAboveStats(c.Request.Context(), radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) getRecentInterestingAircraft(c *gin.Context, group InterestingGroup) {
	limit := s.getLimit("interesting_table_limit")

	aircraft, err := s.stats.GetRecentInterestingAircraft(c.Request.Context(), group, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) getAircraftBySpeed(c *gin.Context, record SpeedRecord) {
	limit := s.getLimit("record_holder_table_limit")

	aircraft, err := s.stats.GetAircraftBySpeed(c.Request.Context(), record, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) getAircraftByAltitude(c *gin.Context, record AltitudeRecord) {
	limit := s.getLimit("record_holder_table_limit")

	aircraft, err := s.stats.GetAircraftByAltitude(c.Request.Context(), record, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) getTopAircraftTypes(c *gin.Context, period Period, basis CountBasis) {
	types, err := s.stats.GetTopAircraftTypes(c.Request.Context(), period, basis)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types)
}

func (s *APIServer) getTopRoutes(c *gin.Context) {
	limit := s.getLimit("route_table_limit")

	routes, err := s.stats.GetTopRoutes(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, routes)
}

func (s *APIServer) getTopCountries(c *gin.Context, side CountrySide) {
	limit := s.getLimit("route_table_limit")

	countries, err := s.stats.GetTopCountries(c.Request.Context(), side, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, countries)
}

func (s *APIServer) getTopAirlines(c *gin.Context) {
	limit := s.getLimit("route_table_limit")

	airlines, err := s.stats.GetTopAirlines(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, airlines)
}

func (s *APIServer) getTopAirports(c *gin.Context, scope AirportScope) {
	limit := s.getLimit("route_table_limit")

	airports, err := s.stats.GetTopAirports(c.Request.Context(), scope, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, airports)
}

func (s *APIServer) getChartOverTime(c *gin.Context, period Period, basis CountBasis) {
	chart, err := s.stats.GetChartOverTime(c.Request.Context(), s.getTimezone(c), period, basis)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chart)
}

func (s *APIServer) getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": version,
		"commit":  commit,
		"date":    date,
	})
}

func (s *APIServer) getLimit(settingKey string) int {

	if settingKey != "" {
		setting, err := s.settings.GetSetting(settingKey)
		if err == nil {
			limit, err := strconv.Atoi(setting.SettingValue)
			if err == nil {
				return limit
			}
		}
	}

	// Default if no setting
	return 5

}

func (s *APIServer) getSettings(c *gin.Context) {
	settings, err := s.settings.GetAllSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (s *APIServer) updateSettings(c *gin.Context) {
	var updates map[string]string
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := s.settings.UpdateSettings(updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated settings
	settings, err := s.settings.GetAllSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}
