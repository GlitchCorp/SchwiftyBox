package jwt

import (
	"time"

	"backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/fx"
)

// Module provides JWT service dependency injection
var Module = fx.Module("jwt",
	fx.Provide(NewJWTService),
)

// Service handles JWT operations
type Service struct {
	config *config.JWTConfig
}

// TokenResponse represents the response containing tokens
type TokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// NewJWTService creates a new JWT service
func NewJWTService(cfg *config.Config) *Service {
	return &Service{
		config: &cfg.JWT,
	}
}

// GenerateToken generates a JWT token with the specified email and expiry
func (s *Service) GenerateToken(email string, expiry time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(expiry).Unix(),
	})
	return token.SignedString([]byte(s.config.SecretKey))
}

// GenerateTokenPair generates both access and refresh tokens
func (s *Service) GenerateTokenPair(email string) (*TokenResponse, error) {
	accessToken, err := s.GenerateToken(email, s.config.AccessTokenDuration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenerateToken(email, s.config.RefreshTokenDuration)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateToken validates a JWT token and returns the email claim
func (s *Service) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrInvalidKey
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", jwt.ErrInvalidKey
	}

	return email, nil
}

// GetAccessTokenDuration returns the access token duration
func (s *Service) GetAccessTokenDuration() time.Duration {
	return s.config.AccessTokenDuration
}
