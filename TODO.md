# TODO — Klyniva

Backlog. When an item is done, move it to `Completed.md` with a timestamp.
Architecture and conventions: `CONTEXT.md`.

Priority: **P0** blocks real-patient use · **P1** core product · **P2** important · **P3** later.
Status tags: **[proposed]** = waiting for product-owner review before work starts.

---

## 0. Testing policy (mandatory for every change)

Tests are what let us change the product safely. A feature or fix is **not
done** until all of these exist and CI is green:

1. **Service unit tests** for every business rule (permissions, validation,
   state transitions) using an in-memory fake repository.
2. **Integration tests** (`-tags integration`, real PostgreSQL via
   testcontainers) for every repository query and migration change.
3. **HTTP handler tests** for status codes / error mapping of every endpoint.
4. **Frontend tests** (Vitest + Testing Library) for pure logic (`*Utils.ts`)
   and for key components (rendering by role, forms, error states).
5. **End-to-end API scenario** (`be/internal/e2e`, `-tags e2e`) for each user
   journey, run against a deployed stack (compose in CI, k3d locally).
6. **Every bug gets a regression test first**, then the fix.
7. Coverage targets for new modules: backend services ≥ 80 %, frontend
   module logic ≥ 80 % (the old patient-creation spec also set 80 %).

Open testing work:
- [ ] P1 Browser E2E for the referral journey (Playwright or Cypress) incl. screenshots — UI is only checked manually today.
- [ ] P1 Tests for legacy handlers still in use: time log, timetable, users, facilities.
- [ ] P1 Component tests for `ReferralDetail`, `ReferralList`, `CreateReferral` with a mocked API (role-dependent buttons, validation).
- [ ] P2 Coverage reporting in CI (Go `-coverprofile`, Vitest `--coverage`) with thresholds.
- [ ] P2 Contract tests: generate OpenAPI from the backend and type-check the frontend API routes against it.
- [ ] P2 Load/smoke test (k6) of login + referral list before production.

## 1. Security, privacy & compliance (P0 — before any real patient data)

- [ ] **GDPR / health-data baseline**: processing record, legal basis, retention, patient data export & erasure, DPA with hosting provider (CNPD guidance, Portugal).
- [ ] **Audit log**: append-only record of who read/changed which patient and referral, with admin viewer.
- [ ] **Password reset**: endpoints are stubs that always "succeed" — signed expiring tokens + e-mail.
- [ ] **Login rate limiting / lockout** (`RATE_LIMIT_*` config exists, unused).
- [ ] **Token storage**: move refresh token to an httpOnly cookie; revocation on logout/password change.
- [ ] **Headers**: `Cache-Control: no-store` on API responses; tighten nginx CSP.
- [ ] **Access-control review of legacy handlers** (users, facilities, timetable have little per-facility scoping, e.g. `GET /users/` lists everyone).
- [ ] Encrypted backups; TLS to managed Postgres (`DB_SSLMODE=verify-full`).
- [ ] Dependency & image scanning in CI (govulncheck, npm audit, Trivy). The IDE scan
      flags critical CVEs in the local `nginx:1.29-alpine` base image — pin a patched
      digest and rebuild regularly.

## 2. Engineering (P1/P2)

- [ ] P1 **Versioned SQL migrations** (goose/Atlas) run as a k8s Job instead of AutoMigrate at startup — prerequisite for >1 backend replica. Also drop the tables of removed CARE modules (beds, locations, consultations, daily_rounds, assets…) that existing databases still contain.
- [ ] P2 Structured logging (`log/slog`, JSON), request IDs, no patient data in logs.
- [ ] P2 Frontend formatting pass (Prettier/ESLint `--fix`, LF endings) on `fe/src`, then lint all of `src/` in CI (today `npm run lint:strict` covers new code only).
- [ ] P2 Frontend upgrades: React 19, Tailwind 4, ESLint 9 flat config; replace deprecated `useTanStackQueryInstead`.
- [ ] P2 Remove remaining ICU-era constants/types (`common/constants.tsx`, `Patient/models.tsx` daily-round types, COVID fields on the patient model).
- [ ] P2 Swagger/OpenAPI regeneration (`make -C be swagger`) and frontend types from it.
- [ ] P2 Legacy models use `UserBareMinimum` relations without a table name (GORM may create `user_bare_minimums`); audit.
- [ ] P3 Fold `be/scripts/portugal_data.sql` (full district list) into seeding or delete it.

## 3. Deployment & operations (P2/P3)

- [ ] P2 Container registry + CD: GitHub Actions → GHCR tagged by git SHA, pin tags in `k8s/overlays/production`.
- [ ] P2 Uploads to S3-compatible storage (needed for >1 backend replica) — add an S3 implementation of `internal/storage.Store`.
- [ ] P2 Virus scanning of uploaded attachments (ClamAV sidecar) before they are downloadable.
- [ ] P3 Managed PostgreSQL or CloudNativePG with PITR backups.
- [ ] P3 Observability: Prometheus metrics, OpenTelemetry traces, Sentry DSN wiring.
- [ ] P3 Redis only if needed for rate limiting/sessions (removed from k8s for now).
- [ ] P3 GitHub repo rename `DocTheraLink` → `klyniva` (product-owner action in GitHub settings).

