# Klyniva

Clinic platform for **physiotherapy clinics and the doctors who refer to
them**: doctors refer patients, physiotherapists accept, plan and treat them,
both follow progress on a shared thread, and the clinic runs its day (time
board, waiting list, absences, stock, people) in one place. Portuguese
(PT-PT) and English.

- **Backend:** Go · Gin · GORM · PostgreSQL (`be/`)
- **Frontend:** React · TypeScript · Vite · Tailwind · TanStack Query (`fe/`)
- **Runs on:** Kubernetes (k3d locally, any k8s in production) or Docker Compose

| Document | What it is for |
|---|---|
| [CONTEXT.md](CONTEXT.md) | Architecture, conventions and key decisions: start here |
| [TODO.md](TODO.md) | Prioritised backlog (and the testing policy) |
| [Completed.md](Completed.md) | What has been done, version by version |
| [docs/HOSTING.md](docs/HOSTING.md) | Moving to a home server, HTTPS, scaling |

## What it does

- **Patients:** registration with drafts and retry, allergies / conditions /
  medication lists, GDPR consents, history timeline, metrics and CSV export.
- **Referrals:** doctor → physio, treatment plan per referral (steps of each
  session), questions and progress with pain scores, files.
- **Time board:** shifts per physio and room, sessions by drag and drop,
  weekly plans, absences with patient moves and handover, waiting list with
  free-slot alerts, missed sessions.
- **Clinic:** stock with per-floor / per-category permissions, people with
  roles and work history, clinics with codes.
- **For everyone:** settings (theme, text size, language, start page),
  suggestions to the admins, "?" help on every page; admins can view the app
  as any role (read-only).

## Open the local app (Windows + WSL)

The development copy runs in a local Kubernetes cluster (k3d) on the Docker
engine inside WSL (Ubuntu-22.04).

1. Double-click `scripts\start-klyniva.cmd` (in Explorer:
   `\\wsl.localhost\Ubuntu-22.04\root\docthera\DocTheraLink\scripts`), or
   run `make k8s-start` in a WSL terminal.
2. Wait for "Klyniva is running"; the browser opens <http://localhost:8088>.
3. **Keep that window open**: Windows stops WSL (and the app) when nothing is
   attached to it ("this site can't be reached").

First time ever (no cluster yet): `make k8s-up` in WSL (builds the images,
about 5 minutes). `make k8s-status` shows the pods.

### Demo accounts

Local demo data only (`SEED_DEMO_USERS=true`), password **`1234@1`** for all.
Usernames follow one standard, applied by the server when a person is added:
`<clinic code>.<first name>.<last surname>`, lower case without accents, with a
number on a clash (`cm.joao.teixeira2`); organisation admins have no code.

**Climutua** (code `cm`, Porto, PT-PT data: 120 patients, referrals, weekly
schedules, sessions, stock):

| Username | Who |
|---|---|
| `cm.ana.ferreira` | coordenadora |
| `cm.miguel.sousa`, `cm.helena.carvalho`, `cm.rui.teixeira` | médicos (fisiatra, ortopedista, neurologista) |
| `cm.ines.moreira`, `cm.pedro.almeida`, `cm.catarina.lopes`, `cm.joao.ribeiro`, `cm.mariana.costa`, `cm.tiago.fonseca` | fisioterapeutas |
| `cm.sofia.martins`, `cm.carla.pinto` | receção |
| `cm.teresa.gomes` | enfermagem (changes all stock) |
| `cm.rui.nogueira` | left the clinic (kept on record, cannot sign in) |

**Demo clinic** (code `de`, English names): `de.john.smith` (doctor),
`de.sarah.johnson` and `de.tiago.martins` (physios), `de.rita.lopes` (front
desk), `de.marta.ribeiro` (coordinator), `de.mary.williams`, `de.nuno.dias`.

**Administrator:** `paula.neves` (all clinics; user menu → "Ver como…" to see
the app as any role).

### Use it from other laptops on the same network

1. Start Klyniva as usual.
2. In **PowerShell as Administrator**, in the repository folder:

   ```powershell
   powershell -ExecutionPolicy Bypass -File scripts\lan-access.ps1
   ```

   It forwards port 8088 into WSL, opens it in the Windows Firewall for
   **private networks only**, and prints the addresses to use, e.g.
   `http://my-pc.local:8088`.
3. Open that address on the other laptop. If Windows marks the network as
   *Public*, switch it to *Private*.

The app always talks to the server on the address it was opened from, so no
other setting is needed. After a reboot run the script again;
`scripts\lan-access.ps1 -Remove` stops sharing. Plain HTTP: fine on a home
network, real patient data needs HTTPS first (docs/HOSTING.md).

## Docker Compose (alternative)

```bash
cp .env.example .env      # set DB_PASSWORD and JWT_SECRET (openssl rand -base64 48)
make up                   # or: docker compose up -d --build --wait
```

App on <http://localhost:4000> (nginx serves the app and proxies `/api/`),
API on <http://localhost:9000>. Same demo accounts.

## Development

```bash
cd be && go run ./cmd/server          # backend (Postgres needed; see .env.example)
cd fe && npm install && npm run dev   # frontend
```

Go: `go.mod` pins the toolchain; with an older Go use `GOTOOLCHAIN=auto`.

## Tests

Every change needs tests (policy at the top of [TODO.md](TODO.md)).

```bash
make test                 # Go unit tests + Vitest (frontend unit/component)
make test-be-integration  # PostgreSQL integration tests (Docker required)
make k8s-e2e              # API journeys against the local cluster
make e2e-browser          # browser journeys (Playwright) against the local cluster
make check                # what CI runs, without image builds
```

CI (`.github/workflows/ci.yml`) runs all of them on every push, including a
coverage gate of 80 % on the patient screens.

## Repository layout

```
be/                     Go API
  cmd/server            entry point
  internal/modules/     one folder per feature: model, repository, service, handler, routes, tests
                        (patient, referral, attachment, schedule, inventory, staff,
                         timeline, feedback, notification, catalog)
  internal/api          router and the older handlers (users, clinics)
  internal/auth         JWT, roles, "view as", active-account check
  internal/database     migrations and demo data
  internal/e2e          API journeys (build tag e2e)
fe/                     React app
  src/components/<Name> one folder per feature (pages, dialogs, api.ts, pure *Utils.ts + tests)
  src/Utils, src/hooks  shared helpers (requests, dates, settings, view-as)
  public/locale         translations (en.json, pt.json)
e2e/browser             Playwright journeys
k8s/                    Kubernetes manifests (kustomize base + overlays: local, production)
scripts/                local cluster, LAN access, version bump
```

## Deployment

- **Production Kubernetes:** see the header of
  `k8s/overlays/production/kustomization.yaml`, and docs/HOSTING.md for a
  home server and the path to scale.
- **Docker:** `GIN_MODE=release`, a strong `JWT_SECRET`, `SEED_DEMO_USERS=false`
  and `PUBLIC_URL` in `.env`; release mode refuses weak settings.
- Versions: one number for app and server (`scripts/bump-version.sh 0.x.y`).

## License

MIT, see `fe/LICENSE`. The frontend started as a fork of the open-source
[CARE](https://github.com/ohcnetwork/care_fe) project.
