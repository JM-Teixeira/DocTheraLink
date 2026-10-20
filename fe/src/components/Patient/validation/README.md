# Patient Data Model Validation Tests

## Overview
This directory contains property-based tests for validating the patient data model used in the centralized patient creation feature.

## Test Framework
- **Vitest**: Unit testing framework
- **fast-check**: Property-based testing library for TypeScript

## Running Tests

### Run all validation tests
```bash
npx vitest --run src/components/Patient/validation/patientValidation.test.ts
```

### Run tests in watch mode
```bash
npx vitest src/components/Patient/validation/patientValidation.test.ts
```

### Run tests with UI
```bash
npx vitest --ui
```

## Test Coverage

### Property 1: Type Safety
**Validates: Requirements 2.1, 3.3**

Tests that all patient data fields maintain their correct TypeScript types:
- `name`: string (non-empty, trimmed)
- `phone_number`: string (10-15 digits)
- `gender`: number (1-3)
- `year_of_birth`: number (1900-2024)
- `address`: string (non-empty, max 500 chars)
- `pincode`: number (6 digits: 100000-999999)
- `id`: string (UUID format)
- `facility`: string (non-empty)
- Optional fields: `emergency_phone_number`, `blood_group`, `is_vaccinated`

### Property 2: Required Fields Validation
**Validates: Requirements 2.1, 3.3**

Tests that required fields cannot be empty or invalid:
- All required fields must be present
- String fields cannot be empty or whitespace-only
- Numeric fields must be within valid ranges
- All required fields must be validated together

### Property 3: Data Integrity
Tests that patient data maintains integrity across transformations:
- Serialization/deserialization preserves all field values
- No data loss during JSON operations

## Test Statistics
- **Total Tests**: 17 property-based tests
- **Test Runs per Property**: 100 random inputs per test
- **Total Validations**: ~1,700 test cases executed

## Property-Based Testing Benefits
1. **Comprehensive Coverage**: Tests with 100 random inputs per property instead of hand-picked examples
2. **Edge Case Discovery**: Automatically finds edge cases (e.g., whitespace-only strings)
3. **Regression Prevention**: Ensures data model changes don't break type safety
4. **Documentation**: Tests serve as executable specifications

## Generators (Arbitraries)
The tests use smart generators that produce valid test data:
- `patientNameArbitrary`: Non-empty strings with trimmed whitespace
- `phoneNumberArbitrary`: Valid 10-15 digit phone numbers
- `genderArbitrary`: Valid gender codes (1-3)
- `yearOfBirthArbitrary`: Realistic birth years (1900-2024)
- `pincodeArbitrary`: Valid 6-digit Indian pincodes
- `patientIdArbitrary`: UUID format strings

## Maintenance
When adding new fields to `PatientModel`:
1. Add appropriate generator in the arbitraries section
2. Add type safety test for the new field
3. If required, add validation test for non-empty constraint
4. Update this README with the new field documentation
