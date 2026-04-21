package main

import "testing"

func TestTrimFlightStrings(t *testing.T) {
	response := Response{
		Aircraft: []Aircraft{
			{Flight: " BAW123   "},
			{Flight: "\tAFR456\n"},
			{Flight: ""},
		},
	}

	response.TrimFlightStrings()

	if response.Aircraft[0].Flight != "BAW123" {
		t.Fatalf("first flight = %q, want %q", response.Aircraft[0].Flight, "BAW123")
	}
	if response.Aircraft[1].Flight != "AFR456" {
		t.Fatalf("second flight = %q, want %q", response.Aircraft[1].Flight, "AFR456")
	}
	if response.Aircraft[2].Flight != "" {
		t.Fatalf("third flight = %q, want empty", response.Aircraft[2].Flight)
	}
}
