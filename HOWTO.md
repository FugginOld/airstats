# HOWTO: developing airstats

This is a developer workflow guide — running the app locally for
development, testing changes, and common recipes. For architecture, see
[ARCHITECTURE.md](ARCHITECTURE.md). For deployment/production setup
(Docker Compose, environment variables), see [README.md](README.md).

## Prerequisites

- Go 1.25+ (matches `go.mod`)
- Node 22+ and npm (matches CI; Vite 6 requires Node ^18/^20/>=22)
- Docker, for a disposable local Postgres (or any Postgres 17 you already have)

## Local dev loop

This is the exact recipe used to verify changes against a real database —
not a toy setup.

### 1. Start Postgres

```bash
docker run -d --name airstats-dev-db \
  -e POSTGRES_USER=airstats -e POSTGRES_PASSWORD=airstats -e POSTGRES_DB=airstats \
  -p 55432:5432 postgres:17
```

Use a non-default host port (`55432` here) if you might already have
Postgres running locally.

### 2. Run the backend

From `core/`:

```bash
DOCKER_ENV=true \
DB_HOST=localhost DB_PORT=55432 DB_USER=airstats DB_PASSWORD=airstats DB_NAME=airstats \
API_PORT=8080 LOG_LEVEL=INFO DOMESTIC_COUNTRY_ISO=US ABOVE_RADIUS=50 \
READSB_AIRCRAFT_JSON=http://localhost:9/aircraft.json \
go run .
```

**`DOCKER_ENV=true` matters even outside Docker.** Without it, `core.go`
re-execs itself as a background daemon (`go-daemon`) and logs go to
`core/airstats.log` instead of your terminal — annoying for a dev loop.
Setting `DOCKER_ENV=true` keeps it in the foreground with logs on stdout,
exactly like the production container does.

You don't need a real `READSB_AIRCRAFT_JSON` to develop — pointing it at
something unreachable is fine. `updateAircraftDatabase` logs a
`connection refused` error every 2 seconds and otherwise no-ops; it doesn't
crash the app. Migrations run automatically on startup
(`RunDatabaseMigrations`, via the same connection pool `NewPG` opened).

### 3. Run the frontend

From `web/`:

```bash
npm install   # first time only
npm run dev -- --port 5183
```

Vite's dev server proxies `/api` to `http://localhost:8080` (see
`web/vite.config.js`) — no CORS config, no separate API URL to set. Open
`http://localhost:5183`.

### 4. Seed some data (optional, for visually checking the UI)

A fresh database is empty, so every card/tab renders its correct-but-boring
empty/zero state. To see populated tables and charts without a real ADS-B
feed, insert rows directly:

```bash
docker exec -i airstats-dev-db psql -U airstats -d airstats <<'EOF'
INSERT INTO aircraft_data (hex, flight, first_seen, last_seen, type, r, t, alt_baro, alt_geom, gs, ias, tas, track, last_seen_lat, last_seen_lon, last_seen_distance)
VALUES ('a10001','UAL123', NOW() - INTERVAL '5 minutes', NOW(), 'B738','N100UA','B738', 35000, 35100, 450, 440, 460, 90, 40.63, -73.78, 3);
EOF
```

Notes on what to seed for which view (see `ARCHITECTURE.md`'s table for
which table feeds what):
- `aircraft_data` rows with `last_seen` in the last 60s and
  `last_seen_distance <= ABOVE_RADIUS` → the "Above" timeline (this window
  is intentionally short — seeded rows age out within a minute).
- `route_data` rows with `route_callsign` matching an `aircraft_data.flight`
  → Route Information tab.
- `fastest_aircraft`/`slowest_aircraft`/`highest_aircraft`/`lowest_aircraft`
  rows directly → Record Holders tab, without waiting for the 120s ticker.
- `interesting_aircraft_seen` rows (with `"group"` one of `Civ`/`Gov`/
  `Mil`/`Pol`) → Interesting Aircraft tab.

This is throwaway data for a throwaway database — there's no seed script
checked into the repo, and there doesn't need to be.

### 5. Tear down

```bash
docker rm -f airstats-dev-db
```

## Running tests

```bash
cd core && go build ./... && go vet ./... && go test ./...
cd web && npm run build   # frontend has no test suite; a clean build is the check
```

This is exactly what [`.github/workflows/ci.yml`](.github/workflows/ci.yml)
runs on push/PR to `main`.

## Common tasks

### Add a new `/api/stats/*` endpoint

1. Add the query method to `StatsService` in `core/stats-service.go` —
   `(ctx context.Context, ...) (T, error)`, no `gin` import. Follow the
   existing methods for the logging/error-wrapping convention.
2. Add the result struct to `core/models.go` with `json` tags matching
   whatever the frontend expects.
3. Add a thin handler in `core/api.go` that calls the service method and
   maps the result/error to a response (copy an existing handler — they're
   all 3-6 lines).
4. Register the route in `APIServer.Start()` (and mirror it in
   `api_test.go`'s `newRouter` if you're adding wiring tests).
5. Add a `stats_service_test.go` test hitting the method directly via
   `mockDB` — see existing tests for the `newTestPG(db)` pattern.

### Add a new database migration

```bash
# migrations/NNNNNN_description.up.sql and .down.sql, sequential numbering
```

Migrations run automatically on startup via golang-migrate — no manual
step needed beyond adding the files. Test locally by starting against a
fresh Postgres (step 1 above) and checking the "Successfully migrated
database to version: N" log line.

### Add a new frontend component that polls an endpoint

Don't hand-roll `fetch`/`onMount`/`onDestroy`/`setInterval`. Use the shared
primitive:

```js
import { createPolledResource } from '../lib/pollResource.js';

const resource = createPolledResource('api/stats/whatever', { refreshMs: 60000 });
$: data = $resource.data;
$: loading = $resource.loading;
$: error = $resource.error;
```

Add `refreshTrigger: someStore` only if the new view should refresh
immediately when a specific setting changes (see `stores/settings.js` for
the existing trigger stores).

## Troubleshooting

- **`UpsertAircraftTaxonomyDb` fails / app exits on startup.** This
  fetches a CSV from GitHub on every boot to populate the interesting-
  aircraft reference table. If your environment has no internet access to
  `raw.githubusercontent.com`/`api.github.com`, this will fail and the app
  will exit — that's existing, not-yet-hardened behavior, not a bug in your
  setup.
- **Logs going to a file instead of your terminal.** You forgot
  `DOCKER_ENV=true` — see step 2 above.
- **Port already in use** (Vite or the Go API). Something from a previous
  run is still bound — check for a leftover `go run`/`vite` process, or
  just pick a different port.
- **Broken airline logo images in Route Information → Top Airlines.** The
  `<img>` src points at a third-party CloudFront CDN
  (`doj0yisjozhv1.cloudfront.net/square-logos/{icao}.png`) that may not be
  reachable from your network. This degrades gracefully (alt text shows
  instead) and isn't something introduced by any recent change — it's
  pre-existing.
