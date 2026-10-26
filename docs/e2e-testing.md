# Running browser E2E tests

The suite runs Chromium against a production Next.js build, the real Go API, real Goose migrations, and a disposable PostgreSQL 17 database. It does not use developer credentials or the development database.

## Prerequisites

- Node.js 24+, Go 1.26+, Docker with a running daemon and Compose v2, and Make.
- Install frontend dependencies and the matching browser once:

```bash
cd frontend
npm ci
npx playwright install chromium
# Linux: npx playwright install --with-deps chromium
cd ..
```

## Commands

```bash
make e2e
make e2e-headed
make e2e-ui
make e2e E2E_ARGS='e2e/specs/checkout.spec.ts'
make e2e E2E_ARGS='--grep onboarding'
```

The root command owns startup, migrations, fixture loading, builds, readiness checks, and teardown. The npm `test:e2e*` scripts only invoke Playwright and require an active harness; prefer the Make targets. UI mode keeps the stack alive until the UI process exits.

Ports **13000** (frontend), **15433** (database), **18080** (API), and **18081** (local providers) must be free. Run only one harness at a time. The frontend build uses `frontend/.next`, so do not run the development frontend concurrently with the test build.

The first run installs Goose **v3.29.0** into `.e2e/bin` and downloads/builds Go dependencies. It also pulls the PostgreSQL image. Subsequent runs reuse downloaded dependencies but always create a fresh database.

## Isolation and safety

- `compose.e2e.yaml` is independent of development Compose and uses tmpfs, not `portal-postgres-data`.
- Every run gets its own `portal-e2e-*` Compose project. Cleanup removes only that project's resources.
- `backend/cmd/e2e` does not load `.env`, requires a local `portal_e2e` database, and does not start scheduled jobs.
- The fixture reset script requires an active run marker. SQL independently checks the database and database user before truncating tables.
- Playwright runs one worker and resets/reseeds before each test, including retries. Each test signs in again; no stale session-state files are reused.
- Test users use `@e2e.invalid` identities. Active/history membership dates are relative to database time; checkout smoke tests use day passes to avoid seasonal purchase windows.
- Test-only routes and implementations are reachable only on the separate local provider server. Production builds use `cmd/api`, which does not import the E2E package.

## External-provider boundaries

**Identity:** A local authorization-code provider supports discovery, PKCE, token exchange, and user info. The frontend follows the real OAuth redirect/callback flow, and Limen creates real database-backed sessions. The normal auth and role middleware remain enabled. Production still uses the hard-coded Zetrova discovery URL.

**Payments:** The Stripe SDK points to a local HTTP substitute that supports fixture-price checkout creation, lookup, expiration, and a small checkout UI. Clicking Pay delivers a signed Stripe-format event to the real webhook handler. Replaying completion verifies application idempotency. Cancellation in this substitute simulates an expired checkout event; it is not a model of every Stripe-hosted cancellation behavior.

**Email:** Resend HTTP requests are redirected to a local capturing sink (`GET http://localhost:18081/emails` during a run). No mail is delivered externally.

The E2E API process blocks unexpected non-loopback HTTP requests. The suite is credential-free, not fully offline: dependency/browser/image installation still requires network access. The frontend may load public static resources.

Local substitutes validate application integration, not compatibility with the actual Zetrova UI or Stripe hosted checkout. Provider SDK contract tests live beside the substitutes in `backend/internal/e2etest`. A real-provider sandbox contract suite is a separate future addition.

## Current coverage

- Public landing/navigation and database-backed pricing.
- Anonymous API denial and protected-page login.
- OAuth callback, session persistence, and member/admin authorization.
- Onboarding validation and persisted student profile.
- Empty, active, and historical membership views.
- Admin user search/update, offline membership assignment, and CSV downloads.
- Checkout success, duplicate signed events, cancellation/expiration, owned-tier unavailability, and invalid webhook signatures.

Expansion items from the plan remain: broader upgrade/group eligibility cases, audit-log assertions, invitation and executive-profile journeys, analytics, responsive coverage, error-state UI tests, and real-provider sandbox checks. Add these incrementally rather than masking unsupported behavior in the provider substitutes.

## Debugging

- Backend, frontend, build, and PostgreSQL logs: `.e2e/logs/`.
- Browser report: `cd frontend && npx playwright show-report`.
- Failed-test traces/screenshots: `frontend/test-results/`.
- Inspect a trace: `cd frontend && npx playwright show-trace test-results/<test>/trace.zip`.
- A missing Docker daemon fails before starting services. A failed startup or build retains logs and tears down the stack.

CI uploads reports, failed-test artifacts, and service logs for seven days. Traces can contain synthetic session cookies; do not publish them as public assets. Stored authentication state is not uploaded.

## Adding tests

1. Put specs in `frontend/e2e/specs` and import `test`, `expect`, and `login` from `../fixtures`.
2. Add only the data needed to `backend/internal/e2etest/seed.sql`. Keep identities synthetic and avoid fixed future expiration dates.
3. Prefer role/label locators and Playwright's retrying assertions. Do not use arbitrary sleeps.
4. Assert persistence after reload or with a read API, not just a success toast.
5. Do not intercept the portal API in full-stack specs. Label any future browser-only error simulations separately.
6. Do not increase workers until database/identity isolation is made per-worker.

The existing CI workflow includes `E2E (Chromium)`, so automatic CD waits for its result. A maintainer must add that check to branch protection after the suite has been run successfully and shown stable in CI; repository code cannot configure branch protection by itself.
