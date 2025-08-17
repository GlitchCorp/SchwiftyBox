# SchwiftyBox

[![Tests](https://github.com/GlitchCorp/schwiftybox/workflows/Run%20tests%20and%20upload%20coverage/badge.svg)](https://github.com/GlitchCorp/schwiftybox/actions)
[![Test Coverage](https://codecov.io/gh/GlitchCorp/schwiftybox/branch/master/graph/badge.svg)](https://codecov.io/gh/GlitchCorp/schwiftybox)
[![Go Version](https://img.shields.io/badge/Go-1.23+-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

A modern, secure user authentication API built with Go, featuring JWT tokens, PostgreSQL database, and comprehensive testing.

## Features

- **User Authentication**: Register and login with email/password
- **JWT Tokens**: Secure access and refresh tokens
- **PostgreSQL Database**: Reliable data storage with migrations
- **RESTful API**: Clean HTTP endpoints
- **Docker Support**: Easy deployment with Docker Compose
- **Comprehensive Testing**: Unit, integration, and E2E tests
- **Modern Architecture**: Clean code with dependency injection

## Tech Stack

- **Backend**: Go 1.23 with Gin framework
- **Database**: PostgreSQL 15
- **ORM**: GORM
- **Authentication**: JWT (JSON Web Tokens)
- **Testing**: Testify with test suites
- **Containerization**: Docker & Docker Compose
- **Dependency Injection**: Uber FX

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.23+ (for local development)

### Using Docker (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd schwiftybox
   ```

2. **Start the application**
   ```bash
   make docker-up
   ```

3. **Verify it's running**
   ```bash
   curl http://localhost:8080/api/users
   ```

The application will be available at `http://localhost:8080`

### Local Development

1. **Install dependencies**
   ```bash
   make deps
   ```

2. **Start PostgreSQL**
   ```bash
   make docker-up
   ```

3. **Run migrations**
   ```bash
   make migrate-up
   ```

4. **Start the application**
   ```bash
   make run
   ```

## API Endpoints

### Authentication

#### Register User
```http
POST /api/users
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "message": "User created successfully"
}
```

#### Login
```http
POST /api/token
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Refresh Token
```http
POST /api/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_USER` | `user` | Database username |
| `DB_PASSWORD` | `password` | Database password |
| `DB_NAME` | `mydb` | Database name |
| `DB_SSLMODE` | `disable` | Database SSL mode |
| `JWT_SECRET` | `secret` | JWT signing secret |
| `SERVER_PORT` | `:8080` | Server port |

### Docker Environment

The Docker setup uses these default values:
- Database: `schwiftybox_user` / `schwiftybox_password_dev`
- Database name: `schwiftybox`
- JWT Secret: `test-secret-key`

## Development

### Project Structure

```
schwiftybox/
├── src/
│   ├── internal/
│   │   ├── config/      # Configuration management
│   │   ├── database/    # Database models and connection
│   │   ├── handlers/    # HTTP request handlers
│   │   ├── jwt/         # JWT token management
│   │   ├── migrations/  # Database migrations
│   │   ├── server/      # HTTP server setup
│   │   └── user/        # User service logic
│   ├── main.go          # Application entry point
│   ├── integration_test.go  # Integration tests
│   └── e2e_test.go      # End-to-end tests
├── migrations/           # Database migration files
├── scripts/             # Utility scripts
├── Dockerfile           # Production Docker image
├── Dockerfile.dev       # Development Docker image
├── docker-compose.yml   # Docker Compose configuration
└── Makefile            # Build and development commands
```

### Available Commands

#### Build and Run
```bash
make build          # Build the binary
make run            # Run the application
make clean          # Clean build files
```

#### Docker Commands
```bash
make docker-up      # Start all services
make docker-down    # Stop all services
make docker-build   # Build Docker images
make docker-dev     # Start development environment
```

#### Database Commands
```bash
make migrate-up     # Run pending migrations
make migrate-down   # Rollback last migration
make migrate-version # Show current migration version
make migrate-create NAME=migration_name  # Create new migration
```

#### Testing
```bash
make test           # Run unit tests
make test-coverage  # Run tests with coverage report
make test-integration # Run integration tests
make test-e2e       # Run end-to-end tests
```

#### Development
```bash
make deps           # Download dependencies
make dev-setup      # Setup development environment
```

## Testing

The project includes comprehensive testing at multiple levels with **44.8% overall test coverage**:

### Coverage by Package
- `config`: 100.0% - Complete configuration testing
- `user`: 85.7% - User service functionality
- `jwt`: 81.8% - JWT token management
- `handlers`: 58.7% - HTTP request handlers
- `migrations`: 17.2% - Database migrations
- `database`: 0.0% - Database connection (needs tests)
- `server`: 0.0% - HTTP server (needs tests)

### Unit Tests
- Test individual components in isolation
- Use mocks and test doubles
- Located in `src/internal/*/test.go` files

### Integration Tests
- Test component interactions with real database
- Use separate test database (`schwiftybox_test`)
- Located in `src/integration_test.go`

### End-to-End Tests
- Test complete application through HTTP API
- Use real running server
- Located in `src/e2e_test.go`

### Running Tests

```bash
# All unit tests
make test

# With coverage report
make test-coverage

# Integration tests (requires Docker)
make test-integration

# End-to-end tests (requires running application)
make test-e2e
```

For detailed testing documentation, see [TESTING.md](TESTING.md).

## Database Migrations

The project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

### Migration Files
- Located in `migrations/` directory
- Format: `000001_description.up.sql` and `000001_description.down.sql`

### Current Migrations
- `000001_create_users_table` - Creates users table with email and password

### Running Migrations
```bash
# Apply all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check current version
make migrate-version

# Create new migration
make migrate-create NAME=add_user_profile
```

## Security

### Password Storage
- Passwords are stored as plain text (for demo purposes)
- In production, use bcrypt or similar hashing

### JWT Configuration
- Access tokens: 15 minutes
- Refresh tokens: 24 hours
- Configurable via environment variables

### Input Validation
- Email format validation
- Password length requirements
- SQL injection protection via GORM

## Continuous Integration

### GitHub Actions
The project includes a GitHub Actions workflow (`.github/workflows/test.yml`) that:
- Runs unit and integration tests
- Generates coverage reports
- Uploads coverage to Codecov

### Codecov Setup
To enable dynamic coverage badges:
1. Connect your repository to [Codecov](https://codecov.io)
2. Update the badge URL in README.md with your actual repository name
3. The badge will automatically update with each commit

### Manual Setup
If you prefer to set up manually, create `.github/workflows/test.yml` with the provided configuration.

## Deployment

### Docker Production
```bash
# Build production image
docker build -t schwiftybox .

# Run with environment variables
docker run -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_USER=your-db-user \
  -e DB_PASSWORD=your-db-password \
  -e JWT_SECRET=your-secret \
  schwiftybox
```

### Docker Compose Production
```bash
# Create production compose file
cp docker-compose.yml docker-compose.prod.yml

# Edit environment variables
# Run production stack
docker-compose -f docker-compose.prod.yml up -d
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Development Guidelines

- Follow Go coding standards
- Add tests for new features
- Update documentation
- Use meaningful commit messages
- Keep commits atomic and focused

## Troubleshooting

### Common Issues

**Database Connection Failed**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check database logs
docker logs schwiftybox-postgres
```

**Port Already in Use**
```bash
# Stop existing containers
make docker-down

# Check for processes using port 8080
netstat -tlnp | grep 8080
```

**Migration Errors**
```bash
# Check migration status
make migrate-version

# Force migration to specific version
make migrate-force VERSION=1
```

**Test Failures**
```bash
# Run tests with verbose output
go test -v ./...

# Check test database
docker exec schwiftybox-postgres psql -U schwiftybox_user -d schwiftybox_test -c "\dt"
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [GORM](https://gorm.io/) - ORM library
- [Uber FX](https://github.com/uber-go/fx) - Dependency injection
- [golang-migrate](https://github.com/golang-migrate/migrate) - Database migrations
- [Testify](https://github.com/stretchr/testify) - Testing framework
