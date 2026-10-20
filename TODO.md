# TODO — DocTheraLink / Klyniva

Prioritised backlog. When an item is done, move it to `Completed.md` with a
timestamp. Context for every item is in `CONTEXT.md`.

Priority: **P0** blocks any real-patient use · **P1** core product ·
**P2** important · **P3** nice to have.

---

## P0 — Security, privacy & compliance (before any real patient data)

- [ ] **Decide the product name** (DocTheraLink vs Klyniva) and apply it everywhere
      (Go module `care-backend`, Swagger `@title CARE Backend API`, contact
      `ohc.network` placeholders, image names, k8s namespace, locale strings).
- [ ] **GDPR / health-data baseline**: data-processing record, legal basis,
      retention policy, patient data export & erasure endpoints, DPA with hosting
      provider; consider CNPD guidance (Portugal).
- [ ] **Audit log**: append-only record of who read/changed which patient and
      referral (required for health data). Table + middleware + admin viewer.
- [ ] **Password reset**: `ForgotPassword`/`CheckResetToken`/`ResetPassword`
      are stubs that always return success — implement signed, expiring tokens
      + email delivery (SMTP config already exists).
- [ ] **Login rate limiting / lockout** (config has `RATE_LIMIT_*`, unused).
- [ ] **Token storage**: frontend keeps JWTs in `localStorage`; evaluate
      httpOnly refresh cookie + short-lived access token; add refresh-token
      revocation (logout, password change).
- [ ] **CORS & headers**: restrict `CORS_ALLOWED_ORIGINS` per env; add
      `Cache-Control: no-store` on API responses; tighten nginx CSP
      (currently allows `http:`/`https:`/`'unsafe-inline'`).
- [ ] **Access control review of legacy handlers** (`api/handlers/*`): user,
      facility and timetable endpoints do little or no per-facility/role
      scoping (e.g. `users.GET /` lists everyone). Move to module pattern with tests.
- [ ] **Encrypt backups and DB at rest**; TLS to a managed Postgres
      (`DB_SSLMODE=verify-full`).
- [ ] **Dependency scanning in CI** (govulncheck, `npm audit`/Snyk, Trivy for images).

## P1 — Core product (doctor ↔ physiotherapist link and beyond)

### Referrals (MVP shipped — extend)
- [ ] Link "Refer to physiotherapy" button from the patient page
      (`/referrals/new?patient=<id>` already supported).
- [ ] Notifications on referral events (new referral to therapist inbox,
      accepted/declined/completed to doctor): in-app + email; later push.
- [ ] Attachments on referrals (imaging, reports) via the existing `/files` API
      — needs implementing (currently 501).
- [ ] Re-assign / hand-over a referral to another therapist; re-open after decline.
- [ ] Filters: by patient, priority, date range; full-text search; counts per status tab.
- [ ] Referral templates per common diagnosis (ACL rehab, low back pain…).
- [ ] ICD-10/ICD-11 code picker (an `/icd/` endpoint exists but is 501).

### Physiotherapy treatment
- [ ] **Treatment plan** per referral: goals, planned sessions, frequency,
      modalities (electro, hydro, massage, gym — matches existing facility sections).
- [ ] **Session notes** (SOAP format), pain scale (VAS/NRS), ROM / strength
      measurements with charts of progress over time.
- [ ] **Home exercise programme** builder; printable / shareable with patient.
- [ ] Re-enable / rebuild the disabled therapy models (`models/therapy.go`,
      `notes_disabled.go`, `handlers/therapy.go.disabled`) on the module pattern.

### Scheduling
- [ ] Appointment booking linked to referrals and therapist timetables
      (timetable/sections + time-log APIs exist; add appointments, conflicts,
      room/section capacity, cancellations, no-shows).
- [ ] Calendar views (day/week per therapist and per section).
- [ ] Reminders (SMS/email) to patients.

### Patients
- [ ] Wire the user's "centralized patient creation" spec
      (`.kiro/specs/centralized-patient-creation`) to the new `/patient` API;
      the frontend form model (demographics/emergency contact/medical history)
      is richer than the backend model — extend the backend model + migration.
- [ ] Patient timeline: referrals, sessions, notes, documents in one view.
- [ ] Duplicate detection on create (phone + DOB + name similarity).
- [ ] Portuguese identifiers (NIF, SNS number) with validation.

### Clinic operations (from inherited CARE scope — decide what to keep)
- [ ] Facilities: finish CRUD + users per facility (many endpoints 501).
- [ ] Staff time tracking: clock in/out works now; add admin reports/exports.
- [ ] Decide on and either implement or remove: assets, inventory, beds,
      shifting, consultations/daily rounds, resource requests, HCX, TeleICU.
      Remove dead frontend screens to shrink the bundle and confusion.
- [ ] Dashboards: referrals per status, waiting time to first session,
      therapist load, outcomes.
- [ ] Billing / insurance (SNS, ADSE, private insurers, self-pay) — later.

## P2 — Engineering quality

- [ ] **Versioned SQL migrations** (goose or Atlas) instead of AutoMigrate at
      startup; run as a k8s Job / init container (two replicas currently migrate concurrently).
- [ ] Structured logging (`log/slog`, JSON in release), request IDs, and no
      SQL logging of patient data in debug.
- [ ] OpenAPI: regenerate Swagger (`make swagger` in be/) and generate the
      frontend API types from it.
- [ ] Frontend: normalise line endings + run Prettier/ESLint `--fix` across
      `fe/src` (~4.5k mostly-formatting errors), then lint all of `src/` in CI.
- [ ] Frontend upgrades: React 19, Tailwind 4, ESLint 9 flat config,
      @tanstack/react-query latest; replace deprecated `useTanStackQueryInstead`.
- [ ] Root `package.json`/`node_modules` duplicate the frontend test deps — remove
      once nobody relies on them.
- [ ] Remove `be/go1.22.3.linux-amd64.tar.gz` (69 MB, git-ignored) and the stray
      empty `wsl.localhost/` folder from the working tree (deletion was not
      permitted from the assistant session).
- [ ] Remove `be/docker-compose.yml` (superseded by the root compose file) and
      refresh `docs/` (several pages describe the old setup).
- [ ] Raise test coverage: legacy handlers (time log, timetable, user, facility),
      frontend pages (ReferralDetail/List with mocked API), Cypress E2E for the
      referral flow.
- [ ] Fix seeded doctor first name `"Dr. John"` (UI now avoids "Dr. Dr.").
- [ ] Legacy model relations use `UserBareMinimum` without a table name —
      GORM may create a stray `user_bare_minimums` table; audit models.

## P3 — Deployment & operations

- [ ] Container registry + CD (GitHub Actions → GHCR, tag by git SHA; pin
      `images:` in `k8s/kustomization.yaml`).
- [ ] Helm chart or kustomize overlays (dev/staging/prod).
- [ ] Managed PostgreSQL in prod (or CloudNativePG operator) with PITR backups.
- [ ] Shared storage for uploads (S3-compatible) — the RWX PVC in
      `k8s/backend.yaml` needs a storage class that supports it.
- [ ] Observability: Prometheus metrics endpoint, OpenTelemetry traces,
      Sentry DSN wiring (frontend already supports `REACT_SENTRY_DSN`).
- [ ] Rootless nginx: switch to `nginxinc/nginx-unprivileged` or keep current
      setup; add `seccompProfile: RuntimeDefault` and disable SA token automount.
- [ ] Redis: either use it (rate limiting, sessions, caching) or remove it from k8s.
