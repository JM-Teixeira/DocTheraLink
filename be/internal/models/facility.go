package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FacilityType int

const (
	FacilityTypePrimaryHealthCentres        FacilityType = 800
	FacilityTypeFamilyHealthCentres         FacilityType = 802
	FacilityTypeCommunityHealthCentres      FacilityType = 803
	FacilityTypeWomenChildHealthCentres     FacilityType = 840
	FacilityTypeTalukHospitals              FacilityType = 830
	FacilityTypeDistrictHospitals           FacilityType = 860
	FacilityTypeGovtMedicalCollegeHospitals FacilityType = 870
	FacilityTypePrivateHospital             FacilityType = 2
	FacilityTypeClinic                      FacilityType = 2000
	FacilityTypePhysiotherapyClinic         FacilityType = 2010
	FacilityTypeDentalClinic                FacilityType = 2020
	FacilityTypeRehabilitationCenter        FacilityType = 2030
	FacilityTypeNursingHome                 FacilityType = 2040
	FacilityTypeGovtLabs                    FacilityType = 9
	FacilityTypePrivateLabs                 FacilityType = 10
	FacilityTypeTeleMedicine                FacilityType = 7
	FacilityTypeAutonomousHealthcare        FacilityType = 910
	FacilityTypeShiftingCentre              FacilityType = 1300
	FacilityTypeRequestApprovingCenter      FacilityType = 1500
	FacilityTypeRequestFulfilmentCenter     FacilityType = 1510
	FacilityTypeOther                       FacilityType = 3
)

type SpokeRelationship int

const (
	SpokeRelationshipRegular SpokeRelationship = 1
	SpokeRelationshipTeleICU SpokeRelationship = 2
)

type State struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"not null"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

type District struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Name    string `json:"name" gorm:"not null"`
	StateID uint   `json:"state"`

	State State `json:"state_object" gorm:"foreignKey:StateID"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

type LocalBody struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	Name          string `json:"name" gorm:"not null"`
	BodyType      int    `json:"body_type"`
	LocalBodyCode string `json:"localbody_code"`
	DistrictID    uint   `json:"district"`

	District District `json:"district_object" gorm:"foreignKey:DistrictID"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

type Ward struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"not null"`
	Number      int    `json:"number"`
	Panchayath  string `json:"panchayath"`
	LocalBodyID uint   `json:"local_body_id"`

	LocalBody LocalBody `json:"local_body_object" gorm:"foreignKey:LocalBodyID"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

type Facility struct {
	ID                string       `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name              string       `json:"name" gorm:"not null"`
	FacilityType      FacilityType `json:"facility_type"`
	Address           string       `json:"address"`
	MiddlewareAddress *string      `json:"middleware_address"`
	Pincode           *string      `json:"pincode"`
	Latitude          *float64     `json:"latitude"`
	Longitude         *float64     `json:"longitude"`
	CoverImageURL     *string      `json:"read_cover_image_url"`
	KaspEmpanelled    bool         `json:"kasp_empanelled" gorm:"default:false"`
	Features          []int        `json:"features" gorm:"type:integer[]"`

	// Foreign keys
	StateID    *uint `json:"state"`
	DistrictID *uint `json:"district"`

	// Relationships
	State    *State    `json:"state_object" gorm:"foreignKey:StateID"`
	District *District `json:"district_object" gorm:"foreignKey:DistrictID"`

	// Computed fields
	PatientCount *int `json:"patient_count" gorm:"-"`
	BedCount     *int `json:"bed_count" gorm:"-"`

	CreatedAt time.Time      `json:"created_date"`
	UpdatedAt time.Time      `json:"modified_date"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type FacilitySpoke struct {
	ID           string            `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	HubID        string            `json:"-"`
	SpokeID      string            `json:"-"`
	Relationship SpokeRelationship `json:"relationship"`

	Hub   Facility `json:"hub_object" gorm:"foreignKey:HubID"`
	Spoke Facility `json:"spoke_object" gorm:"foreignKey:SpokeID"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

type Location struct {
	ID          string  `json:"external_id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string  `json:"name" gorm:"not null"`
	Description *string `json:"description"`
	FacilityID  string  `json:"-"`

	Facility Facility `json:"-" gorm:"foreignKey:FacilityID"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

type Bed struct {
	ID          string  `json:"external_id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string  `json:"name" gorm:"not null"`
	Description *string `json:"description"`
	BedType     string  `json:"bed_type"`
	FacilityID  string  `json:"-"`
	LocationID  *string `json:"-"`

	Facility Facility  `json:"-" gorm:"foreignKey:FacilityID"`
	Location *Location `json:"location_object" gorm:"foreignKey:LocationID"`

	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

// FacilityRequest represents facility creation/update request
type FacilityRequest struct {
	Name              string       `json:"name" binding:"required"`
	FacilityType      FacilityType `json:"facility_type" binding:"required"`
	Address           string       `json:"address" binding:"required"`
	MiddlewareAddress *string      `json:"middleware_address"`
	Pincode           *string      `json:"pincode"`
	Latitude          *float64     `json:"latitude"`
	Longitude         *float64     `json:"longitude"`
	StateID           *uint        `json:"state"`
	DistrictID        *uint        `json:"district"`
	KaspEmpanelled    bool         `json:"kasp_empanelled"`
	Features          []int        `json:"features"`
}

// FacilitySpokeRequest represents facility spoke creation/update request
type FacilitySpokeRequest struct {
	SpokeID      *string            `json:"spoke"`
	Relationship *SpokeRelationship `json:"relationship"`
}

// BeforeCreate hook to set UUID
func (f *Facility) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	return nil
}

func (fs *FacilitySpoke) BeforeCreate(tx *gorm.DB) error {
	if fs.ID == "" {
		fs.ID = uuid.New().String()
	}
	return nil
}

func (l *Location) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}

func (b *Bed) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
