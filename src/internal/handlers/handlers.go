package handlers

import (
	"errors"
	"net/http"

	"backend/internal/jwt"
	"backend/internal/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Module provides handlers dependency injection
var Module = fx.Module("handlers",
	fx.Provide(NewHandlers),
)

// Handlers contains all HTTP handlers
type Handlers struct {
	userService *user.Service
	jwtService  *jwt.Service
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RefreshRequest represents the refresh token request body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// NewHandlers creates a new handlers instance
func NewHandlers(userService *user.Service, jwtService *jwt.Service) *Handlers {
	return &Handlers{
		userService: userService,
		jwtService:  jwtService,
	}
}

// RegisterUser handles user registration
func (h *Handlers) RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if err := h.userService.CreateUser(req.Email, req.Password); err != nil {
		if errors.Is(err, user.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// Login handles user login
func (h *Handlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if err := h.userService.ValidateUser(req.Email, req.Password); err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	tokens, err := h.jwtService.GenerateTokenPair(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// RefreshToken handles token refresh
func (h *Handlers) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	email, err := h.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Verify user still exists
	if _, err := h.userService.GetUser(email); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate user"})
		return
	}

	// Generate new access token
	newToken, err := h.jwtService.GenerateToken(email, h.jwtService.GetAccessTokenDuration())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}
