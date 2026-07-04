package main

import (
	"database/sql"
	"strings"
	"time"
)

type Response struct {
	Now      float64    `json:"now"`
	Messages int        `json:"messages"`
	Aircraft []Aircraft `json:"aircraft"`
}
type Aircraft struct {
	Id                  int
	Hex                 string  `json:"hex"`
	Type                string  `json:"type"`
	Flight              string  `json:"flight"`
	R                   string  `json:"r"`
	T                   string  `json:"t"`
	AltBaro             int     `json:"alt_baro"`
	AltGeom             int     `json:"alt_geom"`
	Gs                  float64 `json:"gs"`
	Ias                 int     `json:"ias"`
	Tas                 int     `json:"tas"`
	Track               float64 `json:"track"`
	BaroRate            int     `json:"baro_rate"`
	NavQnh              float64 `json:"nav_qnh"`
	NavAltitudeMcp      int     `json:"nav_altitude_mcp"`
	NavHeading          float64 `json:"nav_heading"`
	Lat                 float64 `json:"lat"`
	Lon                 float64 `json:"lon"`
	Nic                 int     `json:"nic"`
	Rc                  int     `json:"rc"`
	SeenPos             float64 `json:"seen_pos"`
	RDst                float64 `json:"r_dst"`
	RDir                float64 `json:"r_dir"`
	Version             int     `json:"version"`
	NicBaro             int     `json:"nic_baro"`
	NacP                int     `json:"nac_p"`
	NacV                int     `json:"nac_v"`
	Sil                 int     `json:"sil"`
	SilType             string  `json:"sil_type"`
	Alert               int     `json:"alert"`
	Spi                 int     `json:"spi"`
	Mlat                []any   `json:"mlat"`
	Tisb                []any   `json:"tisb"`
	Messages            int     `json:"messages"`
	Seen                float64 `json:"seen"`
	Rssi                int     `json:"rssi"`
	DbFlags             int     `json:"dbFlags"`
	Squawk              string  `json:"squawk"`
	Category            string  `json:"category"`
	FirstSeen           time.Time
	FirstSeenEpoch      float64
	LastSeen            time.Time
	LastSeenEpoch       float64
	LastSeenLat         sql.NullFloat64
	LastSeenLon         sql.NullFloat64
	LastSeenDistance    sql.NullFloat64
	DestinationDistance sql.NullFloat64
	LowestProcessed     bool
	HighestProcessed    bool
	FastestProcessed    bool
	SlowestProcessed    bool
}

type InterestingAircraft struct {
	Icao         string
	Registration sql.NullString
	Operator     sql.NullString
	Type         sql.NullString
	IcaoType     sql.NullString
	Group        sql.NullString
	Tag1         sql.NullString
	Tag2         sql.NullString
	Tag3         sql.NullString
	Category     sql.NullString
	Hex          string
	Flight       string
	R            string
	T            string
	AltBaro      int
	AltGeom      int
	Gs           float64
	Ias          int
	Tas          int
	Track        float64
	BaroRate     int
	Lat          float64
	Lon          float64
	Alert        int
	DbFlags      int
	Seen         time.Time
	SeenEpoch    float64
}

// Flight string sometimes has trailing whitespace
func (r *Response) TrimFlightStrings() {
	for i := range r.Aircraft {
		r.Aircraft[i].Flight = strings.TrimSpace(r.Aircraft[i].Flight)
	}
}

type RegistrationInfo struct {
	Response struct {
		Aircraft struct {
			Type                            string `json:"type"`
			IcaoType                        string `json:"icao_type"`
			Manufacturer                    string `json:"manufacturer"`
			ModeS                           string `json:"mode_s"`
			Registration                    string `json:"registration"`
			RegisteredOwnerCountryIsoName   string `json:"registered_owner_country_iso_name"`
			RegisteredOwnerCountryName      string `json:"registered_owner_country_name"`
			RegisteredOwnerOperatorFlagCode string `json:"registered_owner_operator_flag_code"`
			RegisteredOwner                 string `json:"registered_owner"`
			URLPhoto                        any    `json:"url_photo"`
			URLPhotoThumbnail               any    `json:"url_photo_thumbnail"`
		} `json:"aircraft"`
	} `json:"response"`
}

