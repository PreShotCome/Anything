package compliance

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/preshotcome/anything/internal/auth"
	"github.com/preshotcome/anything/internal/evidence"
)

// Retention windows. Evidence and audit logs are auditor-grade and kept for
// seven years; login attempts are a short-lived security signal; API
// idempotency records only need to outlive a client's retry window.
const (
	AuditRetention          = 7 * 365 * 24 * time.Hour
	LoginAttemptsRetention  = 30 * 24 * time.Hour
	APIIdempotencyRetention = 24 * time.Hour
)

// Sweeper enforces retention: it purges evidence past retain_until, old
// audit events, and stale login attempts.
type Sweeper struct {
	pool     *pgxpool.Pool
	evidence *evidence.Service
	throttle *auth.LoginThrottle
}

func NewSweeper(pool *pgxpool.Pool, ev *evidence.Service, throttle *auth.LoginThrottle) *Sweeper {
	return &Sweeper{pool: pool, evidence: ev, throttle: throttle}
}

// SweepResult reports what a sweep removed.
type SweepResult struct {
	EvidencePurged     int
	AuditPruned        int64
	LoginsPruned       int64
	IdempotencyPruned  int64
	VerifyTokensPruned int64
	MagicLinksPruned   int64
}

// Sweep runs one retention pass. Each step is independent: a failure in one
// is collected but does not skip the others, so a single broken table can't
// stall all pruning. Any errors are joined and returned together.
func (s *Sweeper) Sweep(ctx context.Context) (SweepResult, error) {
	var res SweepResult
	var errs []error

	n, err := s.evidence.PurgeExpired(ctx)
	res.EvidencePurged = n
	if err != nil {
		errs = append(errs, fmt.Errorf("purge evidence: %w", err))
	}

	if tag, err := s.pool.Exec(ctx, `
		DELETE FROM audit_events WHERE at < $1
	`, time.Now().UTC().Add(-AuditRetention)); err != nil {
		errs = append(errs, fmt.Errorf("prune audit_events: %w", err))
	} else {
		res.AuditPruned = tag.RowsAffected()
	}

	if err := s.throttle.Prune(ctx, LoginAttemptsRetention); err != nil {
		errs = append(errs, fmt.Errorf("prune login_attempts: %w", err))
	} else {
		res.LoginsPruned = -1 // -1 = "ran, count not tracked"
	}

	if tag, err := s.pool.Exec(ctx, `
		DELETE FROM api_idempotency WHERE created_at < $1
	`, time.Now().UTC().Add(-APIIdempotencyRetention)); err != nil {
		errs = append(errs, fmt.Errorf("prune api_idempotency: %w", err))
	} else {
		res.IdempotencyPruned = tag.RowsAffected()
	}

	// Email-verification tokens carry their own expiry; once past it they are
	// dead weight.
	if tag, err := s.pool.Exec(ctx, `
		DELETE FROM email_verification_tokens WHERE expires_at < now()
	`); err != nil {
		errs = append(errs, fmt.Errorf("prune email_verification_tokens: %w", err))
	} else {
		res.VerifyTokensPruned = tag.RowsAffected()
	}

	// Magic-link tokens carry a short expiry; once past it they are dead.
	if tag, err := s.pool.Exec(ctx, `
		DELETE FROM magic_link_tokens WHERE expires_at < now()
	`); err != nil {
		errs = append(errs, fmt.Errorf("prune magic_link_tokens: %w", err))
	} else {
		res.MagicLinksPruned = tag.RowsAffected()
	}

	return res, errors.Join(errs...)
}

// --- River periodic job ---

// RetentionSweepArgs is the periodic retention job.
type RetentionSweepArgs struct{}

func (RetentionSweepArgs) Kind() string { return "compliance.retention_sweep" }

// RetentionWorker runs Sweep on River's periodic schedule.
type RetentionWorker struct {
	river.WorkerDefaults[RetentionSweepArgs]
	Sweeper *Sweeper
	Logger  *slog.Logger
}

func (w *RetentionWorker) Work(ctx context.Context, _ *river.Job[RetentionSweepArgs]) error {
	res, err := w.Sweeper.Sweep(ctx)
	if err != nil {
		return err
	}
	if w.Logger != nil {
		w.Logger.Info("retention sweep complete",
			"evidence_purged", res.EvidencePurged,
			"audit_pruned", res.AuditPruned,
			"idempotency_pruned", res.IdempotencyPruned,
			"verify_tokens_pruned", res.VerifyTokensPruned,
			"magic_links_pruned", res.MagicLinksPruned,
		)
	}
	return nil
}

func (w *RetentionWorker) Timeout(*river.Job[RetentionSweepArgs]) time.Duration {
	return 10 * time.Minute
}

// PeriodicJob returns the River periodic-job definition: a retention sweep
// every 24h. Register it in the River client config.
func PeriodicJob() *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(24*time.Hour),
		func() (river.JobArgs, *river.InsertOpts) {
			return RetentionSweepArgs{}, nil
		},
		&river.PeriodicJobOpts{RunOnStart: false},
	)
}
