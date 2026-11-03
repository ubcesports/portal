#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# Reset is available only to the child test process of an active harness run.
if [[ "${1:-}" == "--reset" ]]; then
  case "${E2E_PROJECT:-}" in portal-e2e-*) ;; *) echo 'Run tests through make e2e' >&2; exit 1 ;; esac
  [[ -f "$ROOT/.e2e/$E2E_PROJECT.running" ]] || { echo 'No active E2E run' >&2; exit 1; }
  docker compose -p "$E2E_PROJECT" -f compose.e2e.yaml exec -T db \
    psql -X -v ON_ERROR_STOP=1 -U portal_e2e -d portal_e2e < backend/internal/e2etest/seed.sql
  exit
fi

for tool in docker node npm go curl; do
  command -v "$tool" >/dev/null || { echo "Missing prerequisite: $tool" >&2; exit 1; }
done
node -e 'const r = require("node:child_process").spawnSync("docker", ["info"], { timeout: 10000, stdio: "ignore" }); process.exit(r.status === 0 ? 0 : 1)' || { echo 'Docker daemon is unavailable. Start Docker and retry make e2e.' >&2; exit 1; }
[[ -d frontend/node_modules/@playwright/test ]] || { echo 'Run cd frontend && npm ci first' >&2; exit 1; }

# Fail before starting anything if another local stack owns a reserved port.
node <<'JS'
const net = require('node:net');
(async () => {
  for (const port of [13000, 15433, 18080, 18081]) {
    await new Promise((resolve, reject) => {
      const s = net.createServer();
      s.once('error', reject);
      s.listen(port, () => s.close(resolve));
    });
  }
})().catch(e => { console.error('E2E ports must be free:', e.message); process.exit(1); });
JS

export E2E_PROJECT="portal-e2e-$$"
mkdir -p .e2e/bin .e2e/logs
api_pid='' frontend_pid=''
cleanup() {
  status=$?
  trap - EXIT INT TERM
  [[ -z "$frontend_pid" ]] || kill "$frontend_pid" 2>/dev/null || true
  [[ -z "$api_pid" ]] || kill "$api_pid" 2>/dev/null || true
  [[ -z "$frontend_pid" ]] || wait "$frontend_pid" 2>/dev/null || true
  [[ -z "$api_pid" ]] || wait "$api_pid" 2>/dev/null || true
  docker compose -p "$E2E_PROJECT" -f compose.e2e.yaml logs --no-color > .e2e/logs/postgres.log 2>&1 || true
  docker compose -p "$E2E_PROJECT" -f compose.e2e.yaml down --volumes --remove-orphans >/dev/null 2>&1 || true
  rm -f ".e2e/$E2E_PROJECT.running"
  if [[ "$status" != 0 ]]; then echo 'E2E failed. See .e2e/logs and frontend/playwright-report.' >&2; fi
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

# Explicit synthetic configuration; cmd/e2e never loads backend/.env.
export POSTGRES_USER=portal_e2e POSTGRES_PASSWORD=portal_e2e POSTGRES_DB=portal_e2e
export POSTGRES_HOST=127.0.0.1 POSTGRES_PORT=15433
export BASE_URL=http://localhost:18080 FRONTEND_URL=http://localhost:13000
export NEXT_PUBLIC_API_URL=http://localhost:18080 NEXT_PUBLIC_SITE_URL=http://localhost:13000
export API_INTERNAL_URL=http://localhost:18080
# Limen requires exactly 32 bytes (synthetic, test-only key).
export LIMEN_SECRET=0123456789abcdef0123456789abcdef
export OAUTH_CLIENT_ID=e2e OAUTH_CLIENT_SECRET=e2e
export OAUTH_CALLBACK_URL=http://localhost:18080/auth/oauth/zetrova/callback
export STRIPE_SECRET_KEY=sk_test_e2e STRIPE_WEBHOOK_SECRET=whsec_e2e
export STRIPE_CHECKOUT_SUCCESS_URL='http://localhost:13000/checkout?success=true'
export STRIPE_CHECKOUT_CANCEL_URL='http://localhost:13000/checkout?canceled=true'
export RESEND_API_KEY=re_e2e SENDER_EMAIL='E2E <portal@e2e.invalid>'
export EXEC_ONBOARDING_CODE=e2e-exec-invite
export NEXT_TELEMETRY_DISABLED=1

# Use a pinned, harness-local migration tool, not whichever goose is on PATH.
if [[ ! -x .e2e/bin/goose ]]; then
  (cd backend && GOBIN="$ROOT/.e2e/bin" go install github.com/pressly/goose/v3/cmd/goose@v3.29.0)
fi
docker compose -p "$E2E_PROJECT" -f compose.e2e.yaml up -d --wait --wait-timeout 60
.e2e/bin/goose -dir backend/sql/migrations postgres \
  'postgres://portal_e2e:portal_e2e@127.0.0.1:15433/portal_e2e?sslmode=disable' up
touch ".e2e/$E2E_PROJECT.running"
bash scripts/e2e.sh --reset
(cd backend && go build -o "$ROOT/.e2e/bin/api" ./cmd/e2e)
.e2e/bin/api > .e2e/logs/api.log 2>&1 &
api_pid=$!
wait_for() {
  local url="$1" pid="$2"
  for ((i=0; i<120; i++)); do
    kill -0 "$pid" 2>/dev/null || { echo "Service exited before ready: $url" >&2; return 1; }
    if curl --silent --fail "$url" >/dev/null; then return; fi
    sleep 1
  done
  echo "Readiness timeout: $url" >&2; return 1
}
wait_for "$BASE_URL/health" "$api_pid"
(cd frontend && npm run build) > .e2e/logs/build.log 2>&1
(cd frontend && exec node node_modules/next/dist/bin/next start --hostname localhost --port 13000) > .e2e/logs/frontend.log 2>&1 &
frontend_pid=$!
wait_for "$FRONTEND_URL" "$frontend_pid"
(cd frontend && npx playwright test "$@")
