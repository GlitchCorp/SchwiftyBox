package database

import (
	"context"
	"log"
	"time"

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
	Email     string    `json:"email" gorm:"primaryKey"`
	Password  string    `json:"password"`
	Prefix    string    `json:"prefix" gorm:"size:10"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	ActiveOrganizationID uint `json:"active_organization_id"`
	ActiveOrganization   Organization `json:"active_organization" gorm:"foreignKey:ActiveOrganizationID"`
	Items                []Item `json:"items" gorm:"foreignKey:UserEmail"`
}

// Organization represents an organization in the database
type Organization struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	Users []User `json:"users" gorm:"many2many:organization_users;"`
	Tags  []Tag  `json:"tags" gorm:"foreignKey:OrganizationID"`
}

// Item represents an item in the database
type Item struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:200"`
	BackpackID  string    `json:"backpack_id" gorm:"size:20"`
	Description string    `json:"description" gorm:"size:1000"`
	AddedAt     time.Time `json:"added_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	UserEmail string `json:"user_email"`
	User      User   `json:"user" gorm:"foreignKey:UserEmail"`
	
	ParentID *uint `json:"parent_id"`
	Parent   *Item `json:"parent" gorm:"foreignKey:ParentID"`
	Children []Item `json:"children" gorm:"foreignKey:ParentID"`
	
	Tags []Tag `json:"tags" gorm:"many2many:item_tags;"`
}

// Tag represents a tag in the database
type Tag struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"size:20"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relationships
	OrganizationID uint `json:"organization_id"`
	Organization   Organization `json:"organization" gorm:"foreignKey:OrganizationID"`
	
	Items []Item `json:"items" gorm:"many2many:item_tags;"`
}

// BackPackIdNextNumber represents the next number for backpack ID generation
type BackPackIdNextNumber struct {
	ID         uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	BackpackID string `json:"backpack_id" gorm:"size:20"`
	Number     int    `json:"number"`
}

// ResetToken represents a password reset token
type ResetToken struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Token     string    `json:"token" gorm:"size:30;uniqueIndex"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	
	// Relationships
	UserEmail string `json:"user_email"`
	User      User   `json:"user" gorm:"foreignKey:UserEmail"`
}

// NewDatabase creates a new database connection
func NewDatabase(lc fx.Lifecycle, cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.ConnectionString()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Note: Migrations are now handled by the migrations package
	// Auto-migration is disabled for production use

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
