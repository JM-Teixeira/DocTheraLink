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

- [ ] P1 **Versioned SQL migrations** (goose/Atlas) run as a k8s Job instead of AutoMigrate at startup — prerequisite for >1 backend replica.
- [ ] P1 **Remove inherited CARE modules we don't need** (frontend screens + backend 501 stubs): see decision D1 below.
- [ ] P2 Structured logging (`log/slog`, JSON), request IDs, no patient data in logs.
- [ ] P2 Frontend formatting pass (Prettier/ESLint `--fix`, LF endings) on `fe/src` (~4.5k mostly-formatting errors), then lint all of `src/` in CI.
- [ ] P2 Frontend upgrades: React 19, Tailwind 4, ESLint 9 flat config; replace deprecated `useTanStackQueryInstead`.
- [ ] P2 Swagger/OpenAPI regeneration (`make -C be swagger`) and frontend types from it.
- [ ] P2 Legacy models use `UserBareMinimum` relations without a table name (GORM may create `user_bare_minimums`); audit.
- [ ] P2 Seeded doctor's first name is `"Dr. John"` — fix seed data.
- [ ] P3 Fold `be/scripts/portugal_data.sql` (full district list) into seeding or delete it.

## 3. Deployment & operations (P2/P3)

- [x] Local Kubernetes (k3d) — see Completed.md.
- [ ] P2 Container registry + CD: GitHub Actions → GHCR tagged by git SHA, pin tags in `k8s/overlays/production`.
- [ ] P2 Uploads to S3-compatible storage (needed for >1 backend replica).
- [ ] P3 Managed PostgreSQL or CloudNativePG with PITR backups.
- [ ] P3 Observability: Prometheus metrics, OpenTelemetry traces, Sentry DSN wiring.
- [ ] P3 Redis only if needed for rate limiting/sessions (removed from k8s for now).
- [ ] P3 GitHub repo rename `DocTheraLink` → `klyniva` (product-owner action in GitHub settings).

---

## 4. Product roadmap **[proposed — please review]**

Each milestone is shippable on its own and comes with the tests from §0.

### M1 — Close the referral loop (P1)
- [ ] "Refer to physiotherapy" button on the patient page (`/referrals/new?patient=<id>` already works).
- [ ] Notifications: in-app bell + e-mail on new referral (therapists), accepted/declined/completed (doctor), new message.
- [ ] Attachments on referrals and messages (imaging, reports, PDFs) with access control.
- [ ] Re-assign / hand-over to another therapist; re-open a declined referral to another therapist.
- [ ] Filters & search on the referral list (patient, priority, date, therapist) and counts per status tab.
- [ ] ICD-10 code picker; referral templates for common conditions (ACL rehab, low back pain, frozen shoulder…).
- [ ] Printable referral / discharge letter (PDF).

### M2 — Physiotherapy clinical record (P1)
- [ ] Treatment plan per referral: goals, planned sessions, frequency, modalities (electrotherapy, hydrotherapy, massage, gym — the facility "sections").
- [ ] Session notes in SOAP format; pain scales (VAS/NRS); ROM and strength measurements.
- [ ] Standardised outcome questionnaires (e.g. Oswestry, DASH, KOOS, NDI) with progress charts.
- [ ] Home exercise programmes from an exercise library (images/video), printable or sent to the patient.
- [ ] Discharge summary generated from the record and sent to the referring doctor automatically.

### M3 — Scheduling (P1)
- [ ] Appointments linked to referrals; book the whole series of N sessions at once.
- [ ] Therapist availability from the existing timetable; room/section capacity; conflict checks.
- [ ] Day/week calendar per therapist and per section; drag to reschedule.
- [ ] Waiting list, cancellations, no-show tracking.
- [ ] SMS/e-mail reminders to patients.

### M4 — Patients (P1)
- [ ] Finish the "centralized patient creation" spec (from the old `.kiro` folder): reusable patient form (create/edit), debounced search, **duplicate detection with confirmation**, draft auto-save, field-level errors with retry, accessibility (WCAG), E2E test of the flow. Backend model must grow to match the form (emergency contact, address parts, insurance, structured allergies/medications/conditions).
- [ ] Portuguese identifiers: NIF and SNS number with validation.
- [ ] Patient timeline: referrals, sessions, notes and documents in one view.
- [ ] Consent records (GDPR) per patient.

### M5 — Clinic administration (P2)
- [ ] Organisations with several clinics (multi-tenant data separation).
- [ ] Staff management with roles incl. a new **Receptionist** role (scheduling, check-in, no clinical notes).
- [ ] Time-tracking reports and exports (clock in/out already works).
- [ ] Dashboards: referrals by status, waiting time to first session, therapist load, outcomes.

### M6 — Billing (P3)
- [ ] Price lists per act, invoices, payments; insurers (ADSE, SNS agreements, private insurers, self-pay).
- [ ] Portuguese certified invoicing / SAF-T compliance (likely via an integration — legal requirement).

### M7 — Patient-facing (P3)
- [ ] Patient portal / app: appointments, exercise programme, messages, documents.
- [ ] Video teleconsultation sessions.
- [ ] Interoperability: HL7 FHIR export/import.

### Decisions needed from the product owner
- **D1** Which inherited CARE modules stay? Candidates to **remove**: assets, inventory, beds, shifting, resource requests, consultations/daily rounds, TeleICU/camera feed, HCX, medicine/prescriptions, death report, external results. Candidates to **keep**: facilities, users, patients, time log, timetable, notifications.
- **D2** Single clinic or multi-clinic organisations from the start (affects M5 and every query)?
- **D3** Add a Receptionist role now?
- **D4** Billing in scope this year, or rely on existing invoicing software?
- **D5** Order of milestones (suggested: M1 → M4 → M3 → M2 → M5).
