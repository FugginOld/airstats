# PostgreSQL setup script

The repository includes an interactive setup helper at:

* `scripts/setup-postgres.sh`

## What it does

The script is intended for local/non-compose PostgreSQL setups and will:

* Detect your OS
* Install PostgreSQL 17 when missing (for supported OS families)
* Start PostgreSQL if needed
* Create or update a role and database
* Update `.env` with `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, and `DB_NAME`
* Run a database connection test

## Usage

From the repository root:

```bash
chmod +x ./scripts/setup-postgres.sh
./scripts/setup-postgres.sh
```

The wizard prompts for:

* Database host
* Database port
* Database name
* Database user
* Database password

## Supported install targets

Automatic PostgreSQL installation is handled for:

* Debian/Ubuntu/Raspbian (APT)
* Fedora/RHEL/CentOS/Rocky/AlmaLinux (DNF)
* Arch/Manjaro (pacman)
* macOS (Homebrew)

If your OS is not supported for auto-install, install PostgreSQL 17 manually and re-run the script.

## Notes

* For Docker Compose installs, you generally do **not** need this script because PostgreSQL runs in the `airstats-db` container.
* If you provide a non-local DB host, local database/role creation is skipped and only `.env` is updated.
