#!/usr/bin/env bash
# Local Kubernetes for Klyniva using k3d (k3s in Docker).
#
#   scripts/k8s-local.sh up       create cluster (if needed), build, import, deploy
#   scripts/k8s-local.sh start    start an existing cluster (e.g. after a reboot) and wait until it answers
#   scripts/k8s-local.sh deploy   rebuild images and roll out (cluster must exist)
#   scripts/k8s-local.sh e2e      run the end-to-end API tests against the cluster
#   scripts/k8s-local.sh status   show pods, services and ingress
#   scripts/k8s-local.sh down     delete the cluster (all its data)
#   scripts/k8s-local.sh lan H,.. tell the backend its names on the network ("off" = localhost only);
#                                 used by scripts/lan-access.ps1 (Windows side)
#
# Safety: the cluster's kubeconfig lives in its own file (~/.kube/klyniva.yaml)
# and every kubectl call passes it explicitly, so other contexts (e.g. cloud
# clusters in ~/.kube/config) are never used or modified.
set -euo pipefail

CLUSTER=klyniva
PORT=${KLYNIVA_PORT:-8088}
URL="http://localhost:${PORT}"
ROOT=$(cd "$(dirname "$0")/.." && pwd)
OVERLAY="$ROOT/k8s/overlays/local"
export KUBECONFIG="$HOME/.kube/klyniva.yaml"
KUBECTL=(kubectl --kubeconfig "$KUBECONFIG" --context "k3d-$CLUSTER" -n klyniva)

log() { printf '\033[36m==>\033[0m %s\n' "$*"; }

cluster_exists() { k3d cluster list -o json | grep -q "\"name\":\"$CLUSTER\""; }

# k3s needs many inotify instances; the Docker/WSL kernel defaults to 128,
# which is exhausted when several k3d clusters run ("too many open files" and
# the node never registers). Raise it (only if lower) — not persistent across
# WSL/Docker restarts, so this runs on every up/start/deploy.
ensure_inotify() {
  local want=1024
  local cur
  cur=$(docker run --rm alpine:3 cat /proc/sys/fs/inotify/max_user_instances)
  if [[ "$cur" -lt "$want" ]]; then
    log "Raising fs.inotify.max_user_instances in the Docker VM ($cur -> $want)"
    docker run --rm --privileged alpine:3 sysctl -w fs.inotify.max_user_instances=$want >/dev/null
  fi
}

create_cluster() {
  if cluster_exists; then
    log "Cluster $CLUSTER already exists"
  else
    log "Creating k3d cluster $CLUSTER (ingress on localhost:$PORT)"
    k3d cluster create "$CLUSTER" --agents 0 -p "${PORT}:80@loadbalancer" \
      --kubeconfig-update-default=false --kubeconfig-switch-context=false --wait
  fi
  mkdir -p "$(dirname "$KUBECONFIG")"
  k3d kubeconfig get "$CLUSTER" > "$KUBECONFIG"
  chmod 600 "$KUBECONFIG"
}

ensure_secrets() {
  if [[ ! -f "$OVERLAY/secrets.env" ]]; then
    log "Generating $OVERLAY/secrets.env (git-ignored)"
    {
      echo "DB_PASSWORD=$(openssl rand -hex 16)"
      echo "JWT_SECRET=$(openssl rand -hex 32)"
    } > "$OVERLAY/secrets.env"
    chmod 600 "$OVERLAY/secrets.env"
  fi
}

build_and_import() {
  log "Building images"
  docker build -q -t klyniva/backend:dev "$ROOT/be"
  docker build -q -t klyniva/frontend:dev "$ROOT/fe"
  log "Importing images into the cluster"
  k3d image import -c "$CLUSTER" klyniva/backend:dev klyniva/frontend:dev
}

