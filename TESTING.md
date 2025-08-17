# Testing Documentation

This document describes the testing strategy and how to run different types of tests in the SchwiftyBox project.

## Test Types

### 1. Unit Tests
Unit tests test individual components in isolation using mocks and test doubles.

**Location:** `src/internal/*/test.go` files

**Coverage:**
- `config`: 100%
- `user`: 85.7%
- `jwt`: 81.8%
- `handlers`: 58.7%
- `migrations`: 17.2%
- `database`: 0%
- `server`: 0%

**Run unit tests:**
```bash
make test
```

**Run unit tests with coverage:**
```bash
make test-coverage
```

### 2. Integration Tests
Integration tests test the interaction between components using a real PostgreSQL database.

**Location:** `src/integration_test.go`

**Features:**
- Tests complete user registration flow
- Tests user login with valid/invalid credentials
- Tests token refresh functionality
- Tests input validation
- Tests database connectivity
- Uses real PostgreSQL database (`schwiftybox_test`)

**Run integration tests:**
```bash
make test-integration
```

**Prerequisites:**
- Docker containers must be running (`make docker-up`)
- PostgreSQL database must be accessible

### 3. End-to-End Tests
E2E tests test the complete application through real HTTP requests to the running server.

**Location:** `src/e2e_test.go`

**Features:**
- Tests API endpoints through real HTTP server
- Tests complete user workflows
- Tests server responsiveness
- Uses real application running on port 8080

**Run E2E tests:**
```bash
make test-e2e
```

**Prerequisites:**
- Application must be running (`make docker-up`)
- Server must be accessible on port 8080

## Test Database Setup

### Integration Tests Database
Integration tests use a separate test database (`schwiftybox_test`) to avoid interfering with the main application data.

**Setup:**
```bash
# Create test database
docker exec schwiftybox-postgres psql -U schwiftybox_user -d postgres -c "CREATE DATABASE schwiftybox_test;"
```

**Cleanup:**
The test suite automatically cleans up data between tests and drops tables after completion.

## Test Environment Variables

### Integration Tests
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=schwiftybox_user
DB_PASSWORD=schwiftybox_password_dev
DB_NAME=schwiftybox_test
JWT_SECRET=test-secret-key
SERVER_PORT=:8080
```

### E2E Tests
```bash
# Uses application running on localhost:8080
# No additional environment variables needed
```

## Running All Tests

### Complete Test Suite
```bash
# Run all unit tests
make test

# Run integration tests
make test-integration

# Run E2E tests
make test-e2e
```

### Test Coverage Report
```bash
make test-coverage
```

This generates:
- `src/coverage.out` - Coverage data
- `src/coverage.html` - HTML coverage report

## Test Structure

### Test Suites
Both integration and E2E tests use the `testify/suite` package for better organization:

```go
type IntegrationTestSuite struct {
    suite.Suite
    db         *gorm.DB
    router     *gin.Engine
    userService *user.Service
    jwtService  *jwt.Service
    handlers    *handlers.Handlers
}
```

### Test Lifecycle
1. **SetupSuite()** - Runs once before all tests
2. **SetupTest()** - Runs before each test
3. **Test methods** - Individual test cases
4. **TearDownSuite()** - Runs once after all tests

## Test Data Management

### Unique Test Data
Tests use unique email addresses to avoid conflicts:

```go
email := fmt.Sprintf("test_%d@example.com", time.Now().Unix())
```

### Database Cleanup
- Integration tests clear the users table before each test
- E2E tests use unique timestamps to avoid conflicts
- Test suites clean up after completion

## Continuous Integration

### GitHub Actions (Recommended)
Add to `.github/workflows/test.yml`:

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: schwiftybox_test
          POSTGRES_USER: schwiftybox_user
          POSTGRES_PASSWORD: schwiftybox_password_dev
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - run: make test
      - run: make test-integration
      - run: make test-coverage
```

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Ensure Docker containers are running: `make docker-up`
   - Check database credentials in environment variables
   - Verify PostgreSQL is healthy: `docker ps`

2. **Port Already in Use**
   - Stop existing application: `make docker-down`
   - Check for processes using port 8080: `netstat -tlnp | grep 8080`

3. **Test Database Not Found**
   - Create test database: `docker exec schwiftybox-postgres psql -U schwiftybox_user -d postgres -c "CREATE DATABASE schwiftybox_test;"`

4. **Slow Tests**
   - Integration tests can be slow due to database operations
   - Consider using test containers for faster setup/teardown

### Debug Mode
Run tests with verbose output:
```bash
go test -v ./...
```

### Skip E2E Tests
```bash
SKIP_E2E=true make test-e2e
```

## Best Practices

1. **Test Isolation** - Each test should be independent
2. **Unique Data** - Use unique identifiers to avoid conflicts
3. **Cleanup** - Always clean up test data
4. **Realistic Data** - Use realistic test data
5. **Error Handling** - Test both success and failure scenarios
6. **Performance** - Monitor test execution time

## Future Improvements

1. **Mock Implementation** - Replace real database with mocks in unit tests
2. **Test Containers** - Use testcontainers for faster database setup
3. **Performance Tests** - Add load testing
4. **Security Tests** - Add security testing
5. **API Documentation Tests** - Test API documentation accuracy
