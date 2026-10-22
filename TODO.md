# TODO — Klyniva

Backlog. When an item is done, move it to `Completed.md` with a timestamp.
Architecture and conventions: `CONTEXT.md`.

Priority: **P0** blocks real-patient use · **P1** core product · **P2** important · **P3** later.
Status tags: **[proposed]** = waiting for product-owner review before work starts.

---

## Requests 2026-10-22 (batch 3, in progress)

Bugs / quick fixes
- [x] Adding a clinic user gets stuck on "loading plugin".
- [x] State admin cannot create inventory items.
- [x] Home: the Klyniva logo moves when switching EN/PT (fixed height).
- [x] Patients: the diagnosis filter text is cut ("CID").
- [x] Clinics: "Nova clínica" too close to the card below.
- [x] Confirm users with status "left" cannot sign in.
- [x] Sidebar collapse animation as smooth as the expand.

Features
- [ ] Settings page (user): date format, theme light/dark, language, software update, other useful preferences.
- [ ] Suggestions: anyone sends ideas from Settings; admins read them in one place.
- [ ] "?" help on each page: what it does and its features, in a pop-up (keeps pages short).
- [ ] Admins can view the app as any role ("ver como") to check what each role sees.
- [ ] Users: filters (dropdowns) for who can change which stock (category x floor), counting "all floors / all categories" grants.
- [ ] User page: professional situation at a glance (joined on, leave from/to, left on), from a dated history.
- [x] Clinic code (demo "de", Climutua "cm") and a username standard `<code>.<first>.<last>` (+ number on clash), applied to everyone.
- [ ] Waiting list: confirm several patients; highlight entries when a matching slot is free (any or a chosen physio) and alert the coordinator.
- [ ] User: field for what the person does with patients (to clarify with the product owner).

Repo / docs
- [ ] Clean up unused files, clear layout, rewrite README.md.

