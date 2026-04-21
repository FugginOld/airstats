package main

import "testing"

func TestCountryIsoToNameGetName(t *testing.T) {
	lookup := CountryIsoToName()

	if name, ok := lookup.GetName("GB"); !ok || name != "United Kingdom" {
		t.Fatalf("GetName(GB) = (%q, %v), want (United Kingdom, true)", name, ok)
	}
	if _, ok := lookup.GetName("ZZ"); ok {
		t.Fatal("GetName(ZZ) expected not found")
	}
}

func TestGetConnectionUrl(t *testing.T) {
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_HOST", "dbhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "skystats_db")

	got := GetConnectionUrl()
	want := "postgres://user:pass@dbhost:5432/skystats_db"
	if got != want {
		t.Fatalf("GetConnectionUrl() = %q, want %q", got, want)
	}
}

func TestGetHeaderMap(t *testing.T) {
	headers := []string{"icao", "registration", "type"}
	mapped := getHeaderMap(headers)

	if len(mapped) != len(headers) {
		t.Fatalf("getHeaderMap(%v) returned %d entries, want %d: %+v", headers, len(mapped), len(headers), mapped)
	}
	if idx, ok := mapped["icao"]; !ok || idx != 0 {
		t.Fatalf("mapped[icao] = (%d, %v), want (0, true); full map: %+v", idx, ok, mapped)
	}
	if idx, ok := mapped["registration"]; !ok || idx != 1 {
		t.Fatalf("mapped[registration] = (%d, %v), want (1, true); full map: %+v", idx, ok, mapped)
	}
	if idx, ok := mapped["type"]; !ok || idx != 2 {
		t.Fatalf("mapped[type] = (%d, %v), want (2, true); full map: %+v", idx, ok, mapped)
	}
}

func TestGetValue(t *testing.T) {
	if got := getValue(""); got != nil {
		t.Fatal("getValue(\"\") should return nil")
	}
	if got := getValue("abc"); got == nil || *got != "abc" {
		t.Fatalf("getValue(\"abc\") = %v, want pointer to abc", got)
	}
}
