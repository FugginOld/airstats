package main

import (
	"database/sql"
	"testing"
)

func TestBuildRouteApiRequestBody(t *testing.T) {
	input := []Aircraft{
		{
			Flight:      "BAW123",
			LastSeenLat: sql.NullFloat64{Float64: 51.5074, Valid: true},
			LastSeenLon: sql.NullFloat64{Float64: -0.1278, Valid: true},
		},
		{
			Flight:      "",
			LastSeenLat: sql.NullFloat64{Float64: 51.5, Valid: true},
			LastSeenLon: sql.NullFloat64{Float64: -0.12, Valid: true},
		},
		{
			Flight:      "AFR456",
			LastSeenLat: sql.NullFloat64{Float64: 48.8566, Valid: false},
			LastSeenLon: sql.NullFloat64{Float64: 2.3522, Valid: true},
		},
	}

	result := buildRouteApiRequestBody(input)

	if len(result.Planes) != 1 {
		t.Fatalf("len(result.Planes) = %d, want 1", len(result.Planes))
	}

	plane := result.Planes[0]
	if plane.Callsign != "BAW123" || plane.Lat != 51.5074 || plane.Lng != -0.1278 {
		t.Fatalf("unexpected plane payload: %+v", plane)
	}
}

func TestGetDistanceBetweenAirports(t *testing.T) {
	t.Setenv("LAT", "51.5074")

	distance := getDistanceBetweenAirports([]float64{-0.1278, 51.5074}, []float64{2.3522, 48.8566})
	if distance == nil {
		t.Fatal("getDistanceBetweenAirports() returned nil")
	}

	if *distance <= 300 || *distance >= 400 {
		t.Fatalf("getDistanceBetweenAirports() = %v, expected London-Paris range", *distance)
	}
}