type RouteInfo struct {
	AirportCodesIata string `json:"_airport_codes_iata"`
	Airports         []struct {
		AltFeet     float64 `json:"alt_feet"`
		AltMeters   float64 `json:"alt_meters"`
		CountryIso2 string  `json:"countryiso2"`
		Iata        string  `json:"iata"`
		Icao        string  `json:"icao"`
		Lat         float64 `json:"lat"`
		Location    string  `json:"location"`
		Lon         float64 `json:"lon"`
		Name        string  `json:"name"`
	} `json:"_airports"`
	AirlineCode  string `json:"airline_code"`
	AirportCodes string `json:"airport_codes"`
	Callsign     string `json:"callsign"`
	Number       string `json:"number"`
	Plausible    bool   `json:"plausible"`
}

type RouteAPIPlane struct {
	Callsign string  `json:"callsign"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

type RouteAPIRequest struct {
	Planes []RouteAPIPlane `json:"planes"`
}

type ChartPoint struct {
	X time.Time `json:"x"`
	Y float64   `json:"y"`
}

type ChartSeries struct {
	ID     string       `json:"id"`
	Label  string       `json:"label"`
	Unit   string       `json:"unit,omitempty"`
	Points []ChartPoint `json:"points"`
}

type ChartXAxisMeta struct {
	Type     string `json:"type"`
	Timezone string `json:"timezone,omitempty"`
	Unit     string `json:"unit,omitempty"`
}

type ChartMeta struct {
	GeneratedAt time.Time `json:"generated_at"`
}

type ChartResponse struct {
	Series []ChartSeries  `json:"series"`
	X      ChartXAxisMeta `json:"x"`
	Meta   ChartMeta      `json:"meta"`
}

// --- StatsService result types ---

type FlightsSeenMetrics struct {
	TotalFlights int `json:"total_flights"`
	TodayFlights int `json:"today_flights"`
	HourFlights  int `json:"hour_flights"`
}

type AircraftSeenMetrics struct {
	TotalAircraft int `json:"total_aircraft"`
	TodayAircraft int `json:"today_aircraft"`
	HourAircraft  int `json:"hour_aircraft"`
}

type RouteMetrics struct {
	TotalRoutes int `json:"total_routes"`
	// json tag typo preserved: web/src/components/MetricRoutes.svelte reads "unqiue_countries" verbatim.
	UniqueCountries int `json:"unqiue_countries"`
	UniqueAirports  int `json:"unique_airports"`
}

type InterestingMetrics struct {
	TotalInteresting int `json:"total_interesting"`
	TodayInteresting int `json:"today_interesting"`
	HourInteresting  int `json:"hour_interesting"`
}

type AboveAircraft struct {
	Hex                  string     `json:"hex"`
	Flight               string     `json:"flight"`
	Registration         string     `json:"registration"`
	Type                 string     `json:"type"`
	FirstSeen            *time.Time `json:"first_seen"`
	LastSeen             *time.Time `json:"last_seen"`
	LastSeenLat          float64    `json:"last_seen_lat"`
	LastSeenLon          float64    `json:"last_seen_lon"`
	LastSeenDistance     float64    `json:"last_seen_distance"`
	DestinationDistance  *float64   `json:"destination_distance"`
	Track                float64    `json:"track"`

	RegType                     *string `json:"reg_type"`
	IcaoType                    *string `json:"icao_type"`
	Manufacturer                *string `json:"manufacturer"`
	RegisteredOwnerCountryName  *string `json:"registered_owner_country_name"`
	RegisteredOwnerCountryISO   *string `json:"registered_owner_country_iso"`
	RegisteredOwnerOperatorFlag *string `json:"registered_owner_operator_flag"`
	RegisteredOwner             *string `json:"registered_owner"`
	URLPhoto                    *string `json:"url_photo"`
	URLPhotoThumbnail           *string `json:"url_photo_thumbnail"`

	AirlineName                *string  `json:"airline_name"`
	AirlineICAO                *string  `json:"airline_icao"`
	OriginCountryName          *string  `json:"origin_country_name"`
	OriginCountryISOName       *string  `json:"origin_country_iso_name"`
	OriginIATACode             *string  `json:"origin_iata_code"`
	OriginICAOCode             *string  `json:"origin_icao_code"`
	OriginName                 *string  `json:"origin_name"`
	DestinationCountryName     *string  `json:"destination_country_name"`
	DestinationCountryISOName  *string  `json:"destination_country_iso_name"`
	DestinationIATACode        *string  `json:"destination_iata_code"`
	DestinationICAOCode        *string  `json:"destination_icao_code"`
	DestinationName            *string  `json:"destination_name"`
	RouteDistance              *float64 `json:"route_distance"`
}

type RecentInterestingAircraft struct {
	Icao         string     `json:"icao"`
	Registration string     `json:"registration"`
	Operator     string     `json:"operator"`
	Type         string     `json:"type"`
	IcaoType     string     `json:"icao_type"`
	Group        string     `json:"group"`
	Category     string     `json:"category"`
	Tag1         string     `json:"tag1"`
	Tag2         string     `json:"tag2"`
	Tag3         string     `json:"tag3"`
	Hex          string     `json:"hex"`
	Flight       string     `json:"flight"`
	Seen         *time.Time `json:"seen"`
	SeenEpoch    float64    `json:"seen_epoch"`
}

type AircraftSpeedRecord struct {
	Hex               string     `json:"hex"`
	Flight            string     `json:"flight"`
	Registration      string     `json:"registration"`
	Type              string     `json:"type"`
	FirstSeen         *time.Time `json:"first_seen"`
	LastSeen          *time.Time `json:"last_seen"`
	GroundSpeed       float64    `json:"ground_speed"`
	IndicatedAirSpeed int        `json:"indicated_air_speed"`
	TrueAirSpeed      int        `json:"true_air_speed"`
}

type AircraftAltitudeRecord struct {
	Hex                string     `json:"hex"`
	Flight             string     `json:"flight"`
	Registration       string     `json:"registration"`
	Type               string     `json:"type"`
	FirstSeen          *time.Time `json:"first_seen"`
	LastSeen           *time.Time `json:"last_seen"`
	BarometricAltitude int        `json:"barometric_altitude"`
	GeometricAltitude  int        `json:"geometric_altitude"`
}

type AircraftTypeCount struct {
	AircraftType string  `json:"aircraft_type"`
	Count        int     `json:"count"`
	Percentage   float64 `json:"percentage"`
}

type TopRoute struct {
	Route               string `json:"route"`
	OriginIATACode      string `json:"origin_iata_code"`
	OriginName          string `json:"origin_name"`
	DestinationIATACode string `json:"destination_iata_code"`
	DestinationName     string `json:"destination_name"`
	FlightCount         int    `json:"flight_count"`
}

type TopCountry struct {
	CountryName string `json:"country_name"`
	CountryISO  string `json:"country_iso"`
	FlightCount int    `json:"flight_count"`
}

type TopAirline struct {
	AirlineName string `json:"airline_name"`
	AirlineICAO string `json:"airline_icao"`
	AirlineIATA string `json:"airline_iata"`
	FlightCount int    `json:"flight_count"`
}

type TopAirport struct {
	AirportCode    string `json:"airport_code"`
	AirportName    string `json:"airport_name"`
	AirportCountry string `json:"airport_country"`
	FlightCount    int    `json:"flight_count"`
}
