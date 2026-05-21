# Restore Drill

Backup verification as a service. We periodically restore your database
dumps in an isolated sandbox, run assertions, and produce auditor-grade
evidence that your backups are actually restorable.

This repo contains the application (Go monolith) at `app.restoredrill.io`.
The marketing site lives in a separate repo.

## Status

Phase 4 — production perimeter + webhooks. CSRF, rate limiting, and login
brute-force throttling harden the app for public traffic; HMAC-signed
webhooks are the first customer-facing integration.

Implemented:
- Chi + Templ + HTMX + Tailwind monolith
- Postgres sessions (Argon2id), audit log, security headers, signup/login
- River-backed drill orchestrator: `provision → fetch → restore → assert → report → teardown`
- LocalRunner sandbox: temp Postgres database per drill on the host cluster
- FlyMachineRunner stub for the production sandbox driver
- `row_count` assertion
- Unsigned PDF reports via `github.com/go-pdf/fpdf`
- Idempotency on `POST /drills` (per-account, per-key)
- Multi-tenant accounts + memberships; signup auto-creates a personal account
- RBAC (`owner`/`admin`/`member`/`viewer`) via a single `Authorize` matrix
- Email invitations (dev: link logged to stdout), account switcher
- Stripe billing skeleton — degrades to a no-op without `STRIPE_SECRET_KEY`
- CSRF double-submit-cookie protection on every unsafe verb
- In-process token-bucket rate limiting (per-IP on auth, per-account elsewhere)
- Login brute-force throttle (lockout after repeated failures)
- HMAC-SHA256-signed webhooks with River-backed retry, delivery log, replay

## Local development

```sh
make dev
```

This starts Postgres in Docker, runs migrations (goose + River), fetches HTMX,
builds CSS, regenerates Templ files, ensures `tmp/evidence` exists, and
runs the server on `http://localhost:8080`.

To exercise a drill end-to-end:

1. Sign up at `/signup`.
2. From the dashboard, click **Connect a database**.
3. Use `testdata/fixtures/tiny.dump` as the source path, `events` as the
   assertion table, `1` as the minimum row count.
4. Go to `/drills`, pick the target, click **Run drill**, watch the steps
   tick through (HTMX polls every 2 s until terminal).
5. Download the PDF.

## Tests

```sh
DATABASE_URL=postgres://restoredrill:restoredrill@localhost:5432/restoredrill?sslmode=disable \
  go test ./...
```

The drill integration test in `internal/drill/drill_integration_test.go`
needs `DATABASE_URL` to be set and `pg_restore` on `PATH`; otherwise it
skips.

## Layout

```
cmd/server               HTTP + River worker entrypoint
cmd/migrate              goose + River migration CLI
internal/auth            sessions, password hashing, RBAC, login throttle
internal/account         accounts, memberships, invitations
internal/billing         Stripe customer wrapper (+ noop fallback)
internal/ratelimit       token-bucket limiter + middleware
internal/webhooks        signed webhook endpoints, delivery worker, dispatch
internal/db              pgx pool, transaction helpers
internal/drill           drill domain (targets, drills, steps, results)
internal/drill/steps     River workers for each pipeline step
internal/runner          Runner interface + LocalRunner + FlyMachineRunner stub
internal/assertions      assertion kinds (Phase 2: row_count)
internal/report          PDF rendering
internal/web             handlers + Templ templates
internal/web/csrf        CSRF double-submit middleware
migrations               goose SQL migrations
runbooks                 operational runbooks
testdata/fixtures        seeded pg_dump used by local dev + CI
assets                   Tailwind input, static files (HTMX, app.css)
```

See [`docs/plan.md`](docs/plan.md) for the full plan against the
11-layer rubric, including the active phase brief.
