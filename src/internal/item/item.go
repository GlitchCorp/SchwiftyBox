package item

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"backend/internal/database"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

// Module provides item service dependency injection
var Module = fx.Module("item",
	fx.Provide(NewItemService),
)

// Service handles item operations
type Service struct {
	db *gorm.DB
}

var (
	// ErrItemNotFound is returned when item is not found
	ErrItemNotFound = errors.New("item not found")
	// ErrItemAlreadyExists is returned when trying to create an item that already exists
	ErrItemAlreadyExists = errors.New("item already exists")
)

// NewItemService creates a new item service
func NewItemService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// generatePrefix generates a random 3-letter prefix
func (s *Service) generatePrefix() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, 3)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

// getNextID gets the next ID for a backpack prefix
func (s *Service) getNextID(prefix string) (string, error) {
	var nextNumber database.BackPackIdNextNumber

	// Use transaction to ensure atomicity
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Try to find existing record
	err := tx.Where("backpack_id = ?", prefix).First(&nextNumber).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new record
			nextNumber = database.BackPackIdNextNumber{
				BackpackID: prefix,
				Number:     1,
			}
			if err := tx.Create(&nextNumber).Error; err != nil {
				tx.Rollback()
				return "", err
			}
		} else {
			tx.Rollback()
			return "", err
		}
	} else {
		// Increment existing record
		nextNumber.Number++
		if err := tx.Save(&nextNumber).Error; err != nil {
			tx.Rollback()
			return "", err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return "", err
	}

	// Format as 4-digit string with leading zeros
	return fmt.Sprintf("%04d", nextNumber.Number), nil
}

// CreateItem creates a new item
func (s *Service) CreateItem(name, description, userEmail string, parentID *uint) (*database.Item, error) {
	// Get user to check prefix
	var user database.User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		return nil, err
	}

	// Generate prefix if not exists
	if user.Prefix == "" {
		user.Prefix = s.generatePrefix()
		if err := s.db.Save(&user).Error; err != nil {
			return nil, err
		}
	}

	// Get next ID
	nextID, err := s.getNextID(user.Prefix)
	if err != nil {
		return nil, err
	}

	backpackID := user.Prefix + nextID

	item := &database.Item{
		Name:        name,
		BackpackID:  backpackID,
		Description: description,
		AddedAt:     time.Now(),
		UserEmail:   userEmail,
		ParentID:    parentID,
	}

	if err := s.db.Create(item).Error; err != nil {
		return nil, err
	}

	// Load relationships
	s.db.Preload("Parent").Preload("Tags").First(item, item.ID)

	return item, nil
}

// GetItem retrieves an item by ID
func (s *Service) GetItem(id uint, userEmail string) (*database.Item, error) {
	var item database.Item

	if err := s.db.Preload("Parent").Preload("Tags").Preload("Children").
		Where("id = ? AND user_email = ?", id, userEmail).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}

	return &item, nil
}

// GetItems retrieves all items for a user
func (s *Service) GetItems(userEmail string, nameFilter string) ([]database.Item, error) {
	var items []database.Item

	query := s.db.Preload("Parent").Preload("Tags").Preload("Children").
		Where("user_email = ?", userEmail)

	if nameFilter != "" {
		query = query.Where("name ILIKE ?", "%"+nameFilter+"%")
	}

	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

// UpdateItem updates an item
func (s *Service) UpdateItem(id uint, userEmail string, name, description string, parentID *uint, tagIDs []uint) (*database.Item, error) {
	item, err := s.GetItem(id, userEmail)
	if err != nil {
		return nil, err
	}

	// Update basic fields
	item.Name = name
	item.Description = description
	item.ParentID = parentID

	if err := s.db.Save(item).Error; err != nil {
		return nil, err
	}

	// Update tags if provided
	if tagIDs != nil {
		// Clear existing tags
		s.db.Exec("DELETE FROM item_tags WHERE item_id = ?", id)

		// Add new tags
		for _, tagID := range tagIDs {
			s.db.Exec("INSERT INTO item_tags (item_id, tag_id) VALUES (?, ?)", id, tagID)
		}
	}

	// Reload with relationships
	s.db.Preload("Parent").Preload("Tags").Preload("Children").First(item, item.ID)

	return item, nil
}

// DeleteItem deletes an item
func (s *Service) DeleteItem(id uint, userEmail string) error {
	result := s.db.Where("id = ? AND user_email = ?", id, userEmail).Delete(&database.Item{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrItemNotFound
	}
	return nil
}

// GetItemsByTag retrieves all items with a specific tag
func (s *Service) GetItemsByTag(tagID uint, userEmail string) ([]database.Item, error) {
	var items []database.Item

	if err := s.db.Joins("JOIN item_tags ON items.id = item_tags.item_id").
		Preload("Parent").Preload("Tags").Preload("Children").
		Where("item_tags.tag_id = ? AND items.user_email = ?", tagID, userEmail).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}
