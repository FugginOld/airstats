#!/usr/bin/env bash
# PostgreSQL 17 setup wizard for Airstats
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$PROJECT_ROOT/.env"
ENV_EXAMPLE="$PROJECT_ROOT/.env.example"
PG_VERSION=17

# ── Colours ───────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
info()    { printf "${BLUE}▶${NC} %s\n" "$*"; }
success() { printf "${GREEN}✔${NC} %s\n" "$*"; }
warn()    { printf "${YELLOW}⚠${NC} %s\n" "$*"; }
die()     { printf "${RED}✖${NC} %s\n" "$*" >&2; exit 1; }

# ── TUI (whiptail with plain-prompt fallback) ─────────────────────────────────
USE_TUI=false
command -v whiptail &>/dev/null && USE_TUI=true
WT="Airstats – PostgreSQL $PG_VERSION Setup"

tui_input() {      # tui_input "prompt" "default"
    if $USE_TUI; then
        whiptail --title "$WT" --inputbox "$1" 10 65 "$2" 3>&1 1>&2 2>&3
    else
        local v; read -rp "$1 [$2]: " v; echo "${v:-$2}"
    fi
}

tui_password() {   # tui_password "prompt"
    if $USE_TUI; then
        whiptail --title "$WT" --passwordbox "$1" 10 65 3>&1 1>&2 2>&3
    else
        local v; read -rsp "$1: " v; echo; echo "$v"
    fi
}

tui_yesno() {      # tui_yesno "prompt"  →  0=yes 1=no
    if $USE_TUI; then
        whiptail --title "$WT" --yesno "$1" 14 65
    else
        local yn; read -rp "$1 [y/N]: " yn; [[ "${yn,,}" == "y" ]]
    fi
}

tui_msg() {        # tui_msg "text"
    if $USE_TUI; then
        whiptail --title "$WT" --msgbox "$1" 18 70
    else
        echo -e "$1"; read -rp $'\nPress Enter to continue...'
    fi
}

# ── OS detection ──────────────────────────────────────────────────────────────
OS_ID=""; OS_ID_LIKE=""
detect_os() {
    if [[ "$(uname)" == "Darwin" ]]; then
        OS_ID="macos"
    elif [[ -f /etc/os-release ]]; then
        # shellcheck disable=SC1091
        . /etc/os-release
        OS_ID="${ID:-unknown}"
        OS_ID_LIKE="${ID_LIKE:-}"
    else
        OS_ID="unknown"
    fi
}

# ── PostgreSQL install ────────────────────────────────────────────────────────
pg_installed() {
    command -v psql &>/dev/null || return 1
    local v; v=$(psql --version 2>/dev/null | grep -oE '[0-9]+' | head -1)
    [[ "${v:-0}" -ge $PG_VERSION ]]
}

install_pg_debian() {
    info "Adding PostgreSQL APT repository..."
    sudo apt-get install -y -qq curl ca-certificates lsb-release
    sudo install -d /usr/share/postgresql-common/pgdg
    sudo curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc \
        -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc
    echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] \
https://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" \
        | sudo tee /etc/apt/sources.list.d/pgdg.list >/dev/null
    sudo apt-get update -qq
    sudo apt-get install -y postgresql-$PG_VERSION
}

install_pg_rhel() {
    info "Adding PostgreSQL DNF repository..."
    local arch; arch=$(uname -m)
    local el; el=$(rpm -E %{rhel} 2>/dev/null || echo 9)
    sudo dnf install -y -q \
        "https://download.postgresql.org/pub/repos/yum/reporpms/EL-${el}-${arch}/pgdg-redhat-repo-latest.noarch.rpm" \
        2>/dev/null || true
    sudo dnf -qy module disable postgresql 2>/dev/null || true
    sudo dnf install -y postgresql${PG_VERSION}-server
    sudo /usr/pgsql-${PG_VERSION}/bin/postgresql-${PG_VERSION}-setup initdb
    sudo systemctl enable --now postgresql-${PG_VERSION}
}

install_pg_arch() {
    info "Installing PostgreSQL via pacman..."
    sudo pacman -Sy --noconfirm postgresql
    sudo -u postgres initdb --locale=C.UTF-8 -D /var/lib/postgres/data 2>/dev/null || true
    sudo systemctl enable --now postgresql
}

install_pg_macos() {
    command -v brew &>/dev/null || die "Homebrew not found. Install it from https://brew.sh then re-run."
    info "Installing PostgreSQL $PG_VERSION via Homebrew..."
    brew install postgresql@$PG_VERSION
    brew services start postgresql@$PG_VERSION
    local prefix; prefix=$(brew --prefix)
    export PATH="${prefix}/opt/postgresql@${PG_VERSION}/bin:$PATH"
}

install_postgres() {
    local id="$OS_ID" like="$OS_ID_LIKE"
    if   [[ "$id" == "macos" ]]; then
        install_pg_macos
    elif [[ "$id" =~ ^(ubuntu|debian|raspbian)$ || "$like" == *debian* ]]; then
        install_pg_debian
    elif [[ "$id" =~ ^(fedora|rhel|centos|rocky|almalinux)$ || "$like" == *rhel* || "$like" == *fedora* ]]; then
        install_pg_rhel
    elif [[ "$id" =~ ^(arch|manjaro)$ ]]; then
        install_pg_arch
    else
        die "Unsupported OS '$id'. Install PostgreSQL $PG_VERSION manually then re-run."
    fi
    success "PostgreSQL $PG_VERSION installed."
}

