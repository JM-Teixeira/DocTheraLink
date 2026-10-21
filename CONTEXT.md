# CONTEXT — Klyniva

> Read this first. It is the short, always-current map of the project so you do
> not have to re-read the code. Update it whenever architecture, conventions or
> key decisions change. Work items live in `TODO.md`; finished work in
> `Completed.md`.

_Last updated: 2026-10-20_

## 1. What the product is

**Klyniva** — a clinic platform that **links doctors and physiotherapists**
around shared patients, plus clinic operations (facilities, staff time
tracking, timetables). The product name is final; only the GitHub repository
is still called `DocTheraLink`.

Every change must ship with tests — see the testing policy at the top of `TODO.md`.

Primary market: **Portugal** (seed data uses Portuguese districts; UI ships in
`en` and `pt`). Health data ⇒ GDPR special-category data; design for privacy.

**Core workflow (implemented):** doctor creates a *referral* for a patient →
it lands in physiotherapists' inbox (or a named therapist) → therapist
accepts → starts treatment → completes with an outcome summary. Doctor and
therapist talk on a per-referral thread; every status change is logged there.

## 2. Repository layout

```
be/                Go backend (Go module name: klyniva)
  cmd/server       HTTP server entrypoint (also: `server -healthcheck`)
  cmd/migrate      migrations + seeding CLI
  internal/
    api/routes.go  all route registration (/api/v1)
    api/handlers/  legacy handlers (users, facilities, time log, timetable;
                   inventory endpoints are still 501 stubs)
    modules/       NEW layered modules — use this pattern for new features
      referral/    model → repository → service → handler → routes (+ tests)
      patient/     same layout
      notification/ in-app + e-mail notifications fed by referral events
      attachment/  files on referrals; access follows the referral
      catalog/     embedded ICD-10 subset + referral templates (EN/PT, read-only)
    storage/       file Store interface (Local today, S3 later)
    mailer/        SMTP or log-only mailer (no patient data in e-mails)
    clinic/        which clinic a user works in (Resolver) — scope new data with it
    auth/          JWT (access/refresh with `typ` claim), middleware, role helpers
    config/        env config + Validate() (release-mode guards)
    database/      GORM connect, phased AutoMigrate, seed data
    integration/   PostgreSQL tests (build tag `integration`, testcontainers)
    e2e/           HTTP tests against a RUNNING stack (build tag `e2e`, E2E_BASE_URL)
    models/        legacy shared GORM models
  scripts/         init.sql (compose), portugal_data.sql (reference data)
fe/                React frontend (started as a fork of CARE; ICU modules removed)
  src/components/Referral/   referral UI module (+ tests)
  src/components/Patient/    Klyniva patient pages (PatientList/Detail/Form,
                             patientFormUtils) + hooks in api/ (+ tests)
  src/Utils/roles.ts         role helpers used by every module's UI
  src/Utils/person.ts        initials / age helpers
  src/Routers/routes/*.tsx   route tables (raviger)
  public/locale/{en,pt}.json translations
k8s/base/          environment-independent manifests (kustomize)
k8s/overlays/local       k3d cluster (http://localhost:8088, demo data)
k8s/overlays/production  nginx-ingress + cert-manager TLS + HPA template
scripts/k8s-local.sh     create/deploy/e2e/status/down for the local cluster
.github/workflows/ci.yml CI (unit, integration, e2e on compose, image builds)
docker-compose.yml local stack;  .env.example → .env
Makefile           make up / test / check / k8s-up / k8s-e2e …
```

## 3. Tech stack (versions as of 2026-10-20)

| Layer | Tech |
|---|---|
| Backend | Go 1.27.3, Gin 1.12, GORM 1.31 + pgx, golang-jwt v5, swaggo |
| DB | PostgreSQL 18 (compose, k8s, tests) |
| Frontend | React 18.3, TypeScript 5, Vite 7, Tailwind 3.4, TanStack Query 5, raviger router, i18next, shadcn/ui (Radix), lucide-react |
| Tests | Go `testing` + testcontainers-go; Vitest 5 + Testing Library + fast-check; Cypress (legacy E2E) |
| Runtime | Docker (distroless backend, nginx frontend — both non-root); Kubernetes manifests |
| Not used | Redis (config keys exist, no client; removed from compose default and k8s) |

## 4. Architecture & conventions

**Backend module pattern** (`internal/modules/<name>/`):
- `model.go` GORM models + request DTOs; `Models()` lists tables for migration.
- `repository.go` `Repository` interface + GORM impl. Only layer touching SQL.
- `service.go` business rules: permissions, validation, state transitions.
  Depends only on the interface → unit-testable with an in-memory fake.
- `handler.go` HTTP ↔ service, maps errors (`ValidationError`→400,
  `ErrForbidden`→403, `ErrNotFound`→404, `ErrInvalidTransition`/`ErrConflict`→409,
  else 500 with details only in logs).
- `routes.go` `RegisterRoutes(group, db)`.
- Register tables in `database.Migrate()`; routes in `api/routes.go`.

**Gotcha — GORM relations to other modules' tables:** read-only "summary"
structs (e.g. `referral.PatientSummary`) must use `gorm:"...;-:migration"` on
the relation field. Otherwise AutoMigrate rewrites the referenced table from
the summary struct (it once turned `patients.id` from uuid into text). A
regression test checks column types.

