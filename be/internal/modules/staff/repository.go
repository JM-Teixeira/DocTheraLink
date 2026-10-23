package staff

import (
	"context"
	"errors"

	"klyniva/internal/models"

	"gorm.io/gorm"
)

// Repository is the persistence boundary of the module.
type Repository interface {
	UserByUsername(ctx context.Context, username string) (*models.User, error)
	UserByID(ctx context.Context, id uint) (*models.User, error)
	UpdateUser(ctx context.Context, id uint, fields map[string]any) error
	Skills(ctx context.Context) ([]models.Skill, error)
	Skill(ctx context.Context, id string) (*models.Skill, error)
	UserSkills(ctx context.Context, userID uint) ([]models.UserSkill, error)
	AddUserSkill(ctx context.Context, userID uint, skillID string) (*models.UserSkill, error)
	RemoveUserSkill(ctx context.Context, userID uint, id string) error
	Clinic(ctx context.Context, id string) (*models.Facility, error)
	AddEmploymentEvent(ctx context.Context, e *EmploymentEvent) error
	// EmploymentEvents of a person, oldest first.
	EmploymentEvents(ctx context.Context, userID uint) ([]EmploymentEvent, error)
}

// GormRepository implements Repository on PostgreSQL via GORM.
type GormRepository struct{ db *gorm.DB }

// NewGormRepository builds a repository on the given connection.
func NewGormRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func notFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (g *GormRepository) UserByUsername(ctx context.Context, username string) (*models.User, error) {
	var u models.User
	err := g.db.WithContext(ctx).Preload("HomeFacility").Where("username = ?", username).First(&u).Error
	return &u, notFound(err)
}

func (g *GormRepository) UserByID(ctx context.Context, id uint) (*models.User, error) {
	var u models.User
	err := g.db.WithContext(ctx).Preload("HomeFacility").First(&u, id).Error
	return &u, notFound(err)
}

func (g *GormRepository) UpdateUser(ctx context.Context, id uint, fields map[string]any) error {
	return g.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(fields).Error
}

func (g *GormRepository) Skills(ctx context.Context) ([]models.Skill, error) {
	var out []models.Skill
	err := g.db.WithContext(ctx).Order("name").Find(&out).Error
	return out, err
}

func (g *GormRepository) Skill(ctx context.Context, id string) (*models.Skill, error) {
	var sk models.Skill
	err := g.db.WithContext(ctx).First(&sk, "id = ?", id).Error
	return &sk, notFound(err)
}

func (g *GormRepository) UserSkills(ctx context.Context, userID uint) ([]models.UserSkill, error) {
	var out []models.UserSkill
	err := g.db.WithContext(ctx).Preload("Skill").Where("user_id = ?", userID).Order("created_at").Find(&out).Error
	return out, err
}

func (g *GormRepository) AddUserSkill(ctx context.Context, userID uint, skillID string) (*models.UserSkill, error) {
	us := models.UserSkill{UserID: userID, SkillID: skillID}
	if err := g.db.WithContext(ctx).Omit("User", "Skill").Create(&us).Error; err != nil {
		return nil, err
	}
	err := g.db.WithContext(ctx).Preload("Skill").First(&us, "id = ?", us.ID).Error
	return &us, err
}

// RemoveUserSkill deletes the link by its own id or by the skill's id.
func (g *GormRepository) RemoveUserSkill(ctx context.Context, userID uint, id string) error {
	res := g.db.WithContext(ctx).Where("user_id = ? AND (id::text = ? OR skill_id::text = ?)", userID, id, id).
		Delete(&models.UserSkill{})
	if res.Error == nil && res.RowsAffected == 0 {
		return ErrNotFound
	}
	return res.Error
}

func (g *GormRepository) Clinic(ctx context.Context, id string) (*models.Facility, error) {
	var f models.Facility
	err := g.db.WithContext(ctx).First(&f, "id = ?", id).Error
	return &f, notFound(err)
}

func (g *GormRepository) AddEmploymentEvent(ctx context.Context, e *EmploymentEvent) error {
	return g.db.WithContext(ctx).Create(e).Error
}

func (g *GormRepository) EmploymentEvents(ctx context.Context, userID uint) ([]EmploymentEvent, error) {
	var out []EmploymentEvent
	err := g.db.WithContext(ctx).Where("user_id = ?", userID).Order("since, created_at").Find(&out).Error
	return out, err
}
