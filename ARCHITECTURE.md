# Architecture

This document describes how airstats is put together: the moving pieces, how
data flows between them, and the conventions the codebase follows. For
domain vocabulary (Aircraft, Route, Interesting Aircraft, Motion record,
Settings, Service) see [CONTEXT.md](CONTEXT.md). For user-facing setup and
deployment, see [README.md](README.md).

## System overview

```
                     ┌─────────────────────┐
  readsb/ultrafeeder  │  Go backend (core/)  │
  aircraft.json  ───▶ │                      │ ───▶  Postgres
  (2s poll)           │  ingestion tickers   │       (aircraft_data,
                       │  + gin API server    │◀───   route_data, ...)
                       └──────────┬───────────┘
                                  │ JSON over /api/*
                                  ▼
                       ┌──────────────────────┐
                       │  Svelte frontend      │
                       │  (web/), polling the  │
                       │  API on its own timer │
                       └──────────────────────┘

  External services the backend calls out to:
    - api.adsbdb.com            (registration lookup)
    - adsb.im/api/0/routeset    (route/callsign lookup)
    - GitHub (aircraft-taxonomy-db CSV) (interesting-aircraft reference data)
```

Everything is a single Go binary (`core/`, package `main`) plus a static
Svelte/Vite frontend (`web/`) that the Go binary also serves. There's no
separate backend-for-frontend or message queue — the whole pipeline is
poll-based: the backend polls readsb, the frontend polls the backend.

## Backend (`core/`)

### Package layout

| File | Responsibility |
|---|---|
| `core.go` | Entrypoint. Connects to Postgres, runs migrations, starts the API server goroutine, then runs the ingestion ticker loop forever. |
| `api.go` | HTTP layer only — gin route registration and thin handlers that call `StatsService`/`SettingsService` and map results/errors to JSON. |
| `stats-service.go` | `StatsService` — the repository seam for all read-side `/api/stats/*` queries. Owns the SQL; handlers never build SQL themselves. |
| `settings.go` | `SettingsService` — same pattern as `StatsService`, for the `user_settings` table. This is the older of the two and the pattern `StatsService` was built to match. |
| `db-connector.go` | `postgres` struct wrapping a `*pgxpool.Pool`; `NewPG`/`GetConnectionUrl`. |
| `db-interface.go` | The `DB` interface — see [The DB interface](#the-db-interface-a-testing-seam-not-a-database-seam) below. |
| `db-migrations.go` | Runs golang-migrate migrations through the same pool as everything else (via `stdlib.OpenDBFromPool`). |
| `db-utils.go` | `DeleteExcessRows`/`MarkProcessed` — small shared helpers used by the record-holder update functions. |
| `aircraft.go` | `updateAircraftDatabase` — polls readsb's `aircraft.json`, upserts sightings into `aircraft_data`. |
| `routes.go` | `updateRoutes` — looks up route/callsign info from adsb.im for unprocessed flights, writes `route_data`. |
| `registrations.go` | `updateRegistrations` — looks up registration info from adsbdb for unprocessed aircraft, writes `registration_data`. |
| `stats-interesting.go` | `updateInterestingSeen` — matches sightings against the `interesting_aircraft` reference table, writes `interesting_aircraft_seen`. |
| `stats-motion.go` / `stats-motion-helpers.go` | Computes the four "record holder" tables (fastest/slowest/highest/lowest) from unprocessed sightings. |
| `db-plane-alert-data.go` | `UpsertAircraftTaxonomyDb` — pulls the `aircraft-taxonomy-db` CSV from GitHub on startup (or a custom `PLANE_DB_URL`) to (re)populate the `interesting_aircraft` reference table. |
| `countries.go` | ISO country code → name lookup used when rendering route/country stats. |
| `models.go` | All shared structs: the raw `Aircraft` readsb model, `StatsService` result types, chart types, `RegistrationInfo`/`RouteInfo` external API response shapes. |
| `settings.go` | See above. |
| `version.go` | Build-time version/commit/date, set via ldflags. |

### The ingestion pipeline

`core.go`'s main loop is a `select` over five tickers, each independent:

| Ticker | Interval | Does |
|---|---|---|
| `updateAircraftDatabase` | 2s | Poll readsb, upsert `aircraft_data` |
| `updateMeasurementStatistics` | 120s | Recompute fastest/slowest/highest/lowest tables |
| `updateRegistrations` | 30s | Enrich unprocessed aircraft with registration data (adsbdb) |
| `updateRoutes` | 300s | Enrich unprocessed flights with route data (adsb.im) |
| `updateInterestingSeen` | 120s | Match sightings against the interesting-aircraft reference list |

Each of these follows the same shape: query `aircraft_data` (or another table)
for rows with a `*_processed = false` flag, do the work, batch-write results,
then call `MarkProcessed` to flip the flag. This is why `aircraft_data` has
four separate `*_processed` boolean columns (`registration_processed`,
`route_processed`, `interesting_processed`, plus the four motion-record
processed flags) — each pipeline stage tracks its own progress over the same
row set independently.

None of this ticker-driven code is exposed through `StatsService` — it's
deliberately out of that seam's scope (see below).

### The repository seam: `StatsService`

`api.go`'s ~15 read-side query handlers used to each build their own SQL,
scan rows manually, and shape JSON inline. `StatsService` is the seam that
was extracted to fix that: handlers now do little more than call a
`StatsService` method and map `(T, error)` to an HTTP response.

```go
func (s *APIServer) getTopRoutes(c *gin.Context) {
	limit := s.getLimit("route_table_limit")
	routes, err := s.stats.GetTopRoutes(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, routes)
}
```

Conventions `StatsService` follows, all deliberate:

- **`(T, error)` only** — `StatsService` never imports `gin`/`net/http`. It
  doesn't know it's being called over HTTP.
- **Typed control parameters, not raw strings.** Where a handler used to pass
  a literal like `"DESC"` or `"="` into a `fmt.Sprintf`-built query, the
  method instead takes a small string-backed type (`SpeedRecord`,
  `AltitudeRecord`, `CountrySide`, `AirportScope`, `CountBasis`, `Period`,
  `InterestingGroup`) with the valid values as constants. A typo can't
  silently reach the SQL.
- **Two error philosophies, deliberately.** Most methods are fail-fast:
  any query error returns immediately. The four "metrics" methods
  (`GetFlightsSeenMetrics`, `GetAircraftSeenMetrics`, `GetRouteMetrics`,
  `GetInterestingMetrics`) are best-effort instead — each runs 3 independent
  counts and a failed one just zero-values that field rather than failing
  the whole call. This matches pre-existing, tested handler behavior; don't
  "fix" it into fail-fast without checking `api_test.go` first.
- **Result structs live in `models.go`**, not next to `StatsService` — this
  matches where `ChartPoint`/`UserSetting` already lived before
  `StatsService` existed.

Scope: `StatsService` only covers `api.go`'s read-side handlers. The
write-side ingestion files (`aircraft.go`, `stats-motion.go`,
`registrations.go`, etc.) are a different concern — batch/ticker-driven
writes, not HTTP query serving — and were deliberately left alone when
`StatsService` was built.

### The `DB` interface: a testing seam, not a database seam

```go
type DB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Ping(ctx context.Context) error
	Close()
}
```

`DB` has exactly one production adapter (`*pgxpool.Pool`, which already
implements this interface) and one test adapter (`mockDB` in
`mock_test.go`, a queue-based mock — tests pre-load the rows/errors each
call should return). It does **not** abstract over multiple database
backends; it exists purely so tests can run without a live Postgres. Don't
read "we support swapping databases" into this seam — that was never the
intent, and code that treats it that way is drifting from what it's for.

`postgres` (in `db-connector.go`) wraps `DB` for everyday queries, but also
keeps a `pool *pgxpool.Pool` field with the concrete type, because a few
things (migrations, via `stdlib.OpenDBFromPool`) need the real pgx pool
object rather than the abstracted interface. Both fields point at the same
pool — there is only ever one connection pool to Postgres in this process.

### Testing

Every DB-touching file has `mockDB`-based tests except the three
network-calling functions (`getRegistration`, `getLatestCommitHash`,
`fetchCSVData`) — those need HTTP client mocking, a different concern, and
are currently untested. `StatsService` has its own `stats_service_test.go`
hitting methods directly (no gin/httptest involved); `api_test.go` is thin
wiring tests that confirm a route calls the right method and maps errors to
the right status code.

## Frontend (`web/`)

Svelte 5 (legacy `export let`/`<slot>` style, not runes), Vite, DaisyUI +
Tailwind for styling, Chart.js for the two time-series charts.

### Layout

```
web/src/
  App.svelte              top-level shell: nav, "Above" timeline, 4 tabs
  components/
    Tab*.svelte           one per tab (Activity, Route Information,
                           Interesting Aircraft, Record Holders)
    Metric*.svelte        small stat-card widgets (flights/aircraft/routes/
                           interesting seen)
    Motion*.svelte         thin <MotionStats endpoint=... /> callers, one
                           per record type (fastest/slowest/highest/lowest)
    MotionStats.svelte     generic table+columns component for record
                           holders
    RouteTopList.svelte    generic list component (endpoint/title props +
                           a slot for the row template)
    InterestingAircraft.svelte  table+modal for one interesting-aircraft
                                 group (props: endpoint/title/icon/type)
    AboveTimeline.svelte    the "5 nearest aircraft" timeline widget
    charts/                Chart.js wrappers (aircraft/flights over time)
  lib/
    pollResource.js         shared polling primitive — see below
    themeStore.js
  stores/
    settings.js              settings store + 3 refresh-trigger stores
```

### The shared data-source primitive: `pollResource.js`

Every component that polls an API endpoint (`Metric*`, `InterestingAircraft`,
`AboveTimeline`, and `RouteTopList`/`MotionStats` internally) uses one
primitive instead of hand-rolling `fetch`/`onMount`/`onDestroy`/`setInterval`:

```js
export function createPolledResource(endpoint, { refreshMs = 60000, refreshTrigger } = {}) {
  return writable({ data: null, loading: true, error: null }, (set, update) => {
    // fetch on mount, setInterval(refreshMs), optional refreshTrigger subscription
    // returns a cleanup function
  });
}
```

It's built on Svelte's native store `start`/stop lifecycle (the second
argument to `writable`): polling begins the moment something subscribes
(i.e. a component renders `$resource`) and stops when the last subscriber
goes away. This is why none of the migrated components need `onMount`/
`onDestroy` at all — the store lifecycle **is** the mount/destroy hook.

Usage is always the same three lines:

```js
const resource = createPolledResource('api/stats/routes/routes', { refreshMs: 60000, refreshTrigger: refreshRouteData });
$: data = $resource.data || [];
$: loading = $resource.loading;
$: error = $resource.error;
```

`refreshTrigger` is optional and per-component — it's one of the three
counters in `stores/settings.js` (`refreshRouteData`,
`refreshInterestingData`, `refreshRecordHolderData`), incremented when the
corresponding table-limit setting is saved. Components that don't care
about settings changes (the `Metric*` cards, `AboveTimeline`) simply omit
it.

### Two "deep" list/table components

`RouteTopList.svelte` and `MotionStats.svelte` both generalize "fetch a
list, render loading/error/empty states, render each row" — but differently,
because their row shapes differ:

- `RouteTopList` takes `endpoint`/`title` and a **slot** for the row markup
  (`let:row`) — used when the caller needs arbitrary markup per row
  (airline logos, flags, route arrows).
- `MotionStats` takes `endpoint`/`title`/**`columns`** (a config array of
  `{header, field, formatter, class}`) — used when rows are uniform enough
  to describe declaratively, rendered generically without a slot.

Don't merge these two — the review that led to their current shape
explicitly considered and rejected a single mega-generic component (see
`CONTEXT.md` for the reasoning: mixed render shapes don't unify cleanly).

## Database

Postgres, migrated via [golang-migrate](https://github.com/golang-migrate/migrate)
(`migrations/`, applied automatically on startup — see `db-migrations.go`).

| Table | Written by | Read by |
|---|---|---|
| `aircraft_data` | `updateAircraftDatabase` (ticker) | almost everything — the core sightings table |
| `registration_data` | `updateRegistrations` (ticker) | `GetAboveStats` (LEFT JOIN) |
| `route_data` | `updateRoutes` (ticker) | `GetAboveStats`, `GetTopRoutes`, `GetTopAirlines`, `GetTopCountries`, `GetTopAirports` |
| `fastest_aircraft` / `slowest_aircraft` | `updateAircraftBySpeed` (ticker) | `GetAircraftBySpeed` |
| `highest_aircraft` / `lowest_aircraft` | `updateAircraftByAltitude` (ticker) | `GetAircraftByAltitude` |
| `interesting_aircraft` | `UpsertAircraftTaxonomyDb` (startup, from GitHub CSV) | `updateInterestingSeen` (matching reference) |
| `interesting_aircraft_seen` | `updateInterestingSeen` (ticker) | `GetRecentInterestingAircraft`, `GetInterestingMetrics` |
| `user_settings` | `SettingsService.UpdateSetting(s)` | `SettingsService.GetSetting(s)`, `getLimit` |
| `schema_migrations` | golang-migrate | golang-migrate |

## Conventions worth knowing before you extend this codebase

- **Naming a new `*Service`?** Follow `SettingsService`/`StatsService`'s
  shape exactly: `NewXService(pg *postgres) *XService`, methods return
  `(T, error)`, no HTTP imports, structs in `models.go`.
- **Adding a query parameter that's really an enum** (a fixed set of
  literal values threaded into SQL)? Use a typed string const, not a raw
  `string` — see `SpeedRecord`/`CountBasis`/etc. in `stats-service.go`.
- **Adding a new frontend component that polls an endpoint?** Use
  `createPolledResource`, don't hand-roll `onMount`/`setInterval` again.
- **The `DB` interface is for tests, not for supporting multiple
  databases.** If you need a genuinely different query for a genuinely
  different data source, that's a new seam, not a new `DB` implementation.