**Auth:** `Authorization: Bearer <access>`. Middleware sets
`authorization_payload` (`*models.TokenClaims`) and `user_id` (uint) in the Gin
context. Roles are strings on `users.user_type`. Admin = State/District/LocalBody
admin (organisation-wide, no clinic).

**Roles in use:** `Doctor` (refers), `PhysioTherapist` (treats), `Nurse`,
`Receptionist` (front desk: registers/edits patients, appointments in M3, **no**
referrals/clinical notes), `Staff`, read-only variants, admins. Frontend checks
live in `fe/src/Utils/roles.ts` and mirror the backend.

**Clinics (decision D2):** single clinic today, built for many. A clinic is a
row in `facilities`; a user's `home_facility_id` is their clinic. Modules get a
`clinic.Resolver` and stamp/scope data by clinic (`clinic.SameClinic` treats an
unset clinic as compatible, so clinic-less setups still work). Seeded demo
clinic: `database.DemoClinicID` ("Klyniva Demo Clinic"), all non-admin demo
users work there.

**Visibility rules:**
- Referrals: admins all; doctors their own; therapists those assigned to them
  plus unassigned *pending* ones **of their clinic** (inbox). A referral takes
  the doctor's clinic; an addressed therapist must work in it.
  Non-participants get 404, not 403. Receptionists/nurses: 403 on list.
- Patients: admins all; others → patients of their clinic, created by /
  assigned to them, or linked to them through a referral. Non-admins cannot
  register patients into another clinic.

**Removed (decision D1):** CARE's ICU features — beds/locations, assets,
consultations/daily rounds, prescriptions, shifting, resource requests,
camera/vitals, HCX insurance, death report — in both frontend and backend.
Inventory is kept for consumables and will be rebuilt as a module (M5).

**Referral state machine:** `pending → accepted → in_progress → completed`;
`pending → declined` (assigned therapist, reason required);
any open state `→ cancelled` (referring doctor/admin, reason required).
Transitions are compare-and-swap (`UPDATE … WHERE status = <from>`), so
concurrent accepts have one winner.

**Frontend conventions:** new code uses TanStack `useQuery`/`useMutation` with
`query()`/`mutate()`/`callApi()` from `@/Utils/request/*` and route objects
(`Type<T>()`), `useAuthUser()` for the current user, `Page` wrapper,
`Notifications` for toasts, all strings via `t()` in both locale files.
Do not use `components/ui/input.tsx` for controlled inputs (it omits `onChange`).
Global CSS styles bare `h2/h3` large — give headings explicit text sizes.
Pure logic goes in `*Utils.ts` with tests; UI permissions mirror the backend.

## 5. Running & testing

```bash
cp .env.example .env         # set DB_PASSWORD, JWT_SECRET
make up                      # http://localhost:4000 (API also on :9000)
make test                    # Go unit + Vitest
make check                   # + gofmt/vet, Postgres integration tests, typecheck, lint
make e2e-compose             # e2e API journeys against the compose stack

make k8s-up                  # local Kubernetes (k3d) → http://localhost:8088
make k8s-e2e                 # same e2e journeys against the cluster
make k8s-deploy | k8s-status | k8s-down
```
The local cluster's kubeconfig is `~/.kube/klyniva.yaml` (context
`k3d-klyniva`); the script never touches `~/.kube/config`, which holds real
cloud clusters — never run kubectl against those from this project.
Demo users (only when `SEED_DEMO_USERS=true`, never in release): `doctor`,
`physiotherapist`, `nurse`, `receptionist`, `staff` (all in the demo clinic),
`stateadmin`, `districtadmin`, `facilitycoordinator` — password `1234@1`.

Go toolchain: go.mod pins `toolchain go1.27.3`; with an older local Go use
`GOTOOLCHAIN=auto`. In this dev environment the repo lives in WSL
(`/root/docthera/DocTheraLink`); run git/go/npm inside WSL (Windows git rejects
the repo ownership).

## 6. Deployment facts

- Images: backend `gcr.io/distroless/static:nonroot` (uid 65532, port 9000,
  `/health` liveness, `/ready` DB readiness); frontend nginx uid 101, port
  **8080**, writes only `/tmp`, proxies `/api/` to `BACKEND_UPSTREAM`.
- Runtime frontend config is generated into `/config.js` from env
  (`REACT_CARE_API_URL` = browser-facing API origin).
- Release mode (`GIN_MODE=release`) refuses a default/short `JWT_SECRET` and
  demo users. Migrations run automatically at backend start.
- k8s base: namespace `klyniva` with the *restricted* Pod Security Standard,
  postgres StatefulSet, services `postgres`/`backend`/`frontend`, one Ingress
  (`/api` → backend, `/` → frontend), NetworkPolicy so only the backend reaches
  postgres. Config via `klyniva-config` (configMapGenerator); secrets
  `klyniva-secrets` (DB_PASSWORD, JWT_SECRET) come from the overlay
  (local: generated `secrets.env`; production: create with kubectl).
- Backend stays at 1 replica until migrations run as a Job and uploads move
  to shared storage. Production placeholders: `klyniva.com`, GHCR image names.

## 7. Known gaps (details in TODO.md)

Inventory endpoints still return 501 (rebuilt in M5); attachments live on
the backend's PVC (`UPLOAD_PATH`), so S3 is needed before scaling out;
password reset is a stub; no audit log; no rate limiting; inherited frontend
code still has many (mostly formatting) lint errors, so CI lints new code via
`npm run lint:strict`. Existing databases keep the tables of removed CARE
modules until versioned migrations drop them.