---

## 4. Product roadmap (approved 2026-10-21)

Order (D5): **M1 → M4 → M3 → M2 → M5**. Each milestone ships with the tests from §0.

### Decisions (product owner, 2026-10-21)
- **D1** Remove the inherited CARE ICU modules (assets, beds/locations, consultations/daily rounds, prescriptions, shifting, resource requests, TeleICU/camera, HCX insurance, death report, external results) — **done**. **Keep inventory**: therapists must track consumables (masks, gloves, gels, tape…) → rebuilt in M5.
- **D2** One clinic for now, but **modular for many**: a clinic = a facility; users work in their `home_facility`; new modules scope data through `internal/clinic` — **foundation done**.
- **D3** Add a **Receptionist** role — **done** (patients yes, referrals no; appointments in M3).
- **D4** Billing **not** in the first stages (M6 stays parked).

### M1 — Close the referral loop (P1) — done (PR #3)
- [x] "Refer to physiotherapy" button on the patient page.
- [x] Notifications: in-app bell + e-mail on new referral (therapists), accepted/declined/completed (doctor), new message. _(e-mails carry no patient data; SMTP optional)_
- [x] Attachments on referrals (imaging, reports, PDFs; content-sniffed, 10 MB, participants upload, uploader/admin delete) — replaced the `/files` stubs.
- [x] Re-assign / hand-over to another therapist; re-open a declined referral to another therapist.
- [x] Filters & search on the referral list (patient, priority, date, therapist) and counts per status tab.
- [x] ICD-10 code picker (curated physio subset, EN/PT, code or word search); 9 built-in referral templates.
- [ ] Later: clinic-defined templates stored in the database; optional full ICD-10 import (licensed data, PT edition).
- [x] Printable referral / discharge letter (A4 page; browser print → PDF).

### M4 — Patients (P1) ← next
- [x] New Klyniva patient pages: list/search, detail, create/edit (replace CARE's 3k-line pages).
- [ ] Rest of the old "centralized patient creation" spec (from `.kiro`): **duplicate detection with confirmation** (reuse `Facility/DuplicatePatientDialog`), draft auto-save, retry on network errors, accessibility checks, browser E2E of the flow, ≥80 % coverage. Backend model grows to match: emergency contact (name/relationship), address parts, insurance, structured allergies/medications/conditions.
- [ ] Portuguese identifiers: NIF and SNS number with validation.
- [ ] Patient timeline: referrals, sessions, notes and documents in one view.
- [ ] Consent records (GDPR) per patient.

### M3 — Scheduling (P1)
- [ ] Appointments linked to referrals; book the whole series of N sessions at once; receptionists manage them.
- [ ] Therapist availability from the timetable; room/section capacity; conflict checks.
- [ ] Wire the existing (unrouted) WIP screens: `Facility/TimetableGrid`, `Therapy/*` (timetables, section manager) — rebuild on the module pattern where needed.
- [ ] Day/week calendar per therapist and per section; drag to reschedule.
- [ ] Waiting list, cancellations, no-show tracking.
- [ ] SMS/e-mail reminders to patients.

### M2 — Physiotherapy clinical record (P1)
- [ ] Treatment plan per referral: goals, planned sessions, frequency, modalities (electrotherapy, hydrotherapy, massage, gym — the clinic "sections").
- [ ] Session notes in SOAP format; pain scales (VAS/NRS); ROM and strength measurements.
- [ ] Standardised outcome questionnaires (e.g. Oswestry, DASH, KOOS, NDI) with progress charts.
- [ ] Home exercise programmes from an exercise library (images/video), printable or sent to the patient.
- [ ] Discharge summary generated from the record and sent to the referring doctor automatically.

### M5 — Clinic administration (P2)
- [ ] **Inventory of consumables** (D1): catalogue (seeded: masks, gloves, gels, tape, electrodes…), stock in/out per clinic, minimum levels with low-stock alerts, usage per session (link to M2). Replaces the CARE inventory screens and `/facility/:id/inventory` stubs.
- [ ] Wire the existing clock in/out button and time-log history (`Users/ClockInOutButton`, `Users/TimeLogHistory` — built but never shown); time-tracking reports and exports.
- [ ] Organisations with several clinics (switch clinic, per-clinic admins) on top of `internal/clinic`.
- [ ] Dashboards: referrals by status, waiting time to first session, therapist load, outcomes.

### M6 — Billing (parked, D4)
- [ ] Price lists per act, invoices, payments; insurers (ADSE, SNS agreements, private insurers, self-pay).
- [ ] Portuguese certified invoicing / SAF-T compliance (likely via an integration — legal requirement).

### M7 — Patient-facing (P3)
- [ ] Patient portal / app: appointments, exercise programme, messages, documents.
- [ ] Video teleconsultation sessions.
- [ ] Interoperability: HL7 FHIR export/import.
