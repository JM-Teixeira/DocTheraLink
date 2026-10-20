# Patient API Service

API service functions for patient-related operations using TanStack Query.

## Files

- `routes.ts` - API route definitions
- `index.ts` - React Query hooks and utilities
- `index.test.ts` - Unit tests

## Usage

### Creating a Patient

```tsx
import { useCreatePatient } from '@/components/Patient/api';

const { mutate: createPatient, isPending } = useCreatePatient();

createPatient(patientData, {
  onSuccess: (data) => console.log('Created:', data),
  onError: (error) => console.error('Error:', error)
});
```

### Searching for Patients (Duplicate Check)

```tsx
import { useSearchPatients, getDuplicateCheckParams } from '@/components/Patient/api';

const searchParams = getDuplicateCheckParams(patientData);
const { data, isLoading } = useSearchPatients(searchParams);
```

### Getting Patient List

```tsx
import { usePatientList } from '@/components/Patient/api';

const { data, isLoading } = usePatientList({
  facility: 'facility-id',
  limit: 20
});
```

### Getting Patient Details

```tsx
import { usePatient } from '@/components/Patient/api';

const { data: patient, isLoading } = usePatient(patientId);
```

### Updating a Patient

```tsx
import { useUpdatePatient } from '@/components/Patient/api';

const { mutate: updatePatient } = useUpdatePatient();

updatePatient({ id: patientId, data: updatedData });
```

## Features

- **Error Handling**: Automatic retry logic for network errors
- **Cache Management**: Intelligent query invalidation
- **TypeScript**: Full type safety
- **Duplicate Detection**: Helper function for finding potential duplicates
