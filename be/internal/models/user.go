package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	RolePharmacist           UserRole = "Pharmacist"
	RoleVolunteer            UserRole = "Volunteer"
	RoleStaffReadOnly        UserRole = "StaffReadOnly"
	RoleStaff                UserRole = "Staff"
	RoleNurse                UserRole = "Nurse"
	RolePhysioTherapist      UserRole = "PhysioTherapist"
	RoleDoctor               UserRole = "Doctor"
	RoleWardAdmin            UserRole = "WardAdmin"
	RoleLocalBodyAdmin       UserRole = "LocalBodyAdmin"
	RoleDistrictLabAdmin     UserRole = "DistrictLabAdmin"
	RoleDistrictReadOnlyAdmin UserRole = "DistrictReadOnlyAdmin"
	RoleDistrictAdmin        UserRole = "DistrictAdmin"
	RoleStateLabAdmin        UserRole = "StateLabAdmin"
	RoleStateReadOnlyAdmin   UserRole = "StateReadOnlyAdmin"
	RoleStateAdmin           UserRole = "StateAdmin"
)

type GenderType string

const (
	GenderMale   GenderType = "Male"
	GenderFemale GenderType = "Female"
	GenderOther  GenderType = "Other"
)

type User struct {
	ID                               uint      `json:"id" gorm:"primaryKey"`
	Username                         string    `json:"username" gorm:"uniqueIndex;not null"`
	Email                            string    `json:"email" gorm:"uniqueIndex;not null"`
	FirstName                        string    `json:"first_name" gorm:"not null"`
	LastName                         string    `json:"last_name" gorm:"not null"`
	Password                         string    `json:"-" gorm:"not null"`
	UserType                         UserRole  `json:"user_type" gorm:"not null"`
	PhoneNumber                      *string   `json:"phone_number"`
	AltPhoneNumber                   *string   `json:"alt_phone_number"`
	Gender                           *GenderType `json:"gender"`
	DateOfBirth                      *time.Time `json:"date_of_birth"`
	IsActive                         bool      `json:"is_active" gorm:"default:true"`
	IsSuperuser                      bool      `json:"is_superuser" gorm:"default:false"`
	IsVerified                       bool      `json:"verified" gorm:"default:false"`
	LastLogin                        *time.Time `json:"last_login"`
	VideoConnectLink                 string    `json:"video_connect_link"`
	ProfilePictureURL                *string   `json:"read_profile_picture_url"`
	Qualification                    *string   `json:"qualification"`
	DoctorExperienceCommencedOn      *time.Time `json:"doctor_experience_commenced_on"`
	DoctorMedicalCouncilRegistration *string   `json:"doctor_medical_council_registration"`
	WeeklyWorkingHours               *string   `json:"weekly_working_hours"`
	
	// Foreign keys
	LocalBodyID  *uint `json:"local_body"`
	DistrictID   *uint `json:"district"`
	StateID      *uint `json:"state"`
	HomeFacilityID *string `json:"home_facility"`

	// Relationships
	LocalBody    *LocalBody `json:"local_body_object" gorm:"foreignKey:LocalBodyID"`
	District     *District  `json:"district_object" gorm:"foreignKey:DistrictID"`
	State        *State     `json:"state_object" gorm:"foreignKey:StateID"`
	HomeFacility *Facility  `json:"home_facility_object" gorm:"foreignKey:HomeFacilityID"`
	Skills       []UserSkill `json:"skills" gorm:"foreignKey:UserID"`

	CreatedAt time.Time      `json:"created_date"`
	UpdatedAt time.Time      `json:"modified_date"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type UserSkill struct {
	ID     string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID uint   `json:"-"`
	SkillID string `json:"-"`
	
	User  User  `json:"-" gorm:"foreignKey:UserID"`
	Skill Skill `json:"skill_object" gorm:"foreignKey:SkillID"`
	
	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

type Skill struct {
	ID          string  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string  `json:"name" gorm:"not null"`
	Description *string `json:"description"`
	
	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

// UserBareMinimum represents the minimal user information
type UserBareMinimum struct {
	ID                    uint      `json:"id"`
	Username              string    `json:"username"`
	FirstName             string    `json:"first_name"`
	LastName              string    `json:"last_name"`
	Email                 string    `json:"email"`
	UserType              UserRole  `json:"user_type"`
	LastLogin             *time.Time `json:"last_login"`
	ProfilePictureURL     *string   `json:"read_profile_picture_url"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

// TokenRefreshRequest represents token refresh request
type TokenRefreshRequest struct {
	Refresh string `json:"refresh" binding:"required"`
}

// UpdatePasswordRequest represents password update request
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	Username    string `json:"username" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// PasswordResetRequest represents password reset request
type PasswordResetRequest struct {
	Username string `json:"username" binding:"required"`
}

// PasswordResetConfirmRequest represents password reset confirmation
type PasswordResetConfirmRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
	Confirm  string `json:"confirm" binding:"required"`
}

// BeforeCreate hook to set UUID for UserSkill
func (us *UserSkill) BeforeCreate(tx *gorm.DB) error {
	if us.ID == "" {
		us.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate hook to set UUID for Skill
func (s *Skill) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}