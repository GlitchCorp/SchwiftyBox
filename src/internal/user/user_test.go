package user

import (
	"errors"
	"testing"

	"backend/internal/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate the User model
	err = db.AutoMigrate(&database.User{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestNewUserService(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	if service.db == nil {
		t.Error("User service database should not be nil")
	}
}

func TestCreateUser_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	email := "test@example.com"
	password := "password123"

	err := service.CreateUser(email, password)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify user was created
	var user database.User
	err = db.First(&user, "email = ?", email).Error
	if err != nil {
		t.Fatalf("Failed to find created user: %v", err)
	}

	if user.Email != email {
		t.Errorf("Expected email '%s', got '%s'", email, user.Email)
	}
	if user.Password != password {
		t.Errorf("Expected password '%s', got '%s'", password, user.Password)
	}
}

func TestCreateUser_DuplicateUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	email := "test@example.com"
	password := "password123"

	// Create user first time
	err := service.CreateUser(email, password)
	if err != nil {
		t.Fatalf("Failed to create user first time: %v", err)
	}

	// Try to create same user again
	err = service.CreateUser(email, password)
	if err == nil {
		t.Error("Should return error for duplicate user")
	}
	// SQLite returns a different error message, so we check for the error type
	if err != nil && !errors.Is(err, ErrUserAlreadyExists) {
		// Check if it's a constraint violation error
		if !errors.Is(err, gorm.ErrDuplicatedKey) {
			// SQLite returns a different error message, so we check the error string
			if err.Error() != "UNIQUE constraint failed: users.email" {
				t.Errorf("Expected duplicate user error, got %v", err)
			}
		}
	}
}

func TestValidateUser_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	email := "test@example.com"
	password := "password123"

	// Create user first
	err := service.CreateUser(email, password)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Validate user with correct credentials
	err = service.ValidateUser(email, password)
	if err != nil {
		t.Fatalf("Failed to validate user with correct credentials: %v", err)
	}
}

func TestValidateUser_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	email := "nonexistent@example.com"
	password := "password123"

	err := service.ValidateUser(email, password)
	if err == nil {
		t.Error("Should return error for non-existent user")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestValidateUser_WrongPassword(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	email := "test@example.com"
	password := "password123"
	wrongPassword := "wrongpassword"

	// Create user first
	err := service.CreateUser(email, password)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Validate user with wrong password
	err = service.ValidateUser(email, wrongPassword)
	if err == nil {
		t.Error("Should return error for wrong password")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestGetUser_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	email := "test@example.com"
	password := "password123"

	// Create user first
	err := service.CreateUser(email, password)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get user
	user, err := service.GetUser(email)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if user.Email != email {
		t.Errorf("Expected email '%s', got '%s'", email, user.Email)
	}
	if user.Password != password {
		t.Errorf("Expected password '%s', got '%s'", password, user.Password)
	}
}

func TestGetUser_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	email := "nonexistent@example.com"

	user, err := service.GetUser(email)
	if err == nil {
		t.Error("Should return error for non-existent user")
	}
	if user != nil {
		t.Error("Should return nil user for non-existent user")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestCreateUser_EmptyEmail(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	email := ""
	password := "password123"

	err := service.CreateUser(email, password)
	if err != nil {
		t.Fatalf("Should not fail with empty email: %v", err)
	}

	// Verify user was created
	var user database.User
	err = db.First(&user, "email = ?", email).Error
	if err != nil {
		t.Fatalf("Failed to find created user: %v", err)
	}

	if user.Email != email {
		t.Errorf("Expected email '%s', got '%s'", email, user.Email)
	}
}
