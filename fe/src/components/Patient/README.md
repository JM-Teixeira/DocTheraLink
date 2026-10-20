# Patient Creation Feature

This directory contains the implementation for the centralized patient creation feature in the CARE healthcare management platform.

## Directory Structure

```
Patient/
├── types/                      # TypeScript type definitions
│   ├── patientCreate.ts       # Core types for patient creation
│   └── index.ts               # Type exports
├── validation/                 # Validation schemas and functions
│   ├── patientValidation.ts   # Zod schemas and validation logic
│   └── index.ts               # Validation exports
├── constants/                  # Constants and configuration
│   ├── patientConstants.ts    # Enums, options, and defaults
│   └── index.ts               # Constants exports
└── README.md                   # This file
```

## Type Definitions

### Core Types

- **Demographics**: Patient personal information (name, DOB, gender, contact info, address)
- **EmergencyContact**: Emergency contact details
- **MedicalHistory**: Allergies, medications, and medical conditions
- **ClinicalNotes**: Reason for visit and initial clinical notes
- **Administrative**: Facility assignment, status, insurance information
- **PatientFormState**: Complete form state combining all sections

### Enums

- **PatientStatus**: `LIVE`, `DISCHARGED`
- **PaymentMethod**: `INSURANCE`, `SELF_PAY`, `GOVERNMENT_SCHEME`
- **AllergySeverity**: `MILD`, `MODERATE`, `SEVERE`, `LIFE_THREATENING`
- **ConditionStatus**: `ACTIVE`, `RESOLVED`, `CHRONIC`

## Validation

The validation layer uses **Zod** for runtime type checking and validation.

### Validation Schemas

- `demographicsSchema`: Validates patient demographics
- `emergencyContactSchema`: Validates emergency contact (optional)
- `medicalHistorySchema`: Validates medical history
- `clinicalNotesSchema`: Validates clinical notes
- `administrativeSchema`: Validates administrative data
- `patientFormSchema`: Complete form validation

### Validation Rules

#### Name Fields
- Min length: 2 characters
- Max length: 50 characters
- Pattern: Letters, spaces, hyphens, apostrophes only

#### Date of Birth
- Cannot be in the future
- Must be after 1900-01-01

#### Phone Numbers
- Pattern: International format with optional country code
- Example: `+1234567890`, `(123) 456-7890`

#### Email
- Standard email format validation

#### Reason for Visit
- Min length: 10 characters
- Max length: 1000 characters

#### Discharge Date
- Required when status is "Discharged"
- Cannot be before admission date

### Helper Functions

```typescript
// Validate a single field
validateField(schema, value) => { isValid, error }

// Validate entire form
validatePatientForm(data) => { isValid, errors }

// Calculate age from date of birth
calculateAge(dateOfBirth) => number
```

## Constants

### Options Arrays

- `PATIENT_STATUS_OPTIONS`: Status dropdown options
- `PAYMENT_METHOD_OPTIONS`: Payment method options
- `ALLERGY_SEVERITY_OPTIONS`: Allergy severity levels
- `CONDITION_STATUS_OPTIONS`: Medical condition status
- `RELATIONSHIP_OPTIONS`: Emergency contact relationships

### Field Limits

```typescript
FIELD_LIMITS = {
  NAME_MIN_LENGTH: 2,
  NAME_MAX_LENGTH: 50,
  REASON_FOR_VISIT_MIN_LENGTH: 10,
  REASON_FOR_VISIT_MAX_LENGTH: 1000,
  INITIAL_NOTES_MAX_LENGTH: 5000,
}
```

### Auto-save Configuration

```typescript
AUTO_SAVE_CONFIG = {
  INTERVAL_MS: 30000,              // 30 seconds
  STORAGE_KEY: "patient_create_draft",
  VERSION: "1.0",
  MAX_AGE_MS: 24 * 60 * 60 * 1000, // 24 hours
}
```

### Default Form Values

Pre-configured default values for initializing the form state.

## Usage Examples

### Import Types

```typescript
import type {
  PatientFormState,
  Demographics,
  MedicalHistory,
} from "@/components/Patient/types";
```

### Import Validation

```typescript
import {
  patientFormSchema,
  validatePatientForm,
  calculateAge,
} from "@/components/Patient/validation";
```

### Import Constants

```typescript
import {
  PATIENT_STATUS_OPTIONS,
  DEFAULT_FORM_VALUES,
  AUTO_SAVE_CONFIG,
} from "@/components/Patient/constants";
```

### Validate Form Data

```typescript
const formData: PatientFormState = {
  demographics: { /* ... */ },
  medicalHistory: { /* ... */ },
  clinicalNotes: { /* ... */ },
  administrative: { /* ... */ },
};

const { isValid, errors } = validatePatientForm(formData);

if (!isValid) {
  console.error("Validation errors:", errors);
}
```

### Calculate Age

```typescript
const age = calculateAge("1990-01-15");
console.log(`Patient age: ${age}`);
```

## Integration with Existing Code

This feature integrates with:

- **CARE UI Components**: Uses existing form components from `@/CAREUI`
- **Common Constants**: Extends `GENDER_TYPES` and `BLOOD_GROUPS` from `@/common/constants`
- **API Client**: Will use `@/lib/ApiClient` for API calls
- **TanStack Query**: For server state management
- **i18next**: For internationalization

## Next Steps

The following components will be implemented in subsequent tasks:

1. API integration layer (services and hooks)
2. Form input components
3. Form sections (Demographics, Medical History, etc.)
4. Main PatientForm component
5. Patient search and duplicate detection
6. PatientCreatePage route component
7. Internationalization
8. Accessibility features
9. Testing (unit, property-based, E2E)

## Requirements Mapping

This task implements:
- **Requirement 1.1**: Project structure for patient creation
- **Requirement 2.1**: Data models for patient demographics
- **Requirement 3.3**: Form validation schemas

## Related Files

- Design Document: `.kiro/specs/centralized-patient-creation/design.md`
- Requirements: `.kiro/specs/centralized-patient-creation/requirements.md`
- Tasks: `.kiro/specs/centralized-patient-creation/tasks.md`
