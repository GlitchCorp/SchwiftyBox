# Docker Setup for SchwiftyBox

This project now includes Docker Compose support with PostgreSQL database.

## Prerequisites

- Docker
- Docker Compose

## Quick Start

### Production Environment

1. **Start all services:**
   ```bash
   make docker-up
   ```

2. **Run migrations:**
   ```bash
   make docker-migrate-up
   ```

3. **Check logs:**
   ```bash
   make docker-logs
   ```

4. **Stop services:**
   ```bash
   make docker-down
   ```

### Development Environment

1. **Start development environment (with hot reload):**
   ```bash
   make docker-dev
   ```

2. **Stop development environment:**
   ```bash
   make docker-dev-down
   ```

## Available Commands

### Docker Commands
- `make docker-up` - Start all services in background
- `make docker-down` - Stop all services
- `make docker-build` - Build Docker images
- `make docker-logs` - Show logs from all services
- `make docker-restart` - Restart all services
- `make docker-dev` - Start development environment with hot reload
- `make docker-dev-down` - Stop development environment

### Database Commands (Docker)
- `make docker-migrate-up` - Run pending migrations
- `make docker-migrate-down` - Rollback last migration
- `make docker-migrate-version` - Show current migration version

## Services

### PostgreSQL Database
- **Container:** `schwiftybox-postgres`
- **Port:** `5432`
- **Database:** `schwiftybox`
- **User:** `schwiftybox_user`
- **Password:** `schwiftybox_password` (production) / `schwiftybox_password_dev` (development)

### Go Application
- **Container:** `schwiftybox-app`
- **Port:** `8080`
- **Environment:** Production (Dockerfile) or Development (Dockerfile.dev)

## Environment Variables

The application uses the following environment variables:

```bash
DB_HOST=postgres
DB_PORT=5432
DB_USER=schwiftybox_user
DB_PASSWORD=schwiftybox_password
DB_NAME=schwiftybox
DB_URL=postgres://schwiftybox_user:schwiftybox_password@postgres:5432/schwiftybox?sslmode=disable
```

## Development vs Production

### Development Environment
- Uses `docker-compose.override.yml` for development-specific settings
- Hot reload enabled (source code mounted as volume)
- Debug mode enabled
- Uses `Dockerfile.dev` with Go toolchain

### Production Environment
- Uses optimized `Dockerfile` with multi-stage build
- Minimal Alpine Linux image
- Non-root user for security
- Compiled binary

## Database Persistence

PostgreSQL data is persisted using Docker volumes:
- **Production:** `postgres_data`
- **Development:** `postgres_data_dev`

## Troubleshooting

### Check if services are running
```bash
docker-compose ps
```

### View specific service logs
```bash
docker-compose logs postgres
docker-compose logs app
```

### Access PostgreSQL directly
```bash
docker-compose exec postgres psql -U schwiftybox_user -d schwiftybox
```

### Rebuild images
```bash
make docker-build
```

### Reset everything (WARNING: This will delete all data)
```bash
docker-compose down -v
docker-compose up -d
```

## Manual Docker Commands

If you prefer to use Docker Compose directly:

```bash
# Start services
docker-compose up -d

# Start with logs
docker-compose up

# Stop services
docker-compose down

# Build images
docker-compose build

# Run migrations
docker-compose exec app ./scripts/migrate.sh up
```
