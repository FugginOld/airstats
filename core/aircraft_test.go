package main

import (
"math"
"testing"
)

func TestIsNonAircraft(t *testing.T) {
tests := []struct {
name     string
aircraft Aircraft
want     bool
}{
{name: "tower type", aircraft: Aircraft{T: "TWR"}, want: true},
{name: "tower registration", aircraft: Aircraft{R: "TWR"}, want: true},
{name: "category C prefix", aircraft: Aircraft{Category: "C3"}, want: true},
{name: "special squawk", aircraft: Aircraft{Squawk: "7777"}, want: true},
{name: "normal aircraft", aircraft: Aircraft{T: "A320", Category: "A1", Squawk: "1234"}, want: false},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
if got := isNonAircraft(tc.aircraft); got != tc.want {
t.Fatalf("isNonAircraft() = %v, want %v", got, tc.want)
}
})
}
}

func TestGetLatLonRadiusFromEnv(t *testing.T) {
t.Setenv("LAT", "51.5074")
t.Setenv("LON", "-0.1278")
t.Setenv("RADIUS", "250.5")

if got := getLat(); got != 51.5074 {
t.Fatalf("getLat() = %v, want 51.5074", got)
}
if got := getLon(); got != -0.1278 {
t.Fatalf("getLon() = %v, want -0.1278", got)
}
if got := getRadius(); got != 250.5 {
t.Fatalf("getRadius() = %v, want 250.5", got)
}
}

func TestGetLatLonRadiusInvalidEnvReturnsZero(t *testing.T) {
t.Setenv("LAT", "not-a-number")
t.Setenv("LON", "not-a-number")
t.Setenv("RADIUS", "not-a-number")

if got := getLat(); got != 0 {
t.Fatalf("getLat() = %v, want 0", got)
}
if got := getLon(); got != 0 {
t.Fatalf("getLon() = %v, want 0", got)
}
if got := getRadius(); got != 0 {
t.Fatalf("getRadius() = %v, want 0", got)
}
}

func TestGetDistance(t *testing.T) {
t.Setenv("LAT", "51.5074")
t.Setenv("LON", "-0.1278")

distance := getDistance([]float64{-0.1278, 51.5074})
if distance == nil {
t.Fatal("getDistance() returned nil")
}
if *distance > 0.01 {
t.Fatalf("getDistance() = %v, want near 0", *distance)
}
}

func TestGetDestinationDistanceRoundsToTwoDecimals(t *testing.T) {
distance := getDestinationDistance(51.5074, -0.1278, 48.8566, 2.3522)

if distance <= 300 || distance >= 400 {
t.Fatalf("getDestinationDistance() = %v, expected realistic London-Paris range", distance)
}

rounded := math.Round(distance*100) / 100
if distance != rounded {
t.Fatalf("getDestinationDistance() = %v, want rounded to two decimals", distance)
}
}
