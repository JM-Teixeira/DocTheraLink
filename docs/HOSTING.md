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

## 1. The Ubuntu laptop as a server

1. Ubuntu Server (or Desktop with the lid-close action set to "do nothing"),
   fixed IP from the router (DHCP reservation), automatic security updates
   (`unattended-upgrades`), SSH with keys only.
2. Install **k3s** (single node, includes Traefik ingress and local storage):
   `curl -sfL https://get.k3s.io | sh -`.
3. Create `k8s/overlays/homeserver` (copy of `production` with the laptop's
   host name, 2 backend replicas off for now, storage on the local disk).
4. Create the secret, `kubectl apply -k k8s/overlays/homeserver`.
5. Database: the PostgreSQL StatefulSet with a nightly `pg_dump` CronJob to a
   second disk or a NAS, and a restore test once a month.
6. Reachability: on the clinic/home network by name (`klyniva.local` via the
   router's DNS or mDNS). From outside only through a VPN (Tailscale or
   WireGuard), never by opening the port on the router.
7. Power: a small UPS; k3s and the app start by themselves after a power cut.

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
