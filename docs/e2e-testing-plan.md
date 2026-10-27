# E2E testing implementation plan

## Implementation status

The isolated harness, local provider substitutes, CI job, and initial public/auth/member/admin/checkout specs are implemented. See [the E2E guide](e2e-testing.md) for commands and precise coverage. Browser execution still needs a running Docker daemon; do not treat compilation or test discovery as a successful E2E run. Expansion coverage and branch-protection setup remain outstanding.

## Goal and approach

Add a repeatable browser suite that exercises **Next.js → Go API → PostgreSQL**, runs locally and on pull requests, and needs no production or sandbox credentials.

Use **Playwright Test**, installed in `frontend`, with Chromium initially. Keep existing Go tests for detailed policy, handler, and service edge cases; E2E tests should prove representative user journeys and integration boundaries rather than duplicate every unit test.

Only replace external providers (Zetrova, Stripe, Resend). Do not mock the portal API in the core suite. Browser-only error-state tests may intercept requests, but should be labeled separately from full-stack tests.

## Current state

- `frontend/package.json` has lint, format, build, and start commands, but no test runner.
- `.github/workflows/ci.yaml` runs quality checks, Go tests/builds, and a Next.js build; no browser tests.
- Go tests already cover portions of auth, membership policy, services, handlers, and scheduling.
- `backend/internal/auth/auth.go` configures Limen sessions and hard-codes Zetrova OIDC discovery.
- `backend/cmd/api/main.go` composes payment, email, and scheduler modules on startup.
- `frontend/src/lib/client.ts` handles credentialed requests and redirects for unauthenticated, incomplete-onboarding, and forbidden responses.
- Existing `compose.yaml` uses the persistent `portal-postgres-data` development volume. Tests must not reset or reuse it.
- Existing catalog and mock SQL seeds are not a dedicated deterministic E2E fixture set.

## 1. Build an isolated test environment

### Work

1. Add a dedicated `compose.e2e.yaml` for PostgreSQL 17, with distinct ports/project naming and disposable storage. Do not inherit the development volume or load developer `.env` files.
2. Add an orchestration script that:
   - supplies explicit test-only environment values;
   - starts PostgreSQL and waits for readiness;
   - applies the real Goose migrations;
   - seeds the E2E catalog and users;
   - starts local external-provider substitutes and the Go API;
   - builds and starts Next.js for CI;
   - checks API `/health` and frontend readiness;
   - runs Playwright and preserves its exit status;
   - collects logs and stops/removes only resources created by this run, including on failure.
3. Set `NEXT_PUBLIC_API_URL` and `NEXT_PUBLIC_SITE_URL` before the frontend build, and configure `API_INTERNAL_URL`, API base URL, trusted frontend origin, and OAuth callback consistently. Use one hostname consistently for browser cookies.
4. Disable scheduled email work in the test composition; use an in-memory/capturing mail implementation for request-driven email.
5. Pin tooling versions, including Goose. Do not depend on an unversioned tool install in CI.

### Acceptance

A fresh checkout can boot the stack without developer credentials. Failed runs clean up reliably. Running it cannot modify the local development database.

## 2. Make external dependencies deterministic

### Authentication

- First inspect the installed Limen implementation to confirm discovery behavior, callback paths, session creation, and cookie requirements.
- Make the OAuth discovery URL injectable, preserving the current Zetrova URL as the production default.
- Run a local minimal OIDC provider supporting the discovery, authorization, token, and user-info behavior Limen actually uses. Support signing/JWKS if required by the library.
- Exercise the real login redirect/callback/session-cookie flow in at least one dedicated spec. No third-party login UI or shared credentials.
- Authenticate fixture users through the local provider and save Playwright storage state for reuse where appropriate. Use fresh browser contexts per test and separate users for mutation tests.
- Do not add an unauthenticated login-as-user endpoint to the production router or disable real auth middleware.

### Payments and email

- Introduce narrow provider interfaces at existing service boundaries where needed; wire real implementations in the production entrypoint and deterministic implementations in a separate E2E entrypoint such as `backend/cmd/e2e/main.go`.
- The fake checkout provider should support the create/retrieve/expire/payment-intent operations the app uses and return a local checkout URL.
- Submit signed Stripe-format events to the real `/webhooks/stripe` handler using a test-only webhook secret. Preserve signature validation, transaction correlation, and application persistence logic.
- Capture email in a local sink; never send real mail during tests.
- Keep test controls local to the harness, not accessible through production routes. Verify production builds do not include the E2E entrypoint or fake-provider wiring.

### Acceptance

The suite can sign in, create checkout sessions, process payment events, and trigger email behavior without reaching Zetrova, Stripe, or Resend. A CI network guard or provider request log should detect unexpected external-provider traffic.

## 3. Add deterministic fixtures and isolation

Create a small dedicated fixture builder, preferably in Go so it can reuse database types and application conventions.

Required personas/data:

- Anonymous visitor.
- Authenticated user who has not completed onboarding.
- Onboarded member with no memberships.
- Member with active and expired memberships.
- Student/non-student and relevant group variants for representative eligibility checks.
- Administrator and non-admin member.
- Executive-group member for later exec-profile coverage.
- Small catalog containing eligible, ineligible, and upgradeable tier examples.

Use synthetic identities, stable fixture identifiers, and membership dates relative to a captured run timestamp. Ensure catalog purchase windows and expiration rules remain valid as calendar time advances; avoid relying on existing dated staging seeds. Cover exact date boundaries in Go tests or through a deliberately injected clock, not browser time mocking alone.

Start with **one Playwright worker**. Reset/reseed mutable data before each independent test, including retries, before creating sessions. Do not reuse storage state whose sessions were removed by a reset. A practical first version can recreate sessions after each reset; optimize later with per-worker databases and isolated identities.

Protect every destructive fixture command with an explicit test-database allowlist/check. Never accept an arbitrary production database URL for reset operations.

### Acceptance

Tests pass independently, in a different order, and on repeat runs. Retrying a failed test does not inherit its previous mutations.

## 4. Add the Playwright harness

Suggested layout:

```text
frontend/playwright.config.ts
frontend/e2e/
  fixtures/
  helpers/
  specs/
    public.spec.ts
    authentication.spec.ts
    onboarding.spec.ts
    memberships.spec.ts
    admin.spec.ts
    checkout.spec.ts
backend/cmd/e2e/main.go
backend/internal/e2etest/           # fixtures/provider wiring; not production routes
compose.e2e.yaml
scripts/e2e.sh
```

- Add `@playwright/test` and commit the lockfile.
- Add `test:e2e`, `test:e2e:ui`, and `test:e2e:headed` npm scripts; document that these run against the harness-managed stack.
- Add root `make e2e` and an interactive/debug variant to manage the complete stack.
- Configure base URL, explicit timeouts, CI `forbidOnly`, one worker initially, and at most one CI retry.
- Retain traces and screenshots on failure; publish an HTML report. Ignore reports, results, and authentication state in Git and formatting checks where necessary.
- Prefer accessible role/label locators; add small stable `data-testid` hooks only where semantic locators are insufficient.
- Use Playwright assertions and readiness polling rather than fixed sleeps.
- Keep page helpers small and journey-focused; avoid a large page-object framework initially.
- Read the installed Next.js documentation before implementing any Next-specific changes, as required by `frontend/AGENTS.md`.

## 5. Implement coverage in priority order

