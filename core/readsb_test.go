package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	t.Setenv("READSB_AIRCRAFT_JSON", server.URL)
	data, err := Fetch()
	if err != nil {
		t.Fatalf("Fetch() unexpected error: %v", err)
	}
	if string(data) != `{"ok":true}` {
		t.Fatalf("Fetch() = %q, want %q", string(data), `{"ok":true}`)
	}
}

func TestFetchNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	t.Setenv("READSB_AIRCRAFT_JSON", server.URL)
	if _, err := Fetch(); err == nil {
		t.Fatal("Fetch() expected error for non-200 response")
	}
}

func TestFetchCSVData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("col1,col2\na,b\n"))
	}))
	defer server.Close()

	records, err := fetchCSVData(server.URL)
	if err != nil {
		t.Fatalf("fetchCSVData() unexpected error: %v", err)
	}
	if len(records) != 2 || len(records[0]) != 2 || records[1][1] != "b" {
		t.Fatalf("fetchCSVData() returned unexpected records: %+v", records)
	}
}

func TestFetchCSVDataRequestError(t *testing.T) {
	originalClient := httpClient
	httpClient = &http.Client{Timeout: 20 * time.Millisecond}
	defer func() { httpClient = originalClient }()

	if _, err := fetchCSVData("http://127.0.0.1:1"); err == nil {
		t.Fatal("fetchCSVData() expected request error")
	}
}
