# Completed — DocTheraLink / Klyniva

Log of finished work, newest first. Timestamps are local time (Europe/Lisbon,
UTC+01:00). Each entry links to the commit(s) on branch
`chore/project-foundation` unless noted.

---

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
