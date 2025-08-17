package migrations

import (
	"errors"
	"log"

	"backend/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

// Module provides migration dependency injection
var Module = fx.Module("migrations",
	fx.Provide(NewMigrationService),
)

// Service handles database migrations
type Service struct {
	migrate *migrate.Migrate
}

// NewMigrationService creates a new migration service
func NewMigrationService(db *gorm.DB, cfg *config.Config) (*Service, error) {
	// Get underlying sql.DB from GORM
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Create postgres driver instance
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations", // Path to migrations folder
		"postgres",
		driver,
	)
	if err != nil {
		return nil, err
	}

	return &Service{
		migrate: m,
	}, nil
}

// Up runs all pending migrations
func (s *Service) Up() error {
	err := s.migrate.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("No new migrations to apply")
	} else {
		log.Println("Migrations applied successfully")
	}

	return nil
}

// Down reverts the last migration
func (s *Service) Down() error {
	err := s.migrate.Steps(-1)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("No migrations to revert")
	} else {
		log.Println("Migration reverted successfully")
	}

	return nil
}

// Version returns current migration version
func (s *Service) Version() (uint, bool, error) {
	return s.migrate.Version()
}

// Close closes the migration instance
func (s *Service) Close() error {
	sourceErr, dbErr := s.migrate.Close()
	if sourceErr != nil {
		return sourceErr
	}
	return dbErr
}
