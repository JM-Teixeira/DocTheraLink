# Completed — Klyniva

Log of finished work, newest first. Timestamps are local time (Europe/Lisbon,
UTC+01:00). Commits are on branch `chore/project-foundation` unless noted.

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
