package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PatientCategory string

const (
	PatientCategoryMild          PatientCategory = "Mild"
	PatientCategoryModerate      PatientCategory = "Moderate"
	PatientCategoryCritical      PatientCategory = "Critical"
	PatientCategoryActivelyDying PatientCategory = "Actively Dying"
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
	RationCardCategoryBPL  RationCardCategory = "BPL"
	RationCardCategoryAPL  RationCardCategory = "APL"
	RationCardCategoryNone RationCardCategory = "NONE"
)

type Patient struct {
	tableName            struct{} `gorm:"table:patients"`
	ID                   string   `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name                 string   `json:"name" gorm:"not null"`
	Gender               *int     `json:"gender"`
	PhoneNumber          *string  `json:"phone_number"`
	EmergencyPhoneNumber *string  `json:"emergency_phone_number"`
	// Who answers the emergency phone and how they relate to the patient
	EmergencyContactName     *string     `json:"emergency_contact_name" gorm:"size:120"`
	EmergencyContactRelation *string     `json:"emergency_contact_relation" gorm:"size:20"`
	Email                    *string     `json:"email" gorm:"size:254"`
	DateOfBirth              *time.Time  `json:"date_of_birth"`
	YearOfBirth              *int        `json:"year_of_birth"`
	Age                      *string     `json:"age" gorm:"-"`
	BloodGroup               *BloodGroup `json:"blood_group"`
	Address                  *string     `json:"address"`
	PermanentAddress         *string     `json:"permanent_address"`
	Village                  *string     `json:"village"`
	Pincode                  *int        `json:"pincode"`
	// Address parts: Address is the street line
	PostalCode *string `json:"postal_code" gorm:"size:8"` // Portuguese 0000-000
	City       *string `json:"city" gorm:"size:100"`
	// Health insurance or subsystem (ADSE, Médis, ...) and the member number
	Insurer         *string `json:"insurer" gorm:"size:100"`
	InsuranceNumber *string `json:"insurance_number" gorm:"size:40"`
	Nationality     *string `json:"nationality"`
	NIF             *string `json:"nif" gorm:"size:9;index"`        // Portuguese tax number (validated)
	SNSNumber       *string `json:"sns_number" gorm:"size:9;index"` // número de utente do SNS
	// CareLevel 1–5: how much of a physio's hour the patient needs (1 =
	// independent … 5 = cannot move by themselves); booking credits.
	CareLevel          int                 `json:"care_level" gorm:"not null;default:1"`
	PassportNo         *string             `json:"passport_no"`
	RationCardCategory *RationCardCategory `json:"ration_card_category"`

	// Medical Information
	Allergies             *string `json:"allergies"`
	MedicalHistoryDetails *string `json:"medical_history_details"`
	OngoingMedication     *string `json:"ongoing_medication"`
	PresentHealth         *string `json:"present_health"`

	// COVID-related fields
	ContactWithCarrier   *bool      `json:"contact_with_carrier"`
	HasSARI              *bool      `json:"has_SARI"`
	PastTravel           *bool      `json:"past_travel"`
	CountriesTravelled   *string    `json:"countries_travelled"` // JSON array as string
	TransitDetails       *string    `json:"transit_details"`
	DateOfTest           *time.Time `json:"date_of_test"`
	DateDeclaredPositive *time.Time `json:"date_declared_positive"`
	IsDeclaredPositive   *bool      `json:"is_declared_positive"`
	CovinID              *string    `json:"covin_id"`

	// Vaccination
	IsVaccinated       *bool      `json:"is_vaccinated"`
	VaccineName        *string    `json:"vaccine_name"`
	NumberOfDoses      *int       `json:"number_of_doses"`
	LastVaccinatedDate *time.Time `json:"last_vaccinated_date"`

	// Healthcare worker information
	IsMedicalWorker               *bool   `json:"is_medical_worker"`
	DesignationOfHealthCareWorker *string `json:"designation_of_health_care_worker"`
	InstitutionOfHealthCareWorker *string `json:"instituion_of_health_care_worker"`
	FrontlineWorker               *string `json:"frontline_worker"`

	// Pregnancy related
	IsAntenatal               *bool      `json:"is_antenatal"`
	LastMenstruationStartDate *time.Time `json:"last_menstruation_start_date"`
	DateOfDelivery            *time.Time `json:"date_of_delivery"`

	// Other information
	IsMigrantWorker                   *bool      `json:"is_migrant_worker"`
	NumberOfAgedDependents            *int       `json:"number_of_aged_dependents"`
	NumberOfChronicDiseasedDependents *int       `json:"number_of_chronic_diseased_dependents"`
	WillDonateBlood                   *bool      `json:"will_donate_blood"`
	FitForBloodDonation               *bool      `json:"fit_for_blood_donation"`
	DateOfReturn                      *time.Time `json:"date_of_return"`

	// Status fields
	IsActive       bool       `json:"is_active" gorm:"default:true"`
	AllowTransfer  *bool      `json:"allow_transfer"`
	ReviewInterval *int       `json:"review_interval"`
	ReviewTime     *time.Time `json:"review_time"`
	DeathDateTime  *time.Time `json:"death_datetime"`

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
	Facility   *Facility        `json:"facility_object" gorm:"foreignKey:FacilityID"`
	State      *State           `json:"state_object" gorm:"foreignKey:StateID"`
	District   *District        `json:"district_object" gorm:"foreignKey:DistrictID"`
	LocalBody  *LocalBody       `json:"local_body_object" gorm:"foreignKey:LocalBodyID"`
	Ward       *Ward            `json:"ward_object" gorm:"foreignKey:WardID"`
	AssignedTo *User            `json:"assigned_to_object" gorm:"foreignKey:AssignedToID"`
	CreatedBy  *UserBareMinimum `json:"created_by" gorm:"foreignKey:CreatedByID"`
	LastEdited *UserBareMinimum `json:"last_edited" gorm:"foreignKey:LastEditedID"`

	CreatedAt time.Time      `json:"created_date"`
	UpdatedAt time.Time      `json:"modified_date"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate assigns a UUID so the ID is known before the insert returns.
func (p *Patient) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
