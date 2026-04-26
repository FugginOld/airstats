# Airstats

Airstats is an application to retrieve, store, and display interesting aircraft ADS-B data received via an SDR.

⚠️ Airstats is still in early development and considered "beta" so expect bugs and instability.

Airstats is a derivative of [tomcarman/skystats](https://github.com/tomcarman/skystats), originally created by [@tomcarman](https://github.com/tomcarman).

## Overview

* [Go](https://go.dev/) app with [PostgreSQL](https://www.postgresql.org/) database and [Svelte](https://svelte.dev/) + [DaisyUI](https://daisyui.com/) front end
* ADS-B data is received via [adsb-ultrafeeder](https://github.com/sdr-enthusiasts/docker-adsb-ultrafeeder) / [readsb](https://github.com/wiedehopf/readsb), running on a Raspberry Pi 4 attached to an SDR + aerial ([see it here!](docs/setup/aerial.jpg))
* The application consumes aircraft data from the readsb [aircraft.json](https://github.com/wiedehopf/readsb-githist/blob/dev/README-json.md) file
* A [gin](https://gin-gonic.com/) API surfaces information from the postgres database to the web frontend
* Registration & routing data is retrieved from the [adsb-db](https://github.com/mrjackwills/adsbdb) API
* "Interesting" aircraft are identified via a local copy of the [aircraft-taxonomy-db](https://github.com/FugginOld/aircraft-taxonomy-db)

## Features

* "Above Me" - live view of 5 nearest aircraft with routing information
* Total aircraft seen (past hour, day, all time)
* Total aircraft with route data
* Unique Countries
* Unique Airports
* Top Airlines
* Top Airports (Domestic, International)
* Top Countries (Origin, Destination)
* Top Routes
* Interesting Aircraft (Military, Government, Police, Civilian)
* Fastest Aircraft
* Slowest Aircraft
* Highest Aircraft
* Lowest Aircraft

## Setup

### Docker Compose (recommended)

This is the easiest way to run Airstats and PostgreSQL together.

1. Install Docker Engine and Docker Compose.
2. Clone this repository:
   * `git clone https://github.com/FugginOld/airstats.git`
   * `cd airstats`
3. Create your env file:
   * `cp .env.example .env`
4. Update `.env` with your values (at minimum: `READSB_AIRCRAFT_JSON`, `LAT`, `LON`, `DOMESTIC_COUNTRY_ISO`; DB values can stay as defaults for compose).
5. Start the stack:
   * `docker compose -f example.compose.yml up -d`
6. Open the UI at `http://localhost:5173` (replace `localhost` with your Docker host IP if remote).

To stop it:

* `docker compose -f example.compose.yml down`

### Docker (without Compose)

If you prefer raw Docker commands, run PostgreSQL first, then run Airstats on the same Docker network.

1. Create a network:
   * `docker network create airstats-net`
2. Start PostgreSQL:
   * `docker run -d --name airstats-db --network airstats-net -e POSTGRES_USER=user -e POSTGRES_PASSWORD=1234 -e POSTGRES_DB=airstats_db -v airstats_postgres_data:/var/lib/postgresql/data postgres:17`
3. Start Airstats:
   * `docker run -d --name airstats --network airstats-net -p 5173:80 -e READSB_AIRCRAFT_JSON=http://yourhost:yourport/data/aircraft.json -e DB_HOST=airstats-db -e DB_PORT=5432 -e DB_USER=user -e DB_PASSWORD=1234 -e DB_NAME=airstats_db -e DOMESTIC_COUNTRY_ISO=US -e LAT=0.000000 -e LON=0.000000 -e RADIUS=500 -e ABOVE_RADIUS=20 -e LOG_LEVEL=INFO ghcr.io/fugginold/airstats:latest`
4. Open `http://localhost:5173`.

### PostgreSQL setup shell script (local/non-compose setups)

If you are **not** using the compose PostgreSQL container, use the interactive PostgreSQL setup script:

* `./scripts/setup-postgres.sh`

This script can install PostgreSQL 17 (supported OS families), create/update a role and database, and write `DB_*` values to `.env`.

Full script documentation: [`docs/setup/postgres-setup-script.md`](docs/setup/postgres-setup-script.md)

Alternatively there are some [Advanced Setup](#advanced-setup) options.

### Environment Variables

| Environment Variable | Description | Example |
| --- | --- | --- |
| READSB_AIRCRAFT_JSON | URL of where readsb [aircraft.json](https://github.com/wiedehopf/readsb-githist/blob/dev/README-json.md) is being served e.g. `http://yourhost:yourport/data/aircraft.json` | `http://192.168.1.100:8080/data/aircraft.json` |
| DB_HOST | Postgres host. If running in Docker use the postgres container name. If running locally use the IP/hostname of your postgres server. Docker: `airstats-db` / Local: `192.168.1.10` | `airstats-db` |
| DB_PORT | Postgres port | `5432` |
| DB_USER | Postgres username | `user` |
| DB_PASSWORD | Postgres password | `1234` |
| DB_NAME | Postgres database name | `airstats_db` |
| DOMESTIC_COUNTRY_ISO | ISO 2-letter country code of the country your receiver is in - used to generate the "Domestic Airport" stats. | `GB` |
| LAT | Latitude of your receiver. | `XX.XXXXXX` |
| LON | Longitude of your receiver. | `YY.YYYYYY` |
| RADIUS | Distance in km from your receiver that you want to record aircraft. Set to a distance greater than that of your receiver to capture all aircraft. | `1000` |
| ABOVE_RADIUS | Radius for the "Above Timeline". Note: currently only 20km is supported. | `20` |
| LOG_LEVEL | Logging level: `TRACE`, `DEBUG`, `INFO`, `WARN`, or `ERROR`. Default is `INFO`. `DEBUG` also enables verbose Gin API logging. | `INFO` |

## Support / Feedback

Airstats is still under early active development. If you're having issues getting it running, or have suggestions/feedback, then the best place to get support is on the [#airstats](https://discord.gg/znkBr2eyev) channel in the [SDR Enthusiasts Discord](https://discord.gg/86Tyxjcd94). Alternatively you can raise an [Issue](https://github.com/FugginOld/airstats/issues) in GitHub, and I'll do my best to support.

## Advanced Setup

The intention is for Airstats to be run via the [provided Docker containers](#setup). However, if you want to run locally or if you want to contribute by developing, see below guidance.

### Running locally

* BYO postgres database (in a Docker container or other)
* Copy the contents of [`.env.example`](.env.example) into a new file called `.env`
* Populate `.env` with all required values. See [Environment Variables](#environment-variables)
* Download the latest [release binary](https://github.com/FugginOld/airstats/releases) for your OS/arch
* Execute e.g. `./airstats`
* TODO: Instructions to run the webserver

### Compile from source (e.g. to develop)

* BYO postgres database (in a Docker container or other)
* Clone this repository
* Copy the contents of `.env.example` into a new file called `.env`
* Populate `.env` with all required values. See [Environment Variables](#environment-variables)
* Change to the `core` folder e.g. `cd core`
* Compile with `go build -o airstats-daemon`
* Run the app `./airstats-daemon`
  * It can be terminated via `kill $(cat airstats.pid)`
* Run the webserver
  * Change to the /web directory e.g. `cd ../web`
  * Start the webserver with `npm run dev -- --host`
* See [`build`](/scripts/build) for a script to automate some of this

## Advanced Use Cases

### Custom aircraft-taxonomy-db csv

If you live in an area where you frequently see planes that you are not interested in, you can provide a custom version of [aircraft-taxonomy-db](https://github.com/FugginOld/aircraft-taxonomy-db).

This expects a file identical in structure to <https://github.com/FugginOld/aircraft-taxonomy-db/blob/main/data/aircraft-taxonomy-db.csv>

Add the following to the `.env` file:

```env
PLANE_DB_URL=some/custom/location/aircraft-taxonomy-db.csv
```

And the following to `compose.yml` under the `airstats` service:

```yaml
- PLANE_DB_URL=${PLANE_DB_URL}
```

⚠️ The format of the csv must match the 10-column format of aircraft-taxonomy-db (`$ICAO`, `$Registration`, `$Operator`, `$Type`, `$ICAO Type`, `#CMPG`, `$Tag 1`, `$#Tag 2`, `$#Tag 3`, `Category`).

## Screenshots

### Home

![Home](docs/screenshots/1_Home.png)

![AboveMeModal](docs/screenshots/2_AboveMeModal.png)

### Route Stats

![RouteStats](docs/screenshots/3_RouteStats.png)

### Interesting Aircraft

![InterestingSeen](docs/screenshots/4_InterestingStats.png)

![InterestingModal](docs/screenshots/5_InterestingModal.png)

### Motion Stats

![MotionStats](docs/screenshots/6_MotionStats.png)
