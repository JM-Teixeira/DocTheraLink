#!/usr/bin/env bash
# Local Kubernetes for Klyniva using k3d (k3s in Docker).
#
#   scripts/k8s-local.sh up       create cluster (if needed), build, import, deploy
#   scripts/k8s-local.sh deploy   rebuild images and roll out (cluster must exist)
#   scripts/k8s-local.sh e2e      run the end-to-end API tests against the cluster
#   scripts/k8s-local.sh status   show pods, services and ingress
#   scripts/k8s-local.sh down     delete the cluster (all its data)
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
  log "Ready: $URL  (demo users: doctor / physiotherapist / nurse, password 1234@1)"
}

case "${1:-up}" in
  up) create_cluster; build_and_import; deploy ;;
  deploy) build_and_import; deploy ;;
  e2e)
    log "Running end-to-end tests against $URL"
    (cd "$ROOT/be" && E2E_BASE_URL="$URL" GOTOOLCHAIN=auto go test -tags e2e -count=1 -v ./internal/e2e/)
    ;;
  status) "${KUBECTL[@]}" get pods,svc,ingress,pvc ;;
  down) k3d cluster delete "$CLUSTER" && rm -f "$KUBECONFIG" ;;
  *) echo "usage: $0 {up|deploy|e2e|status|down}" >&2; exit 2 ;;
esac