Hosting
- [ ] TLS certificates (local CA now, Let's Encrypt when public).
- [ ] Move hosting to the Ubuntu laptop as a server: same k8s manifests (k3s), notes to keep it modular and ready to scale (stateless backend, S3 uploads, managed Postgres, versioned migrations).

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
- [x] Browser E2E suite (Playwright, `e2e/browser`, `make e2e-browser`): login, people status filter, patient filters vs metrics, date preference, dropdown padding, no long dashes, clinic isolation, time board drill-down.
- [x] Browser E2E of the referral journey (doctor refers → physio accepts, starts, writes a progress note the doctor reads, completes → discharge letter; other clinic and front desk locked out), and the browser suite runs in CI against the compose stack (report uploaded on failure).
- [x] Browser E2E: patient registration with soft and blocking duplicates and NIF search; weekly plan from the board; holiday cover with a per-patient handover read by the cover physio; inventory use / open / finish.
- [x] Browser checks: notifications bell, letter to PDF, phone screens, axe accessibility scan (serious/critical) of 9 pages.
- [ ] P3 Accessibility next: moderate axe findings; keyboard-only pass of the time board grid; screen-reader labels of the board cells.
- [x] Access rules of the legacy user/facility/time-log handlers unit-tested (`api/handlers/access_test.go`) and covered by e2e (`TestNoPrivilegeEscalation`, `TestClinicIsolation`).
- [ ] P2 HTTP-level tests (httptest + PostgreSQL) for the legacy user and facility handlers.
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
- [x] **Access-control review of legacy handlers** (2026-10-22): user create/edit/delete, facility create/edit/delete, user and clinic reads, time logs, skills — see Completed.
- [ ] Same review for the remaining legacy reads (`/getallfacilities/`, geography) and a password-change policy (length, reuse).
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

- [x] Reach the local install from other laptops on the same network: `scripts/lan-access.ps1` (Windows, admin: port forwarding into WSL, firewall for private networks only, prints the addresses; `-Remove` undoes) + `scripts/k8s-local.sh lan` (backend address for links and CORS). README "Use it from another laptop".
- [ ] P3 LAN next: run lan-access automatically at start (scheduled task) so a reboot needs nothing; HTTPS with a local CA (Caddy) before any real data; QR code with the address.
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
- [x] Referral conversation: a reply answers a specific question; open questions show "n por responder" on the card and a "Por responder" filter; progress notes can carry pain 0-10, drawn as the evolution; thread filter Tudo / Progresso / Perguntas.
- [ ] Referral conversation next: goals per referral with progress against them; notify when a question waits more than 2 days.
- [ ] Later: clinic-defined templates stored in the database; optional full ICD-10 import (licensed data, PT edition).
- [x] Printable referral / discharge letter (A4 page; browser print → PDF).

### M4 — Patients (P1) ← next
- [x] New Klyniva patient pages: list/search, detail, create/edit (replace CARE's 3k-line pages).
- [x] Duplicate detection with confirmation: same NIF/SNS in the clinic is blocked (also by a partial unique index); similar name + same birth date or phone asks "different person?".
- [x] Patient form: draft auto-save (per user, 24 h, cleared on sign-out; restore or discard), e-mail, street / postal code / city, emergency contact name + relationship, insurer + member number; browser E2E of the flow.
- [x] Rest of the old "centralized patient creation" spec: retry on network errors (only when the server cannot have done the work), structured allergies / conditions / medications, ≥80 % coverage of the patient flow enforced in CI; unused old spec code removed.
- [x] Portuguese identifiers: NIF (mod-11 check digit) and SNS number (9 digits), stored normalised.
- [x] Patient filters (care level, age band, gender, clinic) with metrics of the match; search by name, NIF or SNS number.
- [x] Patient metrics: CSV export (coordinators/admins; contact and demographic columns only, safe against spreadsheet formulas), new patients per month (12 months), filter by physio and by diagnosis / ICD (clinical roles and coordinators only).
- [x] Patient timeline: next sessions on top, then referral steps, messages (with pain), files and sessions by month; filters by kind; each role sees only its own sources (no referrals for the front desk).
- [x] Consent records (GDPR) per patient: health data, sharing with the doctor, e-mail, SMS, photos; given/withdrawn, how, by whom, append-only history; warning while health-data consent is missing.
- [ ] Consents next: printable consent form to sign; block clinical notes / SMS when the matching consent is missing or withdrawn.

### M3 — Scheduling (P1) — started early on request (time board)
- [x] **Time board** (`/time-board`): divisions of the clinic (gym, rooms, pool…) with a capacity of patients per physio per hour; physio shifts per day per division; patients assigned per physio-hour (capacity, on-shift and double-booking checks, safe under concurrency); views by physio and by division; copy a day's shifts; attended / missed.
- [x] **Coordinator** role (D6, confirmed): plans divisions and shifts, sees each physio's load and **assigns patients to physios** (with admins; receptionists no longer assign); reads the clinic's referrals (read-only) for the patient pop-up.
- [x] **Care level 1–5** per patient = credits; each division has an hourly credit limit per physio (default 8) besides the patient limit — two level-5 patients never share an hour.
- [x] **Weekly plans**: book a patient 1/2/3× per week or every weekday for N sessions, with a preview of skipped dates; cancel the rest of a plan. **Repeat week** copies a week of shifts forward.
- [x] **Day / week / month / year** views; the year view counts sessions held/missed, patients seen, referred and discharged per month.
- [x] Clicking a patient on the board opens a pop-up: age, contact, care level, allergies, the doctor's referral and the doctor–physio thread, links to the patient and referral pages.
- [x] **Home page** (logo / Início) per role: physio "my day" (mark sessions in place), coordinator clinic load + what needs attention, doctor's referrals, front-desk arrivals, notifications.
- [x] Second demo clinic **Climutua** (PT-PT, 120 patients, realistic load) — see README.
- [x] **Floors** and filters (floor / division / physio); physio drill-down with patient names per hour; decluttered cells with a legend.
- [x] **Absences with cover**: patients move to the cover physio at the same hour (preview first), handover document + note per patient, cover reads the absent physio's referrals; absent physios can't be booked.
- [x] Holiday per patient: each session to a chosen physio (with the room each one has left at that hour), "who receives whom" summary.
- [x] Misses: justified or not (+ reason); 3 unjustified in 90 days flag the patient to the coordinator.
- [ ] Absences next: partial days; notify the physios who receive patients; only the default cover reads the referrals today - extend to everyone who received a patient.
- [ ] Misses next: threshold and window per clinic; tell the doctor of the referral; list of misses on the patient page.
- [x] Move a session: drag it to another hour in the physio day, or "Mudar sessão" (other day/hour/physio); same checks as booking.
- [x] Reschedule the rest of a plan to new days/hour/physio (preview first, same number of sessions, plan keeps its id).
- [x] Waiting list: urgent first then oldest, preferred days/period/physio; "Marcar plano" books it and the patient leaves the list; count on the board and the coordinator home.
- [ ] Next on the board: a physio's own "my day" view; print the day sheet.
- [ ] Link board appointments to referrals (field exists) and book the whole series of N sessions at once; receptionists manage them.
- [ ] Therapist availability from the timetable; room/section capacity; conflict checks.
- [x] Old unrouted WIP screens (`Facility/TimetableGrid`, `Therapy/*`) and the legacy `/timetable` + `/sections` API removed — superseded by the time board.
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
- [x] **Inventory of consumables** (D1), done early on request (`/inventory`, sidebar): items per clinic with alert levels, use / restock / count with history, low and out-of-stock filters, demo stock seeded. Replaced the CARE inventory screens and 501 stubs.
- [x] Inventory **audit** (coordinators/admins): every use, opening, restock and count with who and when, totals per person, filters; new **Open** action (opened a box/bottle).
- [x] **Automatic use per session**: items with a per-session quantity (e.g. 1 pair of gloves) are taken from stock when a session is marked attended, and given back if that is undone; shortages are reported, never block the session.
- [x] **Use vs open defined**: single-use items are *used*; containers (bottles, packs, rolls) are *opened* (who/when, sealed → in use) and *finished* when empty, with an optional "discard N days after opening" alert.
- [x] Per-session use **per division** (e.g. swim caps only in the pool, 2 pairs of gloves in hydrotherapy).
- [x] **Reorder list** with suggested quantities (up to a target level) and CSV export.
- [x] Inventory permissions by floor × category (coordinators/admins give them from the people list); items have a floor; admins filter by clinic; single-use vs container and floor filters; audit as a plain log; per-session use explained.
- [ ] Inventory next: suppliers and order history; batch / expiry dates of sealed stock; permissions for specific items (today: floors × categories).
- [ ] Wire the existing clock in/out button and time-log history (`Users/ClockInOutButton`, `Users/TimeLogHistory` — built but never shown); time-tracking reports and exports.
- [x] People list with simple filters; employment status (active / on leave / left — kept on record, no sign-in); skills and clinic tabs (`internal/modules/staff`).
- [x] Clinic → people opens `/users?clinic=` (old address redirects); clinics page rebuilt with search + chips.
- [ ] People next: edit/deactivate from the list; remove the old `ManageUsers`/`UserFilter`/`LinkedSkills*`/`LinkedFacilities*` screens; tests for the legacy `ListUsers` handler.
- [ ] Organisations with several clinics (switch clinic, per-clinic admins) on top of `internal/clinic`.
- [ ] Dashboards: referrals by status, waiting time to first session, therapist load, outcomes.

### M6 — Billing (parked, D4)
- [ ] Price lists per act, invoices, payments; insurers (ADSE, SNS agreements, private insurers, self-pay).
- [ ] Portuguese certified invoicing / SAF-T compliance (likely via an integration — legal requirement).

### M7 — Patient-facing (P3)
- [ ] Patient portal / app: appointments, exercise programme, messages, documents.
- [ ] Video teleconsultation sessions.
- [ ] Interoperability: HL7 FHIR export/import.
