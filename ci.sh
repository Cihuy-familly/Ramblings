#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# ci.sh — CI/CD script for blog platform
#
# Usage:
#   ./ci.sh test              # Run tests (vet + build)
#   ./ci.sh docker-build      # Build Docker images
#   ./ci.sh docker-push       # Push Docker images to registry
#   ./ci.sh docker-build-push # Build & push in one step
#
# Environment:
#   REGISTRY_URL       — Docker registry host (required for push)
#   REGISTRY_USERNAME  — registry username (optional, for docker login)
#   REGISTRY_PASSWORD  — registry password (optional, for docker login)
#   CI                 — set to "true" when running in CI (optional)
# ============================================================

MODE="${1:-test}"
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"

# Colors for output (skip if not a terminal)
if [ -t 1 ]; then
  GREEN='\033[0;32m'
  BLUE='\033[0;34m'
  YELLOW='\033[1;33m'
  RED='\033[0;31m'
  NC='\033[0m' # No Color
else
  GREEN=''; BLUE=''; YELLOW=''; RED=''; NC=''
fi

info()  { echo -e "${BLUE}[ci.sh]${NC} $*"; }
ok()    { echo -e "${GREEN}[✓]${NC} $*"; }
warn()  { echo -e "${YELLOW}[!]${NC} $*"; }
fail()  { echo -e "${RED}[✗]${NC} $*"; exit 1; }

# -----------------------------------------------------------
# Mode: test — run backend vet + build, frontend build
# -----------------------------------------------------------
test_all() {
  info "=== Test mode ==="

  # --- Backend ---
  info "Backend: running go vet..."
  (cd "$BACKEND_DIR" && go vet ./...) || fail "go vet failed"
  ok "go vet passed"

  info "Backend: running go build..."
  (cd "$BACKEND_DIR" && go build ./...) || fail "go build failed"
  ok "go build passed"

  # --- Frontend ---
  if [ -d "$FRONTEND_DIR/node_modules" ]; then
    info "Frontend: node_modules found, skipping npm ci"
  else
    info "Frontend: installing dependencies..."
    (cd "$FRONTEND_DIR" && npm ci) || fail "npm ci failed"
  fi
  ok "frontend dependencies installed"

  info "Frontend: building..."
  (cd "$FRONTEND_DIR" && npm run build) || fail "frontend build failed"
  ok "frontend build passed"

  echo ""
  ok "All tests passed!"
}

# -----------------------------------------------------------
# Mode: docker-build — build Docker images
# -----------------------------------------------------------
docker_build() {
  info "=== Docker build mode ==="

  SHORT_SHA="${GITHUB_SHA:-$(git rev-parse --short HEAD 2>/dev/null || echo "dev")}"

  info "Building backend image..."
  docker build \
    -t "blog-backend:latest" \
    -t "blog-backend:$SHORT_SHA" \
    -f "$BACKEND_DIR/Dockerfile" \
    "$BACKEND_DIR" || fail "backend build failed"
  ok "Backend image built"

  info "Building frontend image..."
  docker build \
    -t "blog-frontend:latest" \
    -t "blog-frontend:$SHORT_SHA" \
    -f "$FRONTEND_DIR/Dockerfile" \
    "$FRONTEND_DIR" || fail "frontend build failed"
  ok "Frontend image built"

  # Tag with registry prefix if REGISTRY_URL is set
  if [ -n "${REGISTRY_URL:-}" ]; then
    info "Tagging with registry: $REGISTRY_URL"
    docker tag "blog-backend:latest"  "$REGISTRY_URL/blog-backend:latest"
    docker tag "blog-backend:$SHORT_SHA" "$REGISTRY_URL/blog-backend:$SHORT_SHA"
    docker tag "blog-frontend:latest"  "$REGISTRY_URL/blog-frontend:latest"
    docker tag "blog-frontend:$SHORT_SHA" "$REGISTRY_URL/blog-frontend:$SHORT_SHA"
    ok "Images tagged for registry"
  fi

  echo ""
  ok "Docker images built successfully!"
}

# -----------------------------------------------------------
# Mode: docker-push — push images to registry
# -----------------------------------------------------------
docker_push() {
  info "=== Docker push mode ==="

  if [ -z "${REGISTRY_URL:-}" ]; then
    fail "REGISTRY_URL is not set. Cannot push."
  fi

  # Login if credentials provided
  if [ -n "${REGISTRY_USERNAME:-}" ] && [ -n "${REGISTRY_PASSWORD:-}" ]; then
    info "Logging in to $REGISTRY_URL..."
    echo "$REGISTRY_PASSWORD" | docker login "$REGISTRY_URL" \
      -u "$REGISTRY_USERNAME" --password-stdin || fail "docker login failed"
    ok "Logged in to registry"
  fi

  info "Pushing backend image..."
  docker push "$REGISTRY_URL/blog-backend:latest" || fail "backend push failed"
  ok "Backend image pushed"

  info "Pushing frontend image..."
  docker push "$REGISTRY_URL/blog-frontend:latest" || fail "frontend push failed"
  ok "Frontend image pushed"

  echo ""
  ok "Docker images pushed successfully!"
}

# -----------------------------------------------------------
# Mode: docker-build-push — build + push
# -----------------------------------------------------------
docker_build_push() {
  docker_build
  docker_push
}

# -----------------------------------------------------------
# Main
# -----------------------------------------------------------
case "$MODE" in
  test)
    test_all
    ;;
  docker-build)
    docker_build
    ;;
  docker-push)
    docker_push
    ;;
  docker-build-push)
    docker_build_push
    ;;
  *)
    echo "Usage: $0 {test|docker-build|docker-push|docker-build-push}"
    echo ""
    echo "  test              Run backend vet + build, frontend build"
    echo "  docker-build      Build Docker images"
    echo "  docker-push       Push Docker images (requires REGISTRY_URL)"
    echo "  docker-build-push Build + push in one step"
    exit 1
    ;;
esac