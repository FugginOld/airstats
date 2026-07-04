# CONTEXT.md

<!-- Repo-specific context for AI agents. -->

## Repo Purpose

airstats is an ADS-B flight-tracking stats service. A Go backend (`core/`)
polls a local `readsb` feed for currently-visible aircraft, enriches sightings
with registration and route lookups, persists them to Postgres, and serves
aggregated stats (routes, motion records, interesting aircraft, aircraft type
breakdowns) over a JSON API. A Svelte 5 frontend (`web/`) renders that API as
a live dashboard.

## Domain Vocabulary

| Term | Meaning |
|------|---------|
| Aircraft | A single ADS-B-tracked airframe sighting, decoded from the readsb feed (`Aircraft` in `models.go`). Identified by `hex` (ICAO24 address). |
| Flight | One aircraft sighting session — first-seen to last-seen — stored in `aircraft_data`. |
| Route | The origin/destination pairing inferred for a flight's callsign, looked up via the external adsb.im routeset API (`routes.go`, `RouteInfo`/`RouteAPIRequest` in `models.go`). |
| Interesting Aircraft | An aircraft matching a watchlist (civilian/police/military/government groups) tracked separately from ordinary sightings (`InterestingAircraft` in `models.go`, `stats-interesting.go`). |
| Motion record | A "record holder" aircraft — fastest/slowest/highest/lowest seen — recomputed on a ticker via floor/ceiling thresholds (`stats-motion.go`, `stats-motion-helpers.go`). |
| Settings | User-configurable key/value pairs (e.g. table row limits) read at request time (`SettingsService`, `UserSetting`). |
| Service | The established convention for a `*postgres`-backed module owning one concern's SQL behind named methods returning `(T, error)` — e.g. `SettingsService`. `StatsService` (in progress) follows the same shape for the `/api/stats` read-side handlers. |
| `DB` interface | A testing seam, not a database seam — one production adapter (`*pgxpool.Pool`) and one test adapter (`mockDB` in `mock_test.go`); it doesn't abstract over multiple databases. |

## Important Directories

```text
core/        - Go backend (package main): gin HTTP API, readsb polling, stats aggregation, Postgres access
web/         - Svelte 5 frontend (Vite build)
migrations/  - SQL migrations, run via golang-migrate (pgx/v5 driver)
scripts/     - build/utility scripts
rootfs/      - container filesystem overlay (Home Assistant add-on packaging)
data/        - runtime data directory
docs/        - project docs, screenshots, setup guides
```

## Build & Test Commands

```bash
# Backend (core/)
cd core && go build ./...
cd core && go vet ./...
cd core && go test ./...

# Frontend (web/)
cd web && npm install
cd web && npm run dev      # local dev server
cd web && npm run build    # production build (vite build)
```

## Compatibility Rules

- Go 1.25.3 (see `go.mod`); Postgres access exclusively via `jackc/pgx/v5` (no `lib/pq` — removed, see git history).
- Svelte 5, using the legacy `export let` props / `<slot>` convention throughout — not runes (`$props()`, `{#snippet}`) yet.
- Dev environment is Windows; watch for path-separator assumptions in scripts.

## Known Risks

- `core/api.go`'s ~15 read-side query handlers build SQL and scan rows inline with no repository seam — being addressed via a `StatsService` extraction (in progress).
- `core/db-migrations.go` opens its own `database/sql` connection, bypassing the `DB`/`postgres` seam every other file goes through.
- Several ticker-driven batch-update files (`stats-motion.go`, `registrations.go`, `stats-interesting.go`, `db-plane-alert-data.go`, `db-utils.go`) have no tests, despite the `mockDB` seam needed to test them already existing and being used elsewhere (`aircraft.go`, `api.go`, `settings.go`).
- Frontend metric/chart components (e.g. `MetricAircraftSeen`, `MetricRoutes`, `charts/AircraftByPeriod`) each reimplement fetch/loading/error/poll state, despite `MotionStats.svelte`/`RouteTopList.svelte` already proving a shared pattern.

## Related Repos / Services

- `readsb` — external ADS-B decoder; airstats polls its aircraft JSON feed at `READSB_AIRCRAFT_JSON`.
- `api.adsbdb.com` — external registration lookup (`registrations.go`).
- `adsb.im` routeset API — external route/callsign lookup (`routes.go`).