deploy() {
  ensure_secrets
  log "Applying k8s/overlays/local"
  kubectl --kubeconfig "$KUBECONFIG" --context "k3d-$CLUSTER" apply -k "$OVERLAY"
  # The :dev tag does not change, so restart to pick up freshly imported images
  "${KUBECTL[@]}" rollout restart deployment/backend deployment/frontend
  log "Waiting for rollout"
  "${KUBECTL[@]}" rollout status statefulset/postgres --timeout=180s
  "${KUBECTL[@]}" rollout status deployment/backend --timeout=180s
  "${KUBECTL[@]}" rollout status deployment/frontend --timeout=120s
  [[ -s "$LAN_FILE" ]] && apply_lan
  log "Ready: $URL"
}

# Names other computers use to reach this PC (saved by "lan", re-applied on deploy)
LAN_FILE="$HOME/.klyniva-lan"
apply_lan() {
  local origins="http://localhost:${PORT}" public="$URL" h
  if [[ -s "$LAN_FILE" ]]; then
    IFS=, read -ra hs < "$LAN_FILE" || true # the file has no final newline
    public="http://${hs[0]}:${PORT}"
    for h in "${hs[@]}"; do origins+=",http://${h}:${PORT}"; done
  fi
  log "Backend address: $public"
  "${KUBECTL[@]}" set env deployment/backend PUBLIC_URL="$public" CORS_ALLOWED_ORIGINS="$origins" >/dev/null
  "${KUBECTL[@]}" rollout status deployment/backend --timeout=180s >/dev/null
}
lan() {
  case "${1:-}" in
    "") echo "usage: $0 lan <name,ip,...>|off" >&2; exit 2 ;;
    off) rm -f "$LAN_FILE" ;;
    *) [[ "$1" =~ ^[A-Za-z0-9.,-]+$ ]] || { echo "bad host list: $1" >&2; exit 2; }
       printf "%s" "$1" > "$LAN_FILE" ;;
  esac
  apply_lan
}
print_access() {
  cat <<INFO

  Klyniva is running: $URL   (open it in a browser on this computer)

  Demo logins (local demo data only) — password for all: 1234@1
    doctor            Dr. John Smith (refers patients)
    physiotherapist   Sarah Johnson
    physiotherapist2  Tiago Martins
    receptionist      Rita Lopes (patients, assigns patients on the time board)
    coordinator       Marta Ribeiro (time board: divisions, shifts, physio load)
    nurse, staff      other clinic staff
    stateadmin        administrator (all clinics)

INFO
}

# start: bring an existing cluster back after WSL/Docker restarted.
start() {
  cluster_exists || { echo "No cluster yet — run: $0 up" >&2; exit 1; }
  ensure_inotify
  k3d cluster start "$CLUSTER" >/dev/null
  k3d kubeconfig get "$CLUSTER" > "$KUBECONFIG"
  log "Waiting for the app (can take 1–2 minutes after a restart)"
  "${KUBECTL[@]}" rollout status statefulset/postgres --timeout=300s >/dev/null
  "${KUBECTL[@]}" rollout status deployment/backend --timeout=300s >/dev/null
  "${KUBECTL[@]}" rollout status deployment/frontend --timeout=300s >/dev/null
  for _ in $(seq 1 60); do
    code=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$URL/api/v1/auth/login/"       -H 'Content-Type: application/json' -d '{}' || true)
    [[ "$code" =~ ^4 ]] && { print_access; return; }
    sleep 2
  done
  echo "The cluster started but $URL does not answer yet; check: $0 status" >&2
  exit 1
}

case "${1:-up}" in
  up) ensure_inotify; create_cluster; build_and_import; deploy; print_access ;;
  start) start ;;
  deploy) ensure_inotify; build_and_import; deploy ;;
  e2e)
    log "Running end-to-end tests against $URL"
    (cd "$ROOT/be" && E2E_BASE_URL="$URL" GOTOOLCHAIN=auto go test -tags e2e -count=1 -v ./internal/e2e/)
    ;;
  status) "${KUBECTL[@]}" get pods,svc,ingress,pvc ;;
  down) k3d cluster delete "$CLUSTER" && rm -f "$KUBECONFIG" ;;
  lan) lan "${2:-}" ;;
  *) echo "usage: $0 {up|start|deploy|e2e|status|down|lan}" >&2; exit 2 ;;
esac
