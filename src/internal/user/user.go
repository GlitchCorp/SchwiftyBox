package user

import (
	"errors"

	"backend/internal/database"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

// Module provides user service dependency injection
var Module = fx.Module("user",
	fx.Provide(NewUserService),
)

// Service handles user operations
type Service struct {
	db *gorm.DB
}

var (
	// ErrUserAlreadyExists is returned when trying to create a user that already exists
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = errors.New("user not found")
)

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// CreateUser creates a new user
func (s *Service) CreateUser(email, password string) error {
	user := &database.User{
		Email:    email,
		Password: password,
	}

	if err := s.db.Create(user).Error; err != nil {
		// Check if it's a duplicate key error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

// ValidateUser validates user credentials
func (s *Service) ValidateUser(email, password string) error {
	var user database.User

	if err := s.db.First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidCredentials
		}
		return err
	}

	if user.Password != password {
		return ErrInvalidCredentials
	}

	return nil
}

// GetUser retrieves a user by email
func (s *Service) GetUser(email string) (*database.User, error) {
	var user database.User

	if err := s.db.First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
