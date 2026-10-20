# Klyniva backend

Go REST API (Gin + GORM + PostgreSQL) under `/api/v1`. Architecture,
conventions and module layout are described in the root [CONTEXT.md](../CONTEXT.md).

```bash
make help              # list tasks
make test              # unit tests
make test-integration  # PostgreSQL tests via testcontainers (Docker required)
make run               # run locally; configure via environment or .env (see .env.example)
```

Endpoints of interest:

| Path | Purpose |
|---|---|
| `GET /health` | liveness (process up) |
| `GET /ready` | readiness (database reachable) |
| `/api/v1/auth/login/` | JWT login (access + refresh) |
| `/api/v1/patient/` | patients (module `internal/modules/patient`) |
| `/api/v1/referrals/` | doctor → physiotherapist referrals (module `internal/modules/referral`) |
| `/swagger/index.html` | API docs when `ENABLE_SWAGGER=true` |

New features go in `internal/modules/<name>/` (model → repository → service →
handler → routes, with tests at each layer).
