# Patients

Registering, editing, listing and viewing patients. The backend is
`be/internal/modules/patient` (rules, duplicates, identifiers) and
`be/internal/modules/timeline` (the patient history).

```
Patient/
├── PatientList.tsx           # list with filters and metrics
├── PatientMetrics.tsx        # counts by care level, gender, age
├── PatientForm.tsx           # register / edit (drafts, retry, lists)
├── ClinicalListEditor.tsx    # rows of allergies, conditions, medications
├── DuplicatePatientsDialog.tsx
├── PatientDetail.tsx         # patient page
├── PatientTimeline.tsx       # history: next sessions, referrals, messages, files
├── PatientIds.tsx            # NIF / SNS display
├── patientFormUtils.ts       # form <-> API, validation, drafts (pure, tested)
├── timelineUtils.ts          # filter / group / label history entries (pure, tested)
├── patientFilters.ts         # list filters (pure, tested)
├── models.tsx                # PatientModel (API shape)
└── api/                      # routes and TanStack Query hooks
```

## The form

- Every field is a string (lists are rows of strings) so inputs stay
  controlled; `formToPayload` turns empty values into `null` so an edit can
  clear them.
- Client validation mirrors the backend (NIF check digit, SNS, phone,
  e-mail, postal code `0000-000`, condition years); the server stays
  authoritative and its field errors show next to the input.
- **Drafts**: the form is saved in `localStorage` half a second after each
  change, per user and per patient (`draftKey`). They hold patient data, so
  they expire after 24 h and are removed on save, cancel, "Descartar
  rascunho" and sign-out (`clearAllDrafts` in `AuthUserProvider`).
- **Connection problems**: saves are retried twice (1 s, 2 s) only for
  network errors and 502/503/504 (`Utils/request/retry.ts`), so a request
  that may have reached the server is never sent twice by itself. After
  that the form shows a message with "Tentar de novo"; nothing is lost
  because of the draft.
- **Duplicates**: the server answers 409 with possible matches; same NIF or
  SNS blocks, same name + birth date or phone asks for confirmation.

## Tests

- Unit: `patientFormUtils.test.ts`, `timelineUtils.test.ts`,
  `patientFilters.test.ts`, `api/index.test.ts`, component tests
  `*.test.tsx`; coverage of this folder is checked by `npm run test:coverage`.
- Browser: `e2e/browser/tests/patients.spec.ts` and the registration journey.
