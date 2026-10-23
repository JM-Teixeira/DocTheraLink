#!/usr/bin/env bash
# Klyniva on a small Ubuntu server (e.g. a laptop) with k3s.
# Run on the server, from the repository folder:
#
#   scripts/homeserver.sh install       # k3s (once); Docker must be installed
#   scripts/homeserver.sh deploy        # build, import and start / update Klyniva
#   scripts/homeserver.sh first-admin   # once: create admin.user, prints a one-time password
#   scripts/homeserver.sh status        # pods and the address to open
#   scripts/homeserver.sh backup-now    # a database backup now (also nightly)
#   scripts/homeserver.sh restore FILE  # put a backup back (asks first)
#
# Uses k8s/overlays/homeserver. Secrets are generated once into
# k8s/overlays/homeserver/secrets.env (git-ignored): keep a copy somewhere safe.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OVERLAY="$ROOT/k8s/overlays/homeserver"
KUBECTL=(sudo k3s kubectl -n klyniva)

log() { printf '\033[36m==>\033[0m %s\n' "$*"; }
need() { command -v "$1" >/dev/null || { echo "missing: $1 ($2)" >&2; exit 1; }; }

install() {
  need docker "install Docker first: https://docs.docker.com/engine/install/ubuntu/"
  if ! command -v k3s >/dev/null; then
    log "Installing k3s (single node, with Traefik ingress)"
    curl -sfL https://get.k3s.io | sh -
  fi
  log "k3s is ready: $(sudo k3s --version | head -1)"
}

ensure_config() {
  if [[ ! -s "$OVERLAY/secrets.env" ]]; then
    log "Creating secrets (keep a copy of $OVERLAY/secrets.env somewhere safe)"
    umask 077
    printf 'DB_PASSWORD=%s\nJWT_SECRET=%s\n' "$(openssl rand -hex 24)" "$(openssl rand -base64 48 | tr -d '\n')" >"$OVERLAY/secrets.env"
  fi
  # The names other computers use for this server (e-mail links, CORS)
  local host ips origins
  host="$(hostname).local"
  ips="$(hostname -I | tr ' ' '\n' | grep -E '^[0-9.]+$' | grep -v '^10\.42\.' || true)"
  origins="http://$host"
  for ip in $ips; do origins+=",http://$ip"; done
  printf 'PUBLIC_URL=http://%s\nCORS_ALLOWED_ORIGINS=%s\n' "$host" "$origins" >"$OVERLAY/site.env"
}

deploy() {
  need docker "see install"
  need k3s "run: scripts/homeserver.sh install"
  ensure_config
  log "Building images"
  docker build -q -t klyniva/backend:home "$ROOT/be"
  docker build -q -t klyniva/frontend:home "$ROOT/fe"
  log "Importing images into k3s"
  docker save klyniva/backend:home klyniva/frontend:home | sudo k3s ctr images import -
  log "Applying k8s/overlays/homeserver"
  sudo k3s kubectl apply -k "$OVERLAY"
  "${KUBECTL[@]}" rollout restart deployment/backend deployment/frontend
  "${KUBECTL[@]}" rollout status statefulset/postgres --timeout=300s
  "${KUBECTL[@]}" rollout status deployment/backend --timeout=300s
  "${KUBECTL[@]}" rollout status deployment/frontend --timeout=180s
  status
}

status() {
  "${KUBECTL[@]}" get pods
  echo
  echo "Open Klyniva on this network at:"
  echo "  http://$(hostname).local"
  for ip in $(hostname -I); do [[ "$ip" =~ ^[0-9.]+$ && ! "$ip" =~ ^10\.42\. ]] && echo "  http://$ip"; done
  echo "Backups: /var/backups/klyniva (nightly at 02:30)"
}

backup_now() {
  local job="postgres-backup-manual-$(date +%s)"
  "${KUBECTL[@]}" create job --from=cronjob/postgres-backup "$job"
  "${KUBECTL[@]}" wait --for=condition=complete "job/$job" --timeout=300s
  ls -lh /var/backups/klyniva | tail -3
}

first_admin() {
  "${KUBECTL[@]}" exec deployment/backend -- /app/migrate -create-admin
}

restore() {
  local file="${1:-}"
  [[ -f "$file" ]] || { echo "usage: $0 restore /var/backups/klyniva/klyniva-YYYYMMDD-HHMM.sql.gz" >&2; exit 2; }
  read -r -p "This replaces ALL current data with $file. Type RESTORE to continue: " ok
  [[ "$ok" == "RESTORE" ]] || { echo "cancelled"; exit 1; }
  "${KUBECTL[@]}" scale deployment/backend --replicas=0
  gunzip -c "$file" | "${KUBECTL[@]}" exec -i statefulset/postgres -- \
    sh -c 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d postgres -c "DROP DATABASE IF EXISTS $POSTGRES_DB" -c "CREATE DATABASE $POSTGRES_DB" && psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"'
  "${KUBECTL[@]}" scale deployment/backend --replicas=1
  log "Restored $file"
}

case "${1:-}" in
  install) install ;;
  deploy) deploy ;;
  status) status ;;
  backup-now) backup_now ;;
  first-admin) first_admin ;;
  restore) shift; restore "${1:-}" ;;
  *) sed -n '2,14p' "$0"; exit 2 ;;
esac
