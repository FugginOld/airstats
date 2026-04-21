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

if mapped["icao"] != 0 || mapped["registration"] != 1 || mapped["type"] != 2 {
t.Fatalf("unexpected header map: %+v", mapped)
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
