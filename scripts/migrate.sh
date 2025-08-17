#!/bin/bash

# Database migration script for schwiftybox
# Usage: ./scripts/migrate.sh [up|down|version|create]

set -e

# Default values
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-user}
DB_PASSWORD=${DB_PASSWORD:-password}
DB_NAME=${DB_NAME:-mydb}
DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

MIGRATIONS_PATH="./migrations"
MIGRATE_TOOL="migrate"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if migrate tool is installed
check_migrate_tool() {
    if ! command -v $MIGRATE_TOOL &> /dev/null; then
        echo -e "${RED}Error: migrate tool is not installed${NC}"
        echo "Install it with: make migrate-install"
        exit 1
    fi
}

# Print usage
usage() {
    echo "Usage: $0 [up|down|version|create <name>|force <version>]"
    echo ""
    echo "Commands:"
    echo "  up              - Apply all pending migrations"
    echo "  down            - Rollback the last migration"
    echo "  version         - Show current migration version"
    echo "  create <name>   - Create a new migration file"
    echo "  force <version> - Force database to specific version (dangerous!)"
    echo ""
    echo "Environment variables:"
    echo "  DB_HOST        - Database host (default: localhost)"
    echo "  DB_PORT        - Database port (default: 5432)"
    echo "  DB_USER        - Database user (default: user)"
    echo "  DB_PASSWORD    - Database password (default: password)"
    echo "  DB_NAME        - Database name (default: mydb)"
    exit 1
}

# Apply migrations
migrate_up() {
    echo -e "${YELLOW}Applying migrations...${NC}"
    $MIGRATE_TOOL -path $MIGRATIONS_PATH -database "$DB_URL" up
    echo -e "${GREEN}Migrations applied successfully!${NC}"
}

# Rollback last migration
migrate_down() {
    echo -e "${YELLOW}Rolling back last migration...${NC}"
    $MIGRATE_TOOL -path $MIGRATIONS_PATH -database "$DB_URL" down 1
    echo -e "${GREEN}Migration rolled back successfully!${NC}"
}

# Show current version
migrate_version() {
    echo -e "${YELLOW}Current migration version:${NC}"
    $MIGRATE_TOOL -path $MIGRATIONS_PATH -database "$DB_URL" version
}

# Create new migration
migrate_create() {
    if [ -z "$1" ]; then
        echo -e "${RED}Error: Migration name is required${NC}"
        echo "Usage: $0 create <migration_name>"
        exit 1
    fi
    
    echo -e "${YELLOW}Creating new migration: $1${NC}"
    $MIGRATE_TOOL create -ext sql -dir $MIGRATIONS_PATH -seq "$1"
    echo -e "${GREEN}Migration files created successfully!${NC}"
}

# Force to specific version
migrate_force() {
    if [ -z "$1" ]; then
        echo -e "${RED}Error: Version number is required${NC}"
        echo "Usage: $0 force <version>"
        exit 1
    fi
    
    echo -e "${RED}WARNING: This will force the database to version $1${NC}"
    echo -e "${RED}This operation can be dangerous and may cause data loss!${NC}"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        $MIGRATE_TOOL -path $MIGRATIONS_PATH -database "$DB_URL" force "$1"
        echo -e "${GREEN}Database forced to version $1${NC}"
    else
        echo -e "${YELLOW}Operation cancelled${NC}"
    fi
}

# Main script logic
check_migrate_tool

case "${1:-}" in
    "up")
        migrate_up
        ;;
    "down")
        migrate_down
        ;;
    "version")
        migrate_version
        ;;
    "create")
        migrate_create "$2"
        ;;
    "force")
        migrate_force "$2"
        ;;
    *)
        usage
        ;;
esac
