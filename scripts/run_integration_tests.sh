#!/bin/bash

# Integration Tests Runner Script
# This script runs integration tests against a real PostgreSQL database

set -e

echo "Starting Integration Tests..."

# Check if Docker containers are running
if ! docker ps | grep -q "schwiftybox-postgres"; then
    echo "Error: PostgreSQL container is not running. Please start the application first with 'make docker-up'"
    exit 1
fi

# Create test database if it doesn't exist
echo "Setting up test database..."
docker exec schwiftybox-postgres psql -U schwiftybox_user -d postgres -c "CREATE DATABASE schwiftybox_test;" 2>/dev/null || echo "Test database already exists"

# Run migrations on test database
echo "Running migrations on test database..."
docker exec schwiftybox-postgres psql -U schwiftybox_user -d schwiftybox_test -c "DROP TABLE IF EXISTS users CASCADE;" 2>/dev/null || true

# Set environment variables for tests
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=schwiftybox_user
export DB_PASSWORD=schwiftybox_password_dev
export DB_NAME=schwiftybox_test
export JWT_SECRET=test-secret-key
export SERVER_PORT=:8080

# Run integration tests
echo "Running integration tests..."
cd src
go test -v -run TestIntegrationTestSuite ./integration_test.go

echo "Integration tests completed!"
