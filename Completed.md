# Completed — Klyniva

Log of finished work, newest first. Timestamps are local time (Europe/Lisbon,
UTC+01:00). Commits are on branch `chore/project-foundation` unless noted.

---

## 2026-10-22 · Security review, date preference, tests (branch `feat/m4-patients`)

### 17:00 · Browser tests for the main clinic journeys (TODO §0)
- `e2e/browser/tests/journeys.spec.ts`: register a patient, then the
  same NIF is blocked and the same name + birth date warns ("É outra
  pessoa" registers anyway), and the NIF finds exactly one; a weekly plan
  (3× a week) booked from the board lands Mon/Wed/Fri at the chosen hour;
  a holiday recorded in the Ausências dialog moves the session to the
  cover physio, who reads the per-patient handover in the patient pop-up;
  single-use items are used, containers opened and finished.
- Each test makes and removes its own data; the suite runs one test at a
  time (shared demo users). 15 browser tests, 45/45 over three runs.

### 15:30 · Referral journey in the browser, browser tests in CI (TODO §0)
- `e2e/browser/tests/referral.spec.ts`: two browsers at once, the doctor
  refers a new patient to a named physio; the physio accepts, starts,
  writes a progress note that the doctor then reads, completes with an
  outcome; the discharge letter shows patient and outcome; another
  clinic's physio gets 404 and the front desk 403.
- CI (`End-to-end (docker compose)` job) now runs the Playwright suite
  after the API tests; the HTML report and traces are uploaded on failure.
- 11 browser tests, 33/33 passing over three repeated runs locally.

### 14:00 · Security fixes, date format preference, browser e2e (user request)
- **Security fixes** in legacy handlers (all were open to any signed-in
  user): creating users with any role (incl. admin), editing any account
  incl. `user_type` / `is_superuser` (privilege escalation), deleting
  users, creating/editing/deleting clinics, reading other clinics' people,
  staff lists, skills and time logs. Now: admins manage everything;
  coordinators add people to their own clinic with clinic roles only;
  everyone edits their own profile fields; status, clinic, password and
  flags only through their own endpoints; other clinics answer 404. Lists
  capped at 100; no raw database errors or debug prints in responses.
- "Add person" works again (`POST /users/add_user/` was never routed).
- **Date format preference** (Profile → Preferências): dd/mm/aaaa,
  mm/dd/aaaa, aaaa-mm-dd or "21 out 2026", shown and typed that way
  everywhere (new `DateInput` with a calendar button replaces the browser
  date fields, which always follow the browser's locale).
- Dropdown arrows no longer overlap the text (global select padding).
- No long dashes in the interface ("-" or "·"); Climutua is now
  "Climutua · Medicina Física e Reabilitação".
- Fix: a physio's hours on the board could come out of order (CI flake).
- Tests: access policy unit tests; staff visibility and preferences; e2e
  `TestNoPrivilegeEscalation`, `TestClinicIsolation`, `TestDatePreference`;
  component tests for DateInput and FilterChips; new **browser e2e suite**
  (Playwright, `make e2e-browser`); 120 frontend tests.

## 2026-10-21 — M4 patients (branch `feat/m4-patients`)

### 23:30 · Calmer time board, holiday cover, people & patient filters (user request)
- **Time board**: floors (divisions get a floor; Climutua has 3), filters
  by floor / division / physio (remembered per browser), decluttered cells
  (count + credit bar, details on hover), legend, division view grouped by
  floor. Click a physio → **their day** with every patient's name per hour.
- **Holiday cover**: coordinators record an absence (vacation, sick,
  training, other) with a cover physio; preview shows which sessions move
  at the same hour and which can't; a **handover document** plus a note per
  patient. The cover physio sees it on the home page and in the patient
  pop-up, and can read/post on the absent physio's referrals. Absent
  physios can't be booked; weekly plans skip their days.
- **Assign patient**: search by name, NIF or SNS number (IDs shown).
- **People** (`/users`): new list with one search box (full name too),
  role and status chips, clinic for admins; non-admins see only their clinic.
- **Employment status** (active / on leave / left, date, note): "left"
  keeps the record but blocks sign-in. **Skills** (14 PT-PT physio skills)
  and **clinic** profile tabs rebuilt (they were calling removed endpoints).
- **Patients**: filters by care level, age band, gender, clinic (admins),
  with metrics of the match (total, average age, care level, age and
  gender spread — `GET /patient/stats/`).
- **Referrals**: filters by referring doctor, physio and date range.
- Translations: dates follow the language; breadcrumbs translated and only
  collapse ("•••") on long paths; "doente" → "utente"; sidebar user card
  without the arrow.
- Tests: staff, absence/cover and referral-cover unit tests; Climutua
  integration (floors, holiday, skills, former staff); e2e
  `TestPeopleAndEmployment`, `TestPatientFiltersAndStats`,
  `TestReferralDoctorFilter`, `TestHolidayCover`; frontend filter/query
  helpers and array query params (105 tests).

### 18:00 · Home page, Climutua demo clinic, inventory kinds, People fix (user request)
- **Fix**: Clinic → People failed ("Something went wrong"): the
  `/facility/:id/get_users/` endpoint was a 501 stub; now lists the clinic's
  users with search and paging.
- **Home page** at `/` (logo and "Início"), by role: physio's day with
  patients (pop-up, attended/missed), coordinator's clinic load and needs
  attention (reorder, containers open too long, referrals waiting), doctor's
  referrals, front-desk arrivals; notifications; quick actions.
- **Inventory**: item kinds — single-use (*use*) vs containers (*open*, then
  *finish*; who/when; discard-after-days alert); per-session use per
  division (rules); reorder list with target levels and CSV export; audit
  shows finished too.
- **Climutua** (PT-PT): 13 staff, 6 divisions, 120 patients with valid NIF,
  20 conditions with ICD-10, referrals in every state with messages, weekly
  shifts ±4 weeks, ~37 sessions a day within patient/credit limits, stock
  with deliveries, openings and automatic session use.
- Tests: inventory kinds/rules/reorder (unit), Climutua consistency
  (PostgreSQL: limits, shifts, NIF, idempotent), e2e `TestClinicPeople`,
  `TestContainersAndReorder`, `TestClimutua`; 89 frontend tests.

### 15:30 · Care levels, weekly plans, calendar views, stock audit (user request)
- Patients get a **care level 1–5**; it is stored on each appointment as
  credits. Divisions have a per-physio hourly **credit limit** (default 8)
  next to the patient limit, checked under the booking lock.
- Only **coordinators and admins assign patients**; coordinators can read
  the clinic's referrals (not act on them).
- **Weekly plans** (`POST /schedule/series/`, dry-run preview): weekdays +
  hour + start + N sessions; skipped dates are reported with the reason;
  cancel the rest of a plan. **Repeat week** for N weeks.
- **Week / month / year views** (`/schedule/range/`, `/schedule/year/`).
- **Patient pop-up** on the board: basics, care level, allergies, current
  referral (diagnosis, reason, goals, precautions) and the thread.
- **Inventory audit** tab, **Open** movements, and **per-session automatic
  use** wired from the time board (attended → use; undo → give back).
- Tests: unit (credits, plans, ranges, hook, audit, coordinator read),
  PostgreSQL (credits, auto-use/undo, audit, plan, year stats), e2e
  `TestCarePlanAndAutoUse`; 83 frontend tests.

### 13:30 · Time board, working inventory, Coordinator role (user request)
- **Time board** (`modules/schedule`, `/time-board`): divisions with a
  per-hour capacity per physio; daily shifts per physio and division (no
  overlaps, edits cannot strand booked patients); patients assigned to a
  physio-hour (must be on shift, capacity, patient not double-booked;
  advisory locks make it safe under concurrency); board by physio (load per
  hour, day total) and by division (who is where); copy a day's shifts;
  attended/missed. Demo divisions and two weeks of weekday shifts seeded.
- **Inventory** (`modules/inventory`, `/inventory` in the sidebar): items per
  clinic with alert level, use/restock/count with history (row-locked, never
  below zero), filters for low/out, demo stock seeded.
- **Coordinator** role + demo user `coordinator` (Marta Ribeiro).
- "How referrals work" explainer on the referrals page.
- Removed: CARE inventory pages and 501 stubs, legacy timetable/sections
  API and models, unrouted Therapy/TimetableGrid WIP, dev sample data.
- Tests: service rules (unit), concurrency + SQL on PostgreSQL, e2e
  `TestInventory` and `TestTimeBoard`, frontend helpers (80 tests).

### 11:40 · NIF / SNS number and duplicate-patient detection
- Backend: `nif` and `sns_number` on patients (NIF check digit validated,
  both normalised to 9 digits); partial unique indexes per clinic.
  Registering returns 409 `duplicate_patient` with the matching visible
  patients: same NIF/SNS always blocks; similar name (accent/case
  insensitive, same first and last name) + same birth date or phone
  (last 9 digits) needs `allow_duplicate`. Editing to another patient's
  NIF/SNS is blocked. Shared `internal/textutil` (fold, words, digits).
- Frontend: NIF / SNS fields with inline validation, shown on the patient
  page; "may already be registered" dialog (open existing / register as a
  different person).
- Tests: identifiers, name similarity, duplicate rules (unit), SQL
  pre-filter + unique index + clinic isolation (PostgreSQL), e2e
  `TestPatientIdentifiersAndDuplicates`, form helpers.

### 11:00 · Local access
- `make k8s-start` / `scripts/start-klyniva.cmd` start the cluster after a
  reboot and keep WSL alive; README lists every demo login.

---

## 2026-10-21 — M1 referral loop (branch `feat/m1-referral-loop`)

### 10:45 · Printable referral letter / discharge summary — M1 complete
- `/referrals/:id/letter`: A4 letter with clinic letterhead, reference,
  patient, clinical details, signatures and a confidentiality footer; a
  completed referral prints as a discharge summary (dates, therapist,
  outcome). "Print / save as PDF" uses the browser; the print copy is
  portalled under `<body>` so the app layout does not affect it.
- Print button on the referral page. Tests: letter helpers; print output
  checked by rendering to PDF with Playwright.

### 10:15 · ICD-10 picker and referral templates
- Backend `modules/catalog`: embedded, validated-at-start-up data — 65
  physiotherapy-relevant ICD-10 codes and 9 templates (low back pain, neck
  pain, frozen shoulder, rotator cuff, ACL reconstruction, knee OA, hip
  replacement, ankle sprain, post-stroke), EN/PT. `GET /catalog/icd10/`
  searches by code or accent-insensitive words; `GET /catalog/referral-templates/`.
- Frontend: ICD combobox (keyboard + mouse, fills an empty diagnosis) and
  "Start from a template" chips on the new-referral form (asks before
  overwriting typed text).
- Tests: data integrity, search ranking, language fallback, HTTP; form
  helpers; e2e `TestCatalog`.

### 09:30 · Attachments on referrals
- Backend: `internal/storage` (Store interface, atomic local files) and
  `modules/attachment`: upload (multipart, content sniffed — PDF, JPEG,
  PNG, WebP, text; 10 MB), list, download (`attachment` disposition,
  nosniff, no-store), delete. Participants upload; anyone who can see the
  referral downloads; uploader or admin deletes. File names sanitised.
- Frontend: "Files" panel on the referral page with drag & drop, client-side
  checks, download with the auth header, delete with confirmation (EN/PT).
- Removed the CARE `/files` 501 stubs and the orphaned file-upload
  components/hooks; proxy body limit raised to 12 MB.
- Tests: storage, attachment service/handler, frontend helpers, e2e
  `TestAttachments` (upload, byte-exact download, spoofed .exe → 415,
  outsider → 404, other participant cannot delete).

### 08:30 · Notifications (b8b15ab)
- In-app bell + generic e-mails for referral events; see commit message.

### 07:30 · Filters, counts, hand-over and re-open (046f8d3)
- Referral list search/priority/therapist/date filters and per-status
  counts; hand-over to another therapist; re-open a declined referral.

---

## 2026-10-21 — Roadmap decisions D1–D5 approved; D1–D3 implemented

### 00:10 · D1 remove CARE ICU modules, new patient pages
- Frontend: removed assets, beds/locations, camera/vitals, consultations/daily
  rounds, prescriptions, shifting, resources, HCX insurance, death report,
  external results, patient notes/consents, hub-spoke network (~100 route
  entries, ~40k lines). Kept facilities (now "Clinic"), users, inventory,
  notifications, time log/timetable, therapy WIP screens.
- New Klyniva patient pages: list/search, detail (with referral history and
  "Refer to physiotherapy" — first M1 item), create/edit form with validation
  and server field errors (EN/PT). Shared `Utils/roles.ts`, `Utils/person.ts`.
- Backend: removed asset/consultation handlers and models, bed/location/spoke/
  shifting/transfer code; routes rewritten; migrations as named phases.
- Consumables catalogue seeded for a physiotherapy clinic (inventory rebuilt in M5).

### 00:10 · D2 clinic scoping, D3 Receptionist role
- `internal/clinic` Resolver; patients and referrals scoped by clinic
  (therapist inbox per clinic, no cross-clinic addressing or registration).
- Demo clinic seeded; demo users assigned; seeded doctor name fixed.
- `Receptionist` role (backend + frontend): registers/edits patients, no access
  to referrals; sidebar shows only what the role can open.
- Tests: clinic resolver, clinic isolation (unit + PostgreSQL integration),
  receptionist e2e journey; e2e waits for readiness after rollouts.

## 2026-10-20 — Klyniva rename, cleanup, local Kubernetes

### 23:15 · CI green on GitHub (PR #1)
- Fixed fresh-clone builds: git-ignored generated sources (`pluginMap.ts`,
  `supportedBrowsers.ts`) are now produced by `npm run generate`, used by
  `typecheck`, the Dockerfile and CI. All four CI jobs pass (backend,
  frontend, Docker images, e2e on docker compose).
- Verified locally: e2e suite green on the k3d cluster and on compose.

### 23:00 · Local Kubernetes (k3d) and end-to-end tests
- `k8s/` reorganised into `base/` + `overlays/local` (k3d) + `overlays/production`
  (nginx-ingress, cert-manager, HPA). Restricted Pod Security Standard,
  NetworkPolicy isolating PostgreSQL, Postgres as a StatefulSet, Redis removed.
- `scripts/k8s-local.sh` (`make k8s-up/deploy/e2e/status/down`): cluster
  `klyniva` with its own kubeconfig, images built and imported, generated
  secrets, app at http://localhost:8088.
- End-to-end API test suite `be/internal/e2e` (referral journey, permissions,
  token rules, clock in/out) runnable against compose (`make e2e-compose`), the
  cluster (`make k8s-e2e`) and in CI (new `e2e` job).
- TODO.md rewritten: mandatory testing policy, proposed roadmap M1–M7 and
  decisions D1–D5 for product-owner review.

### 22:45 · Rename to Klyniva and cleanup
- Go module `care-backend` → `klyniva`; namespace, images, compose project,
  DB name/user, Swagger metadata and UI titles renamed.
- Removed: `.kiro/` (relevant items moved into TODO.md M4), stray
  `wsl.localhost/`, 69 MB Go tarball, root `package.json`/`node_modules`,
  `be/docker-compose.yml`, one-off SQL fix scripts, disabled therapy code,
  outdated `docs/` and `be/docs/API_MAPPING.md`, CARE leftovers (Netlify,
  Crowdin, CodeClimate, devcontainer, CARE community docs, `.example.env`).
- Pushed over HTTPS with the `JM-Teixeira` account (`origin` now uses HTTPS).

## 2026-10-20 — Project foundation pass

### 22:30 · Docs, CI and tooling
- Added `CONTEXT.md` (architecture/conventions map), `TODO.md` (prioritised
  backlog), this `Completed.md`, and refreshed `README.md`.
- GitHub Actions CI (`.github/workflows/ci.yml`): gofmt, go vet, Go unit tests
  with race detector, PostgreSQL integration tests, frontend typecheck, strict
  lint of new modules, Vitest, Vite build, Docker image builds.
- Root `Makefile` (`make up / test / check / k8s-render`).
- `k8s/kustomization.yaml`; committed `k8s/secrets.yaml` renamed to
  `k8s/secrets.example.yaml` and the real file git-ignored.

### 22:25 · Containers & Kubernetes — `5b99103`
- Backend image: distroless, non-root (uid 65532), reproducible (`go.sum`),
  `server -healthcheck`; `/ready` endpoint with DB ping.
- Frontend image: build no longer fails (devDependencies installed; test deps
  present), nginx non-root on 8080 writing only `/tmp`, configurable API upstream.
- Compose: secrets from `.env`, no `trust` auth, Postgres 18, localhost-only DB
  ports, Redis/Adminer as optional profiles.
- k8s: probe/port/UID fixes, `DB_SSLMODE` for in-cluster DB, demo users off.
- Verified end-to-end with `docker compose` (referral lifecycle via nginx proxy,
  role scoping, token checks, clock-in).

### 22:25 · Frontend referral module — `38e9200`
- New screens: referral list / therapist inbox with status filters, referral
  detail (progress tracker, role-aware actions, doctor ↔ therapist conversation),
  three-step "new referral" form; sidebar entry; EN + PT strings.
- Vitest 5 + Testing Library set up in `fe/` (tests previously could not run);
  45 tests passing. Fixed a flaky fast-check property test (invalid dates).

### 22:08 · Backend core — `e09c2a4`, formatting `238aa74`
- `internal/modules/referral`: doctor → physiotherapist referrals with state
  machine, visibility rules, shared notes thread, atomic transitions.
- `internal/modules/patient`: list/search/create/get/update (were all 501) with
  visibility scoping and a typed update whitelist.
- Auth: access/refresh tokens are no longer interchangeable; refresh re-checks
  the account; middleware sets `user_id` (clock in/out always returned 401).
- Fixed `TimeLog.BeforeSave` hook signature; time-log/timetable tables now
  migrated; release-mode guards for JWT secret and demo users.
- Go 1.22 → 1.27.3, Gin 1.12, GORM 1.31; gofmt + LF normalisation (`.gitattributes`).
- Tests: unit (auth, config, referral, patient) + PostgreSQL integration tests
  via testcontainers, including a migration column-type regression test.

### 21:54 · Preserved in-progress work — `aaff097`
- Committed the previously uncommitted time-log, timetable and
  patient-creation work exactly as found, before any changes.
- Initial analysis of the repository (findings captured in `CONTEXT.md` and `TODO.md`).