# ── PostgreSQL service ────────────────────────────────────────────────────────
ensure_running() {
    if pg_isready -q 2>/dev/null; then
        success "PostgreSQL is running."
        return
    fi
    warn "PostgreSQL is not running — attempting to start..."
    sudo systemctl start postgresql 2>/dev/null \
        || sudo systemctl start postgresql-$PG_VERSION 2>/dev/null \
        || die "Cannot start PostgreSQL. Start it manually and re-run."
    sleep 2
    pg_isready -q || die "PostgreSQL still not accepting connections."
    success "PostgreSQL is running."
}

# ── Database and role creation ────────────────────────────────────────────────
create_db_objects() {
    local user="$1" pass="$2" name="$3"

    info "Creating role '$user'..."
    sudo -u postgres psql -v ON_ERROR_STOP=1 -q <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${user}') THEN
    CREATE ROLE "${user}" WITH LOGIN PASSWORD '${pass}';
  ELSE
    ALTER ROLE "${user}" WITH LOGIN PASSWORD '${pass}';
  END IF;
END
\$\$;
SQL

    info "Creating database '$name'..."
    sudo -u postgres psql -v ON_ERROR_STOP=1 -q <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = '${name}') THEN
    CREATE DATABASE "${name}" OWNER "${user}";
  ELSE
    ALTER DATABASE "${name}" OWNER TO "${user}";
  END IF;
END
\$\$;
SQL

    sudo -u postgres psql -v ON_ERROR_STOP=1 -q -d "${name}" <<SQL
GRANT ALL PRIVILEGES ON DATABASE "${name}" TO "${user}";
GRANT ALL ON SCHEMA public TO "${user}";
SQL

    success "Role and database ready."
}

# ── .env writing ──────────────────────────────────────────────────────────────
set_env_var() {    # set_env_var KEY VALUE FILE
    local key="$1" val="$2" file="$3"
    if grep -q "^${key}=" "$file" 2>/dev/null; then
        if [[ "$(uname)" == "Darwin" ]]; then
            sed -i '' "s|^${key}=.*|${key}=${val}|" "$file"
        else
            sed -i "s|^${key}=.*|${key}=${val}|" "$file"
        fi
    else
        printf '%s=%s\n' "$key" "$val" >> "$file"
    fi
}

write_env() {
    local host="$1" port="$2" user="$3" pass="$4" name="$5"
    if [[ ! -f "$ENV_FILE" ]]; then
        if [[ -f "$ENV_EXAMPLE" ]]; then
            cp "$ENV_EXAMPLE" "$ENV_FILE"
            info "Created .env from .env.example"
        else
            touch "$ENV_FILE"
        fi
    fi
    set_env_var DB_HOST     "$host" "$ENV_FILE"
    set_env_var DB_PORT     "$port" "$ENV_FILE"
    set_env_var DB_USER     "$user" "$ENV_FILE"
    set_env_var DB_PASSWORD "$pass" "$ENV_FILE"
    set_env_var DB_NAME     "$name" "$ENV_FILE"
    success ".env updated with database connection settings."
}

# ── Main ──────────────────────────────────────────────────────────────────────
main() {
    detect_os

    tui_msg "Welcome to the Airstats PostgreSQL $PG_VERSION setup wizard.

This script will:
  • Install PostgreSQL $PG_VERSION if not already present
  • Create a dedicated database role and database
  • Write the connection settings to your .env file

Airstats will apply schema migrations automatically on first start."

    # Collect connection settings
    local host port name user pass pass2
    host=$(tui_input "Database host" "localhost")
    port=$(tui_input "Database port" "5432")
    name=$(tui_input "Database name" "airstats_db")
    user=$(tui_input "Database user" "airstats")

    while true; do
        pass=$(tui_password  "Database password")
        pass2=$(tui_password "Confirm password")
        if [[ "$pass" == "$pass2" ]]; then
            break
        fi
        tui_msg "Passwords do not match — please try again."
    done

    [[ -z "$pass" ]] && die "Password cannot be empty."

    # Mask password for display
    local stars="${pass//?/*}"

    if ! tui_yesno "Review your settings:

  Host     :  $host
  Port     :  $port
  Database :  $name
  User     :  $user
  Password :  $stars

Write these settings and proceed?"; then
        echo "Aborted."
        exit 0
    fi

    # Install PostgreSQL if needed
    if pg_installed; then
        success "PostgreSQL $PG_VERSION (or newer) is already installed."
    else
        if tui_yesno "PostgreSQL $PG_VERSION was not found. Install it now?"; then
            install_postgres
        else
            warn "Skipping installation. Will attempt database setup anyway."
        fi
    fi

    # Create role + database for localhost installs
    if [[ "$host" == "localhost" || "$host" == "127.0.0.1" ]]; then
        ensure_running
        create_db_objects "$user" "$pass" "$name"
    else
        warn "Remote host specified — skipping local database creation."
        warn "Ensure the role '$user' and database '$name' exist on '$host:$port'."
    fi

    # Write .env
    write_env "$host" "$port" "$user" "$pass" "$name"

    # Test connection
    info "Testing connection..."
    if PGPASSWORD="$pass" psql -h "$host" -p "$port" -U "$user" -d "$name" -c '\q' &>/dev/null; then
        success "Connection verified."
    else
        warn "Connection test failed — check pg_hba.conf and that the service is running."
    fi

    tui_msg "Setup complete!

Your .env has been updated with the database settings.
Airstats will apply schema migrations automatically on first start.

Next steps:
  • Fill in the remaining .env settings:
      READSB_AIRCRAFT_JSON, LAT, LON, RADIUS, etc.
  • Start Airstats:
      docker compose up
      — or —
      go run ./core"
}

main "$@"
