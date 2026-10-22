# Klyniva

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

## Open the local app

The development copy runs in a local Kubernetes cluster (k3d) on the Docker
engine **inside WSL (Ubuntu-22.04)**. It is only reachable from this computer.

1. Double-click `scriptsstart-klyniva.cmd` (from Explorer:
   `\wsl.localhostUbuntu-22.04ootdoctheraDocTheraLinkscripts`), or in a
   WSL terminal run `make k8s-start`.
2. Wait for "Klyniva is running"; the browser opens <http://localhost:8088>.
3. **Keep that window open** — Windows stops WSL (and the app) when nothing is
   attached to it, which shows as "this site can't be reached".

Demo logins (local demo data only), password **`1234@1`** for all:

Usernames follow one standard: `<clinic code>.<first name>.<last surname>`
(lower case, no accents), a number on a clash (`cm.joao.teixeira2`);
organisation admins have no clinic code. The server picks the username when
a person is added.

Demo clinic (code `de`):

| Username | Who | Can do |
|---|---|---|
| `de.john.smith` | Dr. John Smith | patients, create/follow referrals |
| `de.sarah.johnson` | Sarah Johnson | referral inbox, accept/treat/complete |
| `de.tiago.martins` | Tiago Martins | second therapist (hand-over) |
| `de.rita.lopes` | Rita Lopes | register/edit patients, assign patients on the time board |
| `de.marta.ribeiro` | Marta Ribeiro | plans divisions and physio shifts, sees physio load, inventory |
| `de.mary.williams`, `de.nuno.dias` | clinic staff | patients |
| `paula.neves` | administrator | everything, all clinics |

**Second demo clinic — Climutua** (Porto, everything in PT-PT: 120 patients,
diagnoses, referrals with messages, weekly schedules, sessions and stock),
same password `1234@1`:

| Username | Who |
|---|---|
| `cm.ana.ferreira` | Ana Ferreira, coordenadora |
| `cm.miguel.sousa` / `cm.helena.carvalho` / `cm.rui.teixeira` | médicos: fisiatra / ortopedista / neurologista |
| `cm.ines.moreira`, `cm.pedro.almeida`, `cm.catarina.lopes`, `cm.joao.ribeiro`, `cm.mariana.costa`, `cm.tiago.fonseca` | fisioterapeutas |
| `cm.sofia.martins` / `cm.carla.pinto` | receção |
| `cm.teresa.gomes` | enfermagem |
| `cm.rui.nogueira` | **left the clinic** (kept on record; cannot sign in) |

Stock permissions in Climutua: Teresa (nursing) changes everything, Inês the
floor-1 stock, Mariana the hygiene items of floors 2 and 3; the other physios
only look (the coordinator changes this from Pessoas → key button).

Climutua has three floors (gym and hydrotherapy on 1, rooms and electrotherapy
on 2, neuro rehab on 3) and
a planned holiday: Inês Moreira is away next week, covered by João Ribeiro,
with handover notes per patient (time board → *Ausências*).

The accounts are created by `SEED_DEMO_USERS=true` (local overlay/compose
only); production must run with it off. First time ever (no cluster yet):
`make k8s-up` in WSL (builds images, ~5 min).

Troubleshooting: `make k8s-status` shows the pods; `http://localhost:4000`
is the Docker Compose variant below and is not running unless you `make up`.

### Use it from another laptop on the same network

By default Klyniva answers only on this PC (`http://localhost:8088`). To open it
from other computers on your **home or clinic network**:

1. Start Klyniva as usual (`scripts\start-klyniva.cmd`).
2. Open **PowerShell as Administrator** in the repository folder and run:

   ```powershell
   powershell -ExecutionPolicy Bypass -File scripts\lan-access.ps1
   ```

   It forwards port 8088 into WSL, opens it in the Windows Firewall for
   **private networks only**, tells the backend its new address, and prints
   the addresses to use, e.g. `http://my-pc.local:8088` or `http://192.168.1.20:8088`.
3. On the other laptop open one of those addresses. If Windows marks your
   network as *Public*, switch it to *Private* (Settings → Network → Properties).

After a reboot the WSL address changes: run the script again. To stop sharing:
`scripts\lan-access.ps1 -Remove`. Demo logins share one password - fine for a
home network, not for shared ones; real patient data needs HTTPS first (TODO.md).

## Quick start (Docker Compose)

Prerequisites: Docker with Compose v2.

```bash
cp .env.example .env      # set DB_PASSWORD and JWT_SECRET (openssl rand -base64 48)
make up                   # or: docker compose up -d --build --wait
```

- App: <http://localhost:4000> (nginx serves the SPA and proxies `/api/`)
- API: <http://localhost:9000> — Swagger UI at `/swagger/index.html` when `ENABLE_SWAGGER=true`
- Database UI (optional): `docker compose --profile tools up -d` → <http://localhost:8080>

Same demo accounts as above (password `1234@1`).

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
make e2e-compose          # end-to-end API journeys against the running compose stack
make e2e-browser          # browser journeys (Playwright) against the local cluster
```

Every change needs tests — see the testing policy at the top of [TODO.md](TODO.md).

CI runs on every push and pull request (`.github/workflows/ci.yml`).

## Deployment

- **Docker:** set `GIN_MODE=release`, a strong `JWT_SECRET`, `SEED_DEMO_USERS=false`
  and `PUBLIC_URL` in `.env`. Release mode refuses to start with weak settings.
- **Local Kubernetes (k3d on Docker in WSL):** `make k8s-up` creates a cluster
  called `klyniva`, builds and imports the images and deploys
  `k8s/overlays/local` → <http://localhost:8088>. `make k8s-e2e` runs the
  end-to-end tests against it; `make k8s-start` brings it back after a
  reboot; `make k8s-down` removes it. The cluster uses
  its own kubeconfig (`~/.kube/klyniva.yaml`) and never changes your default context.
- **Production Kubernetes:** see the header of `k8s/overlays/production/kustomization.yaml`
  (create the secret, set domain/e-mail/image tags, then `kubectl apply -k k8s/overlays/production`).

## Project layout

```
be/    Go API — new features live in be/internal/modules/<name>/
fe/    React app — new features live in fe/src/components/<Name>/
k8s/   Kubernetes manifests (kustomize base + overlays)
scripts/  Local cluster tooling
```

## License

MIT — see `fe/LICENSE`. The frontend started as a fork of the open-source
[CARE](https://github.com/ohcnetwork/care_fe) project.
