# Hosting Klyniva

Today Klyniva runs in a k3d cluster inside WSL on the development PC. This
note is the plan to (1) move it to the Ubuntu laptop as a small home server,
(2) give it real HTTPS certificates, and (3) keep every step reusable so the
same setup later moves to a cloud provider and grows without big changes.
The open items are tracked in [TODO.md](../TODO.md) under "Hosting".

## Principles (what keeps it portable)

- **Kubernetes manifests are the contract.** The same `k8s/base` runs on k3d
  (dev), k3s (the laptop) and a managed cluster (cloud); only an overlay
  changes (`k8s/overlays/<place>`): host name, image tags, storage size,
  secrets. Nothing is installed by hand on the server besides k3s.
- **Images from a registry, tagged by git commit.** CI builds and pushes
  `backend` and `frontend` images (GHCR); servers pull a fixed tag. No
  building on the server.
- **Configuration from the environment, secrets outside git.** Every setting
  is an env var (see `.env.example`); secrets live in a k8s Secret created
  once per place (`k8s/overlays/*/secrets.env` is never committed).
- **Stateless app, state in two places only:** PostgreSQL and the uploaded
  files. That is what makes more replicas and moving servers easy.
- **The browser calls the API on its own address** (`same-origin`), so one
  image works behind any host name.

## 1. The Ubuntu laptop as a server (ready: k8s/overlays/homeserver)

What is ready in the repository:

- `k8s/overlays/homeserver`: k3s with its own ingress (Traefik), one
  backend replica (uploads on the local disk), no demo data, any host name,
  a nightly database backup to `/var/backups/klyniva` (14 days kept).
- `scripts/homeserver.sh`: `install`, `deploy`, `status`,
  `first-admin`, `backup-now`, `restore`.

Steps on the laptop (Ubuntu, Docker installed, repository cloned):

1. Router: give the laptop a fixed IP (DHCP reservation). Laptop: lid-close
   action "do nothing", automatic security updates, SSH with keys only.
2. `scripts/homeserver.sh install` (k3s, once).
3. `scripts/homeserver.sh deploy` (builds the images on the laptop, starts
   everything; run again after each `git pull` to update).
4. `scripts/homeserver.sh first-admin` prints `admin.user` and a one-time
   password; sign in and choose your own. Then create the clinic (with its
   code) and its people; each person chooses their password at their first
   sign-in.
5. Keep a copy of `k8s/overlays/homeserver/secrets.env` and copy
   `/var/backups/klyniva` to another disk or a NAS regularly; try
   `scripts/homeserver.sh restore <file>` on a test day.
6. Open `http://<laptop-name>.local` (or its IP) from the clinic computers.
   From outside the clinic only through a VPN (Tailscale / WireGuard), never
   by opening ports on the router.
7. A small UPS; k3s and Klyniva start by themselves after a power cut.

Still open before real patient data: HTTPS (section 2), and the GDPR items
in TODO.md (processing record, retention, export / erasure).

## 2. Certificates (HTTPS)

- **On the local network only:** a local certificate authority (mkcert, or
  step-ca for several machines). Install its root certificate on each laptop
  once; cert-manager issues and renews the server certificate inside k3s.
- **With a public domain:** cert-manager + Let's Encrypt. The DNS-01
  challenge works without opening any port (the DNS provider's API token is a
  k8s Secret), so it also works for a server reached only through the VPN.
- Once HTTPS works: HSTS, `Secure` cookies, and the refresh token moves to an
  httpOnly cookie (TODO.md, security).

## 3. Growing later (no big changes needed if these are done in order)

1. **Versioned SQL migrations** run as a k8s Job before each release instead
   of AutoMigrate at start-up (needed before more than one backend replica).
2. **Uploads to S3-compatible storage** (MinIO on the server, any S3 in the
   cloud) behind the existing `internal/storage.Store` interface; the
   backend then keeps no files.
3. **Managed PostgreSQL** (or CloudNativePG) with point-in-time recovery.
4. **More backend replicas** (`replicas: 2+`, a HorizontalPodAutoscaler);
   the API is already stateless (JWT, no server sessions).
5. **Observability:** structured logs, Prometheus metrics, error reporting.
6. **Several clinics:** already modelled (clinic codes, data scoped per
   clinic); a bigger organisation adds clinics, not servers.

## Moving to the cloud later

Same images, same base manifests: create a cluster, a managed database and an
S3 bucket, write `k8s/overlays/cloud` with their addresses, restore the last
database dump and copy the uploaded files. The DNS name moves; the app does
not change.