| Priority | Journey | Assertions |
| --- | --- | --- |
| P0 | Public landing and pricing | Pages load, navigation works, real database-backed catalog appears, public pages do not force login. |
| P0 | Authentication and authorization | Login callback creates a valid session; protected access without a session redirects to login; non-admin receives forbidden UI and backend denial. |
| P0 | Onboarding | Incomplete user is redirected appropriately; invalid input is rejected; valid submission persists and unlocks member routes after reload. |
| P0 | Member profile | Correct identity, active memberships, expired/history records, and empty states appear for the selected user. |
| P0 | Admin user management | Search finds fixture user; detail view loads; one supported profile update persists after reload. |
| P1 | Admin offline membership | Eligible tier can be assigned; membership appears for the member; related transaction and audit record are persisted. |
| P1 | Checkout success | Real checkout endpoint creates transaction; local provider completes payment; signed webhook activates membership; profile reflects persisted result. |
| P1 | Checkout cancel/failure | Supported cancellation/expiry behavior does not grant membership; UI permits the intended next action. |
| P1 | Webhook safety | Duplicate completion does not duplicate membership/payment effects; invalid signature is rejected. These may be API-level tests in the E2E harness rather than additional browser journeys. |
| P1 | Eligibility | Representative ineligible tier is unavailable/rejected, and one supported upgrade follows expected price/access behavior. |
| P2 | Admin exports and audit logs | Downloaded CSV has expected headers and fixture records; filters and audit results correspond to a known mutation. |
| P2 | Invitations, exec profiles, analytics | Add one happy path per feature after confirming current UI workflows and data contracts. |
| P2 | Responsive and failure UI | Small mobile smoke suite; explicitly labeled request-interception tests for network failures and user-facing errors. |

For mutations, assert the user-visible result after reload and, where useful, verify persisted state through a read API or harness database assertion. Do not treat a toast or checkout return URL as proof of a completed purchase.

## 6. Integrate with CI

Add an `e2e` job to the existing `.github/workflows/ci.yaml`, so its result contributes to the CI workflow that automatic CD already waits for.

Job sequence:

1. Check out the repository; install Go using `backend/go.mod` and Node.js 24.
2. Install npm/Go dependencies and pinned migration tooling.
3. Install the matching Playwright Chromium browser and Linux dependencies.
4. Execute the same full-stack orchestration command used locally, using a production Next.js build.
5. Always upload Playwright reports, failure traces/screenshots, and service logs; always clean up.

Use a job timeout and existing workflow cancellation behavior. Run with synthetic secrets only so fork pull requests do not need repository secrets. Configure branch protection to require the E2E check once stable. Report retry-only passes as flakiness rather than silently accepting ongoing instability.

Avoid coupling this first iteration to deployment or a shared staging database. A later manually triggered Stripe sandbox contract test can validate provider compatibility separately; local fake checkout coverage does not prove Stripe's hosted UI works.

## Delivery sequence

1. **PR 1 — Harness and public smoke:** isolated PostgreSQL, migrations/seeds, credential-free backend startup, Playwright configuration, local command, CI job, public/pricing specs.
2. **PR 2 — Identity and core access:** local OIDC flow, session fixtures/reset strategy, authentication, onboarding, profile, and authorization specs.
3. **PR 3 — Admin workflows:** user search/update, offline membership assignment, persistence and audit assertions.
4. **PR 4 — Payments:** deterministic checkout provider, real signed webhook processing, successful/cancelled checkout and duplicate-event coverage.
5. **PR 5 — Expansion:** invitations, exec profiles, exports, analytics, mobile smoke, and optional real-provider contract tests.

## Definition of done

- One documented root command runs the suite from a clean checkout.
- No real credentials, external-provider availability, or persistent development data are required.
- P0 journeys exercise the real frontend, API, migrations, database, and authorization middleware.
- Each test is independently rerunnable; cleanup works on failure and retry.
- PR failures produce useful traces, screenshots, and backend logs without exposing stored auth state.
- Existing checks remain green, E2E is a required check, and the README explains setup, debugging, fixture additions, and provider coverage limits.
- Target an initial CI runtime under ten minutes; measure it before adding parallel workers or more browser projects.
