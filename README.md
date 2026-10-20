# DocTheraLink (Klyniva)

Clinic platform that links **doctors and physiotherapists** around shared
patients: doctors refer patients to physiotherapy, therapists accept and treat
them, and both sides follow progress on a shared thread — plus clinic
operations such as facilities, staff time tracking and timetables.

- **Backend:** Go 1.27 · Gin · GORM · PostgreSQL 18 (`be/`)
- **Frontend:** React · TypeScript · Vite · Tailwind · TanStack Query (`fe/`)
- **Deploy:** Docker Compose today, Kubernetes manifests in `k8s/`

| Document | What it is for |
|---|---|
| [CONTEXT.md](CONTEXT.md) | Architecture, conventions and key decisions — start here |
| [TODO.md](TODO.md) | Prioritised backlog |
| [Completed.md](Completed.md) | What has been done, with timestamps |
| [docs/](docs/) | Older long-form docs (partly outdated, see TODO) |

## Quick start (Docker)

Prerequisites: Docker with Compose v2.

```bash
cp .env.example .env      # set DB_PASSWORD and JWT_SECRET (openssl rand -base64 48)
make up                   # or: docker compose up -d --build --wait
```

- App: <http://localhost:4000> (nginx serves the SPA and proxies `/api/`)
- API: <http://localhost:9000> — Swagger UI at `/swagger/index.html` when `ENABLE_SWAGGER=true`
- Database UI (optional): `docker compose --profile tools up -d` → <http://localhost:8080>

Demo accounts (local only, `SEED_DEMO_USERS=true`), password `1234@1`:
`doctor`, `physiotherapist`, `nurse`, `staff`, `stateadmin`.

Try the core flow: log in as **doctor** → *Referrals* → *New referral* →
log in as **physiotherapist** → accept, start and complete it from the inbox.

## Development

```bash
# Backend (needs Go ≥1.21; go.mod pins toolchain 1.27.3, GOTOOLCHAIN=auto downloads it)
cd be && go run ./cmd/server          # expects Postgres; see be/.env.example

# Frontend
cd fe && npm install && npm run dev   # set REACT_CARE_API_URL=http://localhost:9000
```

## Tests

```bash
make test                 # Go unit tests + Vitest
make test-be-integration  # PostgreSQL integration tests (Docker required)
make check                # everything CI runs except image builds
```

CI runs on every push and pull request (`.github/workflows/ci.yml`).

## Deployment

- **Docker:** set `GIN_MODE=release`, a strong `JWT_SECRET`, `SEED_DEMO_USERS=false`
  and `PUBLIC_URL` in `.env`. Release mode refuses to start with weak settings.
- **Kubernetes:** create the secrets (see `k8s/secrets.example.yaml`), then
  `kubectl apply -k k8s`. Replace the placeholder domains and image tags first.

## Project layout

```
be/    Go API — new features live in be/internal/modules/<name>/
fe/    React app — new features live in fe/src/components/<Name>/
k8s/   Kubernetes manifests (kustomize)
docs/  Additional documentation
```

## License

MIT — see `fe/LICENSE`. The frontend started as a fork of the open-source
[CARE](https://github.com/ohcnetwork/care_fe) project.
