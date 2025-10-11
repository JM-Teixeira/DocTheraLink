package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConsultationSuggestion string

const (
	ConsultationSuggestionHI ConsultationSuggestion = "HI"  // Home Isolation
	ConsultationSuggestionA  ConsultationSuggestion = "A"   // Admission
	ConsultationSuggestionR  ConsultationSuggestion = "R"   // Refer to another facility
	ConsultationSuggestionOP ConsultationSuggestion = "OP"  // OP Consultation
	ConsultationSuggestionDC ConsultationSuggestion = "DC"  // Domiciliary Care
	ConsultationSuggestionDD ConsultationSuggestion = "DD"  // Declare Death
)

type DischargeReason string

const (
	DischargeReasonRecovered        DischargeReason = "REC"
	DischargeReasonReferral         DischargeReason = "REF"
	DischargeReasonExpired          DischargeReason = "EXP"
	DischargeReasonLAMA             DischargeReason = "LAMA"
	DischargeReasonUnknown          DischargeReason = "UNKNOWN"
)

type Consultation struct {
	ID                        string                  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PatientID                 string                  `json:"-"`
	FacilityID                string                  `json:"facility"`
	EncounterDate             time.Time               `json:"encounter_date" gorm:"not null"`
	ICUAdmissionDate          *time.Time              `json:"icu_admission_date"`
	Admitted                  *bool                   `json:"admitted"`
	AdmittedTo                *string                 `json:"admitted_to"`
	Category                  *PatientCategory        `json:"category"`
	ExaminationDetails        *string                 `json:"examination_details"`
	HistoryOfPresentIllness   *string                 `json:"history_of_present_illness"`
	TreatmentPlan             *string                 `json:"treatment_plan"`
	ConsultationNotes         *string                 `json:"consultation_notes"`
	CourseInFacility          *string                 `json:"course_in_facility"`
	Suggestion                *ConsultationSuggestion `json:"suggestion"`
	ReferredTo                *string                 `json:"referred_to"`
	ReferredToExternal        *string                 `json:"referred_to_external"`
	TransferredFromLocation   *string                 `json:"transferred_from_location"`
	DischargeDate             *time.Time              `json:"discharge_date"`
	NewDischargeReason        *DischargeReason        `json:"new_discharge_reason"`
	DischargeNotes            *string                 `json:"discharge_notes"`
	DeathDateTime             *time.Time              `json:"death_datetime"`
	DeathConfirmedDoctor      *string                 `json:"death_confirmed_doctor"`
	
	// Vitals and monitoring
	Temperature               *float64                `json:"temperature"`
	Temperature_measured_at   *time.Time              `json:"temperature_measured_at"`
	SpO2                      *int                    `json:"spo2"`
	VentilatorInterfaceType   *string                 `json:"ventilator_interface_type"`
	
	// Foreign keys
	ConsultingDoctorID        *uint                   `json:"-"`
	CreatedByID               uint                    `json:"-"`
	LastEditedByID            *uint                   `json:"-"`
	
	// Relationships
	Patient           Patient          `json:"patient" gorm:"foreignKey:PatientID"`
	Facility          Facility         `json:"facility_object" gorm:"foreignKey:FacilityID"`
	ConsultingDoctor  *User            `json:"consulting_doctor" gorm:"foreignKey:ConsultingDoctorID"`
	CreatedBy         UserBareMinimum  `json:"created_by" gorm:"foreignKey:CreatedByID"`
	LastEditedBy      *UserBareMinimum `json:"last_edited_by" gorm:"foreignKey:LastEditedByID"`
	DailyRounds       []DailyRound     `json:"daily_rounds" gorm:"foreignKey:ConsultationID"`
	
	CreatedAt time.Time      `json:"created_date"`
	UpdatedAt time.Time      `json:"modified_date"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type DailyRoundType string

const (
	DailyRoundTypeNormal              DailyRoundType = "NORMAL"
	DailyRoundTypeCommunityNursesLog  DailyRoundType = "COMMUNITY_NURSES_LOG"
	DailyRoundTypeDoctorsLog          DailyRoundType = "DOCTORS_LOG"
	DailyRoundTypeVentilator          DailyRoundType = "VENTILATOR"
)

type DailyRound struct {
	ID                        string          `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ConsultationID            string          `json:"-"`
	RoundsType                DailyRoundType  `json:"rounds_type" gorm:"default:'NORMAL'"`
	PatientCategory           *PatientCategory `json:"patient_category"`
	PhysicalExaminationInfo   *string         `json:"physical_examination_info"`
	OtherDetails              *string         `json:"other_details"`
	Recommendation            *string         `json:"recommendation"`
	ActionTaken               *string         `json:"action_taken"`
	ReviewInterval            *int            `json:"review_interval"`
	ICUAdmissionDate          *time.Time      `json:"icu_admission_date"`
	
	// Vitals
	Temperature               *float64        `json:"temperature"`
	Temperature_measured_at   *time.Time      `json:"temperature_measured_at"`
	SpO2                      *int            `json:"spo2"`
	VentilatorSpO2            *int            `json:"ventilator_spo2"`
	VentilatorFiO2            *int            `json:"ventilator_fio2"`
	VentilatorOxygenModality  *string         `json:"ventilator_oxygen_modality"`
	VentilatorOxygenModalityOxygenRate *float64 `json:"ventilator_oxygen_modality_oxygen_rate"`
	VentilatorOxygenModalityFlowRate   *float64 `json:"ventilator_oxygen_modality_flow_rate"`
	VentilatorInterface       *string         `json:"ventilator_interface"`
	VentilatorMode            *string         `json:"ventilator_mode"`
	VentilatorPEEP            *float64        `json:"ventilator_peep"`
	VentilatorPIP             *float64        `json:"ventilator_pip"`
	VentilatorMeanAirwayPressure *float64     `json:"ventilator_mean_airway_pressure"`
	VentilatorRespRate        *int            `json:"ventilator_resp_rate"`
	VentilatorPressureSupport *float64        `json:"ventilator_pressure_support"`
	VentilatorTidalVolume     *float64        `json:"ventilator_tidal_volume"`
	
	// Blood Pressure
	BP                        *string         `json:"bp"` // JSON string for systolic/diastolic
	PulseRate                 *int            `json:"pulse"`
	RespiratoryRate           *int            `json:"resp"`
	
	// Consciousness
	ConsciousnessLevel        *string         `json:"consciousness_level"`
	LeftPupilLightReaction    *string         `json:"left_pupil_light_reaction"`
	RightPupilLightReaction   *string         `json:"right_pupil_light_reaction"`
	LeftPupilSize             *int            `json:"left_pupil_size"`
	RightPupilSize            *int            `json:"right_pupil_size"`
	LimbResponse              *string         `json:"limb_response"`
	GlasgowEyeOpen            *int            `json:"glasgow_eye_open"`
	GlasgowVerbalResponse     *int            `json:"glasgow_verbal_response"`
	GlasgowMotorResponse      *int            `json:"glasgow_motor_response"`
	GlasgowTotalCalculated    *int            `json:"glasgow_total_calculated"`
	
	// Additional fields
	RhythmType                *string         `json:"rhythm"`
	RhythmDetail              *string         `json:"rhythm_detail"`
	
	// Foreign keys
	TakenByID                 uint            `json:"-"`
	
	// Relationships
	Consultation Consultation    `json:"-" gorm:"foreignKey:ConsultationID"`
	TakenBy      UserBareMinimum `json:"taken_by" gorm:"foreignKey:TakenByID"`
	
	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

type ConsultationBed struct {
	ID             string       `json:"external_id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ConsultationID string       `json:"-"`
	BedID          string       `json:"-"`
	StartDate      time.Time    `json:"start_date" gorm:"not null"`
	EndDate        *time.Time   `json:"end_date"`
	
	Consultation Consultation `json:"-" gorm:"foreignKey:ConsultationID"`
	Bed          Bed          `json:"bed" gorm:"foreignKey:BedID"`
	
	CreatedAt time.Time `json:"created_date"`
	UpdatedAt time.Time `json:"modified_date"`
}

// Request types
type ConsultationRequest struct {
	PatientID                 string                  `json:"patient" binding:"required"`
	FacilityID                string                  `json:"facility" binding:"required"`
	EncounterDate             time.Time               `json:"encounter_date" binding:"required"`
	ICUAdmissionDate          *time.Time              `json:"icu_admission_date"`
	Admitted                  *bool                   `json:"admitted"`
	AdmittedTo                *string                 `json:"admitted_to"`
	Category                  *PatientCategory        `json:"category"`
	ExaminationDetails        *string                 `json:"examination_details"`
	HistoryOfPresentIllness   *string                 `json:"history_of_present_illness"`
	TreatmentPlan             *string                 `json:"treatment_plan"`
	ConsultationNotes         *string                 `json:"consultation_notes"`
	Suggestion                *ConsultationSuggestion `json:"suggestion"`
	ReferredTo                *string                 `json:"referred_to"`
	ReferredToExternal        *string                 `json:"referred_to_external"`
	ConsultingDoctorID        *uint                   `json:"consulting_doctor"`
}

type DailyRoundRequest struct {
	RoundsType                DailyRoundType   `json:"rounds_type"`
	PatientCategory           *PatientCategory `json:"patient_category"`
	PhysicalExaminationInfo   *string          `json:"physical_examination_info"`
	OtherDetails              *string          `json:"other_details"`
	Recommendation            *string          `json:"recommendation"`
	ActionTaken               *string          `json:"action_taken"`
	ReviewInterval            *int             `json:"review_interval"`
	Temperature               *float64         `json:"temperature"`
	SpO2                      *int             `json:"spo2"`
	BP                        *string          `json:"bp"`
	PulseRate                 *int             `json:"pulse"`
	RespiratoryRate           *int             `json:"resp"`
	ConsciousnessLevel        *string          `json:"consciousness_level"`
}

// BeforeCreate hooks
func (c *Consultation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

func (dr *DailyRound) BeforeCreate(tx *gorm.DB) error {
	if dr.ID == "" {
		dr.ID = uuid.New().String()
	}
	return nil
}

func (cb *ConsultationBed) BeforeCreate(tx *gorm.DB) error {
	if cb.ID == "" {
		cb.ID = uuid.New().String()
	}
	return nil
}