# plane-alert-db Image Integration Backup

This directory contains backups of the source files as they existed before the transition from
`sdr-enthusiasts/plane-alert-db` to `FugginOld/aircraft-taxonomy-db`.

The new database (`aircraft-taxonomy-db`) does not include image links, so those columns were
removed from the schema and the application code.  If you want to restore image support in the
future, these backups contain all the relevant original code.

## Files

| Backup file | Original location | What it contains |
|---|---|---|
| `db-plane-alert-data.go.bak` | `core/db-plane-alert-data.go` | Full ingestion logic including `Row.Link`, `Row.Image1-4`, image-link columns in the `INSERT` statement, and `getLatestCommitHash` pointed at `sdr-enthusiasts/plane-alert-db` |
| `models.go.bak` | `core/models.go` | `InterestingAircraft` struct with `Link`, `ImageLink1`, `ImageLink2`, `ImageLink3`, `ImageLink4` fields |
| `stats-interesting.go.bak` | `core/stats-interesting.go` | `updateInterestingSeen` SELECT and INSERT with `link`, `image_link_1`–`image_link_4` |
| `api.go.bak` | `core/api.go` | `getRecentInterestingAircraft` query/scan/response with `image_link_1`–`image_link_3` |
| `InterestingAircraft.svelte.bak` | `web/src/components/InterestingAircraft.svelte` | Aircraft modal with image carousel grid |
| `Settings.svelte.bak` | `web/src/components/Settings.svelte` | Disable-tags checkbox linked to `sdr-enthusiasts/plane-alert-db` |
| `000001_initial_schema.up.sql.bak` | `migrations/000001_initial_schema.up.sql` | Original schema including `link`, `image_link_1`–`image_link_4` in `interesting_aircraft` and `interesting_aircraft_seen` |

## Restoring

1. Copy the `.bak` files back to their original paths (drop the `.bak` suffix).
2. Run the existing down migration for `000007_remove_image_link_columns.up.sql` (that is, `migrations/000007_remove_image_link_columns.down.sql`), or roll the database back to migration 6, to restore the removed image-link columns.
3. Rebuild the backend (`go build ./...`) and frontend (`cd web && npm run build`).
