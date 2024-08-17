package reset

import (
	"errors"
	"math/rand"
	"time"

	"backend/internal/database"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

// Module provides reset service dependency injection
var Module = fx.Module("reset",
	fx.Provide(NewResetService),
)

// Service handles password reset operations
type Service struct {
	db *gorm.DB
}

var (
	// ErrResetTokenNotFound is returned when reset token is not found
	ErrResetTokenNotFound = errors.New("reset token not found")
	// ErrResetTokenExpired is returned when reset token has expired
	ErrResetTokenExpired = errors.New("reset token expired")
)

// NewResetService creates a new reset service
func NewResetService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// generateToken generates a random 15-character token
func (s *Service) generateToken() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 15)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

// CreateResetToken creates a new reset token for a user
func (s *Service) CreateResetToken(userEmail string) (string, error) {
	// Check if user exists
	var user database.User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("user not found")
		}
		return "", err
	}

	// Delete any existing tokens for this user
	s.db.Where("user_email = ?", userEmail).Delete(&database.ResetToken{})

	// Generate new token
	token := s.generateToken()
	expiredAt := time.Now().Add(5 * time.Minute) // Token expires in 5 minutes

	resetToken := &database.ResetToken{
		Token:     token,
		ExpiredAt: expiredAt,
		UserEmail: userEmail,
	}

	if err := s.db.Create(resetToken).Error; err != nil {
		return "", err
	}

	return token, nil
}

// ValidateResetToken validates a reset token
func (s *Service) ValidateResetToken(token string) (*database.ResetToken, error) {
	var resetToken database.ResetToken

	if err := s.db.Where("token = ?", token).First(&resetToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResetTokenNotFound
		}
		return nil, err
	}

	// Check if token has expired
	if time.Now().After(resetToken.ExpiredAt) {
		// Delete expired token
		s.db.Delete(&resetToken)
		return nil, ErrResetTokenExpired
	}

	return &resetToken, nil
}

// ResetPassword resets a user's password using a valid token
func (s *Service) ResetPassword(token, newPassword string) error {
	resetToken, err := s.ValidateResetToken(token)
	if err != nil {
		return err
	}

	// Update user password
	if err := s.db.Model(&database.User{}).
		Where("email = ?", resetToken.UserEmail).
		Update("password", newPassword).Error; err != nil {
		return err
	}

	// Delete the used token
	s.db.Delete(resetToken)

	return nil
}

// CleanupExpiredTokens removes all expired reset tokens
func (s *Service) CleanupExpiredTokens() error {
	return s.db.Where("expired_at < ?", time.Now()).Delete(&database.ResetToken{}).Error
}
