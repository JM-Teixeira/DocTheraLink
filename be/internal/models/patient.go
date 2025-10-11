package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PatientCategory string

const (
	PatientCategoryMild           PatientCategory = "Mild"
	PatientCategoryModerate       PatientCategory = "Moderate"
	PatientCategoryCritical       PatientCategory = "Critical"
	PatientCategoryActivelyDying  PatientCategory = "Actively Dying"
)

type BloodGroup string

const (
	BloodGroupAPositive  BloodGroup = "A+"
	BloodGroupANegative  BloodGroup = "A-"
	BloodGroupBPositive  BloodGroup = "B+"
	BloodGroupBNegative  BloodGroup = "B-"
	BloodGroupABPositive BloodGroup = "AB+"
	BloodGroupABNegative BloodGroup = "AB-"
	BloodGroupOPositive  BloodGroup = "O+"
	BloodGroupONegative  BloodGroup = "O-"
)

type RationCardCategory string

const (
	RationCardCategoryBPL RationCardCategory = "BPL"
	RationCardCategoryAPL RationCardCategory = "APL"
	RationCardCategoryNone RationCardCategory = "NONE"
)

type Patient struct {
	tableName struct{} `gorm:"table:patients"`
	ID                                    string              `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name                                  string              `json:"name" gorm:"not null"`
	Gender                                *int                `json:"gender"`
	PhoneNumber                           *string             `json:"phone_number"`
	EmergencyPhoneNumber                  *string             `json:"emergency_phone_number"`
	DateOfBirth                           *time.Time          `json:"date_of_birth"`
	YearOfBirth                           *int                `json:"year_of_birth"`
	Age                                   *string             `json:"age" gorm:"-"`
	BloodGroup                            *BloodGroup         `json:"blood_group"`
	Address                               *string             `json:"address"`
	PermanentAddress                      *string             `json:"permanent_address"`
	Village                               *string             `json:"village"`
	Pincode                               *int                `json:"pincode"`
	Nationality                           *string             `json:"nationality"`
	PassportNo                            *string             `json:"passport_no"`
	RationCardCategory                    *RationCardCategory `json:"ration_card_category"`
	
	// Medical Information
	Allergies                             *string             `json:"allergies"`
	MedicalHistoryDetails                 *string             `json:"medical_history_details"`
	OngoingMedication                     *string             `json:"ongoing_medication"`
	PresentHealth                         *string             `json:"present_health"`
	
	// COVID-related fields
	ContactWithCarrier                    *bool               `json:"contact_with_carrier"`
	HasSARI                               *bool               `json:"has_SARI"`
	PastTravel                            *bool               `json:"past_travel"`
	CountriesTravelled                    *string             `json:"countries_travelled"` // JSON array as string
	TransitDetails                        *string             `json:"transit_details"`
	DateOfTest                            *time.Time          `json:"date_of_test"`
	DateDeclaredPositive                  *time.Time          `json:"date_declared_positive"`
	IsDeclaredPositive                    *bool               `json:"is_declared_positive"`
	CovinID                               *string             `json:"covin_id"`
	
	// Vaccination
	IsVaccinated                          *bool               `json:"is_vaccinated"`
	VaccineName                           *string             `json:"vaccine_name"`
	NumberOfDoses                         *int                `json:"number_of_doses"`
	LastVaccinatedDate                    *time.Time          `json:"last_vaccinated_date"`
	
	// Healthcare worker information
	IsMedicalWorker                       *bool               `json:"is_medical_worker"`
	DesignationOfHealthCareWorker         *string             `json:"designation_of_health_care_worker"`
	InstitutionOfHealthCareWorker         *string             `json:"instituion_of_health_care_worker"`
	FrontlineWorker                       *string             `json:"frontline_worker"`
	
	// Pregnancy related
	IsAntenatal                           *bool               `json:"is_antenatal"`
	LastMenstruationStartDate             *time.Time          `json:"last_menstruation_start_date"`
	DateOfDelivery                        *time.Time          `json:"date_of_delivery"`
	
	// Other information
	IsMigrantWorker                       *bool               `json:"is_migrant_worker"`
	NumberOfAgedDependents                *int                `json:"number_of_aged_dependents"`
	NumberOfChronicDiseasedDependents     *int                `json:"number_of_chronic_diseased_dependents"`
	WillDonateBlood                       *bool               `json:"will_donate_blood"`
	FitForBloodDonation                   *bool               `json:"fit_for_blood_donation"`
	DateOfReturn                          *time.Time          `json:"date_of_return"`
	
	// Status fields
	IsActive                              bool                `json:"is_active" gorm:"default:true"`
	AllowTransfer                         *bool               `json:"allow_transfer"`
	ReviewInterval                        *int                `json:"review_interval"`
	ReviewTime                            *time.Time          `json:"review_time"`
	DeathDateTime                         *time.Time          `json:"death_datetime"`
	
	// Foreign keys
	FacilityID   *string `json:"facility"`
	StateID      *uint   `json:"state"`
	DistrictID   *uint   `json:"district"`
	LocalBodyID  *uint   `json:"local_body"`
	WardID       *uint   `json:"ward"`
	AssignedToID *uint   `json:"assigned_to"`
	CreatedByID  *uint   `json:"-"`
	LastEditedID *uint   `json:"-"`
	
	// Relationships
	Facility     *Facility        `json:"facility_object" gorm:"foreignKey:FacilityID"`
	State        *State           `json:"state_object" gorm:"foreignKey:StateID"`
	District     *District        `json:"district_object" gorm:"foreignKey:DistrictID"`
	LocalBody    *LocalBody       `json:"local_body_object" gorm:"foreignKey:LocalBodyID"`
	Ward         *Ward            `json:"ward_object" gorm:"foreignKey:WardID"`
	AssignedTo   *User            `json:"assigned_to_object" gorm:"foreignKey:AssignedToID"`
	CreatedBy    *UserBareMinimum `json:"created_by" gorm:"foreignKey:CreatedByID"`
	LastEdited   *UserBareMinimum `json:"last_edited" gorm:"foreignKey:LastEditedID"`
	
	CreatedAt time.Time      `json:"created_date"`
	UpdatedAt time.Time      `json:"modified_date"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// PatientNote model temporarily moved to notes_disabled.go due to GORM constraint conflicts

type PatientTransfer struct {
	ID                string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PatientID         string    `json:"-"`
	FromFacilityID    string    `json:"-"`
	ToFacilityID      string    `json:"-"`
	TransferReason    string    `json:"transfer_reason"`
	TransferNotes     *string   `json:"transfer_notes"`
	InitiatedByID     uint      `json:"-"`
	
	Patient      Patient  `json:"patient" gorm:"foreignKey:PatientID"`
	FromFacility Facility `json:"from_facility" gorm:"foreignKey:FromFacilityID"`
	ToFacility   Facility `json:"to_facility" gorm:"foreignKey:ToFacilityID"`
	InitiatedBy  User     `json:"initiated_by" gorm:"foreignKey:InitiatedByID"`
	
	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

// Request/Response types
type PatientTransferRequest struct {
	FacilityID     string  `json:"facility" binding:"required"`
	TransferReason string  `json:"transfer_reason" binding:"required"`
	TransferNotes  *string `json:"transfer_notes"`
}

// PatientNoteRequest temporarily disabled - see notes_disabled.go

// BeforeCreate hooks
func (p *Patient) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// PatientNote BeforeCreate hook temporarily disabled - see notes_disabled.go

func (pt *PatientTransfer) BeforeCreate(tx *gorm.DB) error {
	if pt.ID == "" {
		pt.ID = uuid.New().String()
	}
	return nil
}