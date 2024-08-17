package tag

import (
	"errors"

	"backend/internal/database"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

// Module provides tag service dependency injection
var Module = fx.Module("tag",
	fx.Provide(NewTagService),
)

// Service handles tag operations
type Service struct {
	db *gorm.DB
}

var (
	// ErrTagNotFound is returned when tag is not found
	ErrTagNotFound = errors.New("tag not found")
	// ErrTagAlreadyExists is returned when trying to create a tag that already exists
	ErrTagAlreadyExists = errors.New("tag already exists")
)

// NewTagService creates a new tag service
func NewTagService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// CreateTag creates a new tag
func (s *Service) CreateTag(name string, organizationID uint) (*database.Tag, error) {
	tag := &database.Tag{
		Name:           name,
		OrganizationID: organizationID,
	}

	if err := s.db.Create(tag).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrTagAlreadyExists
		}
		return nil, err
	}

	return tag, nil
}

// GetTag retrieves a tag by ID
func (s *Service) GetTag(id uint) (*database.Tag, error) {
	var tag database.Tag

	if err := s.db.First(&tag, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	return &tag, nil
}

// GetTagsByOrganization retrieves all tags for an organization
func (s *Service) GetTagsByOrganization(organizationID uint) ([]database.Tag, error) {
	var tags []database.Tag

	if err := s.db.Where("organization_id = ?", organizationID).Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

// DeleteTag deletes a tag
func (s *Service) DeleteTag(id uint) error {
	result := s.db.Delete(&database.Tag{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTagNotFound
	}
	return nil
}

// AddTagToItem adds a tag to an item
func (s *Service) AddTagToItem(tagID uint, itemID uint) error {
	return s.db.Exec("INSERT INTO item_tags (tag_id, item_id) VALUES (?, ?) ON CONFLICT DO NOTHING", 
		tagID, itemID).Error
}

// RemoveTagFromItem removes a tag from an item
func (s *Service) RemoveTagFromItem(tagID uint, itemID uint) error {
	return s.db.Exec("DELETE FROM item_tags WHERE tag_id = ? AND item_id = ?", 
		tagID, itemID).Error
}

// GetTagsByItem retrieves all tags for an item
func (s *Service) GetTagsByItem(itemID uint) ([]database.Tag, error) {
	var tags []database.Tag

	if err := s.db.Joins("JOIN item_tags ON tags.id = item_tags.tag_id").
		Where("item_tags.item_id = ?", itemID).
		Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}
