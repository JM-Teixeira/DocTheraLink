# CONTEXT — DocTheraLink / Klyniva

> Read this first. It is the short, always-current map of the project so you do
> not have to re-read the code. Update it whenever architecture, conventions or
> key decisions change. Work items live in `TODO.md`; finished work in
> `Completed.md`.

_Last updated: 2026-10-20_

## 1. What the product is

A clinic platform that **links doctors and physiotherapists** around shared
patients, plus general clinic operations (facilities, staff, scheduling,
assets). Working names: **DocTheraLink** (repo/backend) and **Klyniva**
(frontend branding, `fe/package.json`). Name not final — see TODO.

Primary market: **Portugal** (seed data uses Portuguese districts; UI ships in
`en` and `pt`). Health data ⇒ GDPR special-category data; design for privacy.

**Core workflow (implemented):** doctor creates a *referral* for a patient →
it lands in physiotherapists' inbox (or a named therapist) → therapist
accepts → starts treatment → completes with an outcome summary. Doctor and
therapist talk on a per-referral thread; every status change is logged there.

## 2. Repository layout

```
be/                Go backend (module name: care-backend)
  cmd/server       HTTP server entrypoint (also: `server -healthcheck`)
  cmd/migrate      migrations + seeding CLI
  internal/
    api/routes.go  all route registration (/api/v1)
    api/handlers/  legacy handlers (CARE-style; many endpoints still 501)
    modules/       NEW layered modules — use this pattern for new features
      referral/    model → repository → service → handler → routes (+ tests)
      patient/     same layout
    auth/          JWT (access/refresh with `typ` claim), middleware, role helpers
    config/        env config + Validate() (release-mode guards)
    database/      GORM connect, phased AutoMigrate, seed data
    integration/   PostgreSQL tests (build tag `integration`, testcontainers)
    models/        legacy shared GORM models
  scripts/         SQL helpers (init.sql is mounted into postgres)
fe/                React frontend (fork of open-source CARE frontend)
  src/components/Referral/   NEW referral UI module (+ tests)
  src/components/Patient/    patient UI (+ user's validation/api WIP with tests)
  src/Routers/routes/*.tsx   route tables (raviger)
  public/locale/{en,pt}.json translations
k8s/               Kubernetes manifests + kustomization.yaml (secrets NOT included)
docs/              older long-form docs (may be partly stale)
.github/workflows/ci.yml   CI
docker-compose.yml local stack;  .env.example → .env
Makefile           make up / test / check / k8s-render
```

## 3. Tech stack (versions as of 2026-10-20)

| Layer | Tech |
|---|---|
| Backend | Go 1.27.3, Gin 1.12, GORM 1.31 + pgx, golang-jwt v5, swaggo |
| DB | PostgreSQL 18 (compose, k8s, tests) |
| Frontend | React 18.3, TypeScript 5, Vite 7, Tailwind 3.4, TanStack Query 5, raviger router, i18next, shadcn/ui (Radix), lucide-react |
| Tests | Go `testing` + testcontainers-go; Vitest 5 + Testing Library + fast-check; Cypress (legacy E2E) |
| Runtime | Docker (distroless backend, nginx frontend — both non-root); Kubernetes manifests |
| Not used yet | Redis (configured, no client code) |

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
context. Roles are strings on `users.user_type` (`Doctor`, `PhysioTherapist`,
`Nurse`, `StateAdmin`, …). Admin = State/District/LocalBody admin.

**Visibility rules:**
- Referrals: admins all; doctors their own; therapists those assigned to them
  plus unassigned *pending* ones (inbox). Non-participants get 404, not 403.
- Patients: admins all; others → same home facility, created by / assigned to
  them, or linked to them through a referral.

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
```
Demo users (only when `SEED_DEMO_USERS=true`, never in release): `doctor`,
`physiotherapist`, `nurse`, `staff`, `stateadmin`, `districtadmin`,
`facilitycoordinator` — password `1234@1`.

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
- k8s: `kubectl apply -k k8s` after creating `doctheralink-secrets`
  (see `k8s/secrets.example.yaml`). Placeholder domains `doctheralink.com`.

## 7. Known gaps (details in TODO.md)

Most legacy endpoints (assets, consultations, inventory, shifting, transfer,
notes) still return 501; password reset is a stub; no audit log; no rate
limiting; frontend still has CARE-era screens calling unimplemented APIs;
inherited frontend code has ~4.5k (mostly formatting) lint errors.
