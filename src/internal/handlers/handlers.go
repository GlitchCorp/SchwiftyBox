package handlers

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"backend/internal/database"
	"backend/internal/jwt"
	"backend/internal/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

// Module provides handlers dependency injection
var Module = fx.Module("handlers",
	fx.Provide(NewHandlers),
)

// Handlers contains all HTTP handlers
type Handlers struct {
	userService *user.Service
	jwtService  *jwt.Service
	db          *gorm.DB
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

// VerifyRequest represents the verify token request body
type VerifyRequest struct {
	Token string `json:"token" binding:"required"`
}

// ItemCreateRequest represents the item creation request body
type ItemCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// ItemUpdateRequest represents the item update request body
type ItemUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    *uint  `json:"parent"`
	Tags        []uint `json:"tags"`
}

// TagCreateRequest represents the tag creation request body
type TagCreateRequest struct {
	Name string `json:"name" binding:"required"`
}

// PasswordResetRequest represents the password reset request body
type PasswordResetRequest struct {
	Username string `json:"username" binding:"required"`
}

// NewPasswordRequest represents the new password request body
type NewPasswordRequest struct {
	Password string `json:"password" binding:"required"`
	Token    string `json:"token" binding:"required"`
}

// UserUpdateRequest represents the user update request body
type UserUpdateRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// NewHandlers creates a new handlers instance
func NewHandlers(userService *user.Service, jwtService *jwt.Service, db *gorm.DB) *Handlers {
	return &Handlers{
		userService: userService,
		jwtService:  jwtService,
		db:          db,
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
	log.Printf("Login attempt - Content-Type: %s", c.GetHeader("Content-Type"))

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Login validation error: %v", err)
		log.Printf("Request body: %+v", req)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	log.Printf("Login request for email: %s", req.Email)

	if err := h.userService.ValidateUser(req.Email, req.Password); err != nil {
		log.Printf("User validation error: %v", err)
		if errors.Is(err, user.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	log.Printf("User validation successful for: %s", req.Email)

	log.Printf("Calling GenerateTokenPair for email: %s", req.Email)
	tokens, err := h.jwtService.GenerateTokenPair(req.Email)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	log.Printf("Login successful for: %s", req.Email)
	log.Printf("Generated tokens: %+v", tokens)
	c.JSON(http.StatusOK, tokens)
}

// RefreshToken handles token refresh
func (h *Handlers) RefreshToken(c *gin.Context) {
	log.Printf("=== REFRESH TOKEN REQUEST START ===")
	log.Printf("Method: %s", c.Request.Method)
	log.Printf("URL: %s", c.Request.URL.String())
	log.Printf("Content-Type: %s", c.GetHeader("Content-Type"))
	log.Printf("User-Agent: %s", c.GetHeader("User-Agent"))

	// Log raw body for debugging
	body, err := c.GetRawData()
	if err != nil {
		log.Printf("Failed to read raw body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	log.Printf("Raw body: %s", string(body))

	// Restore body for binding
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("JSON binding error: %v", err)
		log.Printf("Request struct: %+v", req)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	log.Printf("Parsed request: %+v", req)
	log.Printf("Refresh token length: %d", len(req.RefreshToken))
	if len(req.RefreshToken) > 20 {
		log.Printf("Refresh token preview: %s...", req.RefreshToken[:20])
	} else {
		log.Printf("Refresh token: %s", req.RefreshToken)
	}

	// Validate refresh token
	email, err := h.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		log.Printf("Refresh token validation failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	log.Printf("Refresh token validated for user: %s", email)

	// Verify user still exists
	if _, err := h.userService.GetUser(email); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			log.Printf("User not found during refresh: %s", email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		log.Printf("Failed to validate user during refresh: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate user"})
		return
	}

	log.Printf("User verified: %s", email)

	// Generate new token pair (both access and refresh tokens)
	tokens, err := h.jwtService.GenerateTokenPair(email)
	if err != nil {
		log.Printf("Failed to generate new tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	log.Printf("New tokens generated successfully")
	log.Printf("=== REFRESH TOKEN REQUEST END ===")

	c.JSON(http.StatusOK, tokens)
}

// TokenPair handles generating a new token pair for a user
func (h *Handlers) TokenPair(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Verify user exists
	if _, err := h.userService.GetUser(req.Email); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate user"})
		return
	}

	tokens, err := h.jwtService.GenerateTokenPair(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// VerifyToken handles token verification
func (h *Handlers) VerifyToken(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	email, err := h.jwtService.ValidateToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true, "email": email})
}

// GetUserStatistics handles getting user statistics
func (h *Handlers) GetUserStatistics(c *gin.Context) {
	var count int64
	if err := h.db.Model(&database.User{}).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": count})
}

// RequestPasswordReset handles password reset request
func (h *Handlers) RequestPasswordReset(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check if user exists
	var user database.User
	if err := h.db.Where("email = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// Generate reset token (simple random string for now)
	token := "reset_" + strconv.FormatInt(time.Now().Unix(), 10)
	expiredAt := time.Now().Add(5 * time.Minute)

	resetToken := database.ResetToken{
		Token:     token,
		ExpiredAt: expiredAt,
		UserEmail: user.Email,
	}

	// Save reset token
	if err := h.db.Create(&resetToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reset token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset token sent"})
}

// SetNewPassword handles setting new password with reset token
func (h *Handlers) SetNewPassword(c *gin.Context) {
	var req NewPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Find reset token
	var resetToken database.ResetToken
	if err := h.db.Where("token = ? AND expired_at > ?", req.Token, time.Now()).First(&resetToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate token"})
		return
	}

	// Update user password (for now, store as plain text - in production should be hashed)
	if err := h.db.Model(&database.User{}).Where("email = ?", resetToken.UserEmail).Update("password", req.Password).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Delete used reset token
	h.db.Delete(&resetToken)

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// DeactivateUser handles user deactivation
func (h *Handlers) DeactivateUser(c *gin.Context) {
	_, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// For now, we'll just return success since we don't have a deactivated field
	// In a real implementation, you'd set a deactivated flag
	c.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}

// GetUserDetails handles getting user details
func (h *Handlers) GetUserDetails(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user database.User
	if err := h.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"email": user.Email})
}

// UpdateUserDetails handles updating user details
func (h *Handlers) UpdateUserDetails(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	var user database.User
	if err := h.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Update email
	if err := h.db.Model(&user).Update("email", req.Email).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"email": req.Email})
}

// DeleteUser handles user deletion
func (h *Handlers) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.db.Delete(&database.User{}, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetItems handles getting all items for the authenticated user
func (h *Handlers) GetItems(c *gin.Context) {
	userEmail, exists := c.Get("user_email")
	log.Printf("GetItems called for user: %s", userEmail)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	nameFilter := c.Query("name")
	log.Printf("Name filter: %s", nameFilter)

	var items []database.Item
	query := h.db.Where("user_email = ?", userEmail).Preload("Tags").Preload("Parent")

	if nameFilter != "" {
		// Use LIKE for SQLite compatibility (case-insensitive search)
		query = query.Where("name LIKE ?", "%"+nameFilter+"%")
	}

	if err := query.Find(&items).Error; err != nil {
		log.Printf("Failed to get items: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get items"})
		return
	}

	log.Printf("Found %d items 1", len(items))
	log.Printf("Returning items in object with key 'items'")
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GetItem handles getting a specific item by ID
func (h *Handlers) GetItem(c *gin.Context) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	itemIDStr := c.Param("item_id")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var item database.Item
	if err := h.db.Where("id = ? AND user_email = ?", itemID, userEmail).Preload("Tags").Preload("Parent").First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// CreateItem handles creating a new item
func (h *Handlers) CreateItem(c *gin.Context) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req ItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Get user to generate backpack_id
	var user database.User
	if err := h.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Generate backpack_id (simple implementation)
	backpackID := user.Prefix + "_" + strconv.FormatInt(time.Now().Unix(), 10)

	item := database.Item{
		Name:        req.Name,
		Description: req.Description,
		BackpackID:  backpackID,
		AddedAt:     time.Now(),
		UserEmail:   userEmail.(string),
	}

	if err := h.db.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
		return
	}

	// Load relationships
	h.db.Preload("Tags").Preload("Parent").First(&item, item.ID)

	c.JSON(http.StatusCreated, item)
}

// UpdateItem handles updating an existing item
func (h *Handlers) UpdateItem(c *gin.Context) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	itemIDStr := c.Param("item_id")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var req ItemUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	var item database.Item
	if err := h.db.Where("id = ? AND user_email = ?", itemID, userEmail).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get item"})
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ParentID != nil {
		updates["parent_id"] = req.ParentID
	}

	if len(updates) > 0 {
		if err := h.db.Model(&item).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
			return
		}
	}

	// Update tags if provided
	if req.Tags != nil {
		var tags []database.Tag
		if err := h.db.Where("id IN ?", req.Tags).Find(&tags).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tags"})
			return
		}
		h.db.Model(&item).Association("Tags").Replace(tags)
	}

	// Load updated item with relationships
	h.db.Preload("Tags").Preload("Parent").First(&item, item.ID)

	c.JSON(http.StatusOK, item)
}

// DeleteItem handles deleting an item
func (h *Handlers) DeleteItem(c *gin.Context) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	itemIDStr := c.Param("item_id")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	if err := h.db.Where("id = ? AND user_email = ?", itemID, userEmail).Delete(&database.Item{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetTags handles getting all tags for the user's organization
func (h *Handlers) GetTags(c *gin.Context) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user's organization
	var user database.User
	if err := h.db.Where("email = ?", userEmail).Preload("ActiveOrganization").First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	var tags []database.Tag
	if err := h.db.Where("organization_id = ?", user.ActiveOrganizationID).Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tags"})
		return
	}

	c.JSON(http.StatusOK, tags)
}

// CreateTag handles creating a new tag
func (h *Handlers) CreateTag(c *gin.Context) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req TagCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Get user's organization
	var user database.User
	if err := h.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	tag := database.Tag{
		Name:           req.Name,
		OrganizationID: user.ActiveOrganizationID,
	}

	if err := h.db.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag"})
		return
	}

	c.JSON(http.StatusCreated, tag)
}

// GetJWTService returns the JWT service for middleware
func (h *Handlers) GetJWTService() *jwt.Service {
	return h.jwtService
}
