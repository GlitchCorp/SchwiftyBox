package organization

import (
	"errors"

	"backend/internal/database"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

// Module provides organization service dependency injection
var Module = fx.Module("organization",
	fx.Provide(NewOrganizationService),
)

// Service handles organization operations
type Service struct {
	db *gorm.DB
}

var (
	// ErrOrganizationNotFound is returned when organization is not found
	ErrOrganizationNotFound = errors.New("organization not found")
	// ErrOrganizationAlreadyExists is returned when trying to create an organization that already exists
	ErrOrganizationAlreadyExists = errors.New("organization already exists")
)

// NewOrganizationService creates a new organization service
func NewOrganizationService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// CreateOrganization creates a new organization
func (s *Service) CreateOrganization(name string) (*database.Organization, error) {
	organization := &database.Organization{
		Name: name,
	}

	if err := s.db.Create(organization).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrOrganizationAlreadyExists
		}
		return nil, err
	}

	return organization, nil
}

// GetOrganization retrieves an organization by ID
func (s *Service) GetOrganization(id uint) (*database.Organization, error) {
	var organization database.Organization

	if err := s.db.First(&organization, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrganizationNotFound
		}
		return nil, err
	}

	return &organization, nil
}

// GetOrganizationsByUser retrieves all organizations for a user
func (s *Service) GetOrganizationsByUser(userEmail string) ([]database.Organization, error) {
	var organizations []database.Organization

	if err := s.db.Joins("JOIN organization_users ON organizations.id = organization_users.organization_id").
		Where("organization_users.user_email = ?", userEmail).
		Find(&organizations).Error; err != nil {
		return nil, err
	}

	return organizations, nil
}

// AddUserToOrganization adds a user to an organization
func (s *Service) AddUserToOrganization(organizationID uint, userEmail string) error {
	// Check if organization exists
	if _, err := s.GetOrganization(organizationID); err != nil {
		return err
	}

	// Add user to organization using raw SQL to avoid GORM many-to-many complexity
	return s.db.Exec("INSERT INTO organization_users (organization_id, user_email) VALUES (?, ?) ON CONFLICT DO NOTHING",
		organizationID, userEmail).Error
}

// SetUserActiveOrganization sets the active organization for a user
func (s *Service) SetUserActiveOrganization(userEmail string, organizationID uint) error {
	return s.db.Model(&database.User{}).
		Where("email = ?", userEmail).
		Update("active_organization_id", organizationID).Error
}
