package database

import (
	"context"
	"log"

	"backend/internal/config"

	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Module provides database dependency injection
var Module = fx.Module("database",
	fx.Provide(NewDatabase),
)

// User represents a user in the database
type User struct {
	Email    string `json:"email" gorm:"primaryKey"`
	Password string `json:"password"`
}

// NewDatabase creates a new database connection
func NewDatabase(lc fx.Lifecycle, cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.ConnectionString()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Println("Database connection established")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			log.Println("Closing database connection")
			return sqlDB.Close()
		},
	})

	return db, nil
}
