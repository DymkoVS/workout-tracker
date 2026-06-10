package db

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// migrateLockID is an arbitrary constant for pg_advisory_lock so that two
// server instances starting at once can't apply migrations concurrently.
const migrateLockID = 91234567

// Migrate applies any *.sql files in fsys that have not yet been recorded in the
// schema_migrations table, in filename order, each inside its own transaction.
//
// Baseline behaviour: if schema_migrations is empty on first run but the schema
// already exists (the users table is present — e.g. a prod DB built earlier via
// docker initdb or applied by hand), every currently-bundled migration is
// recorded as applied WITHOUT being executed. This makes it safe to introduce
// the runner against an existing database.
//
// Caveat: do not add a brand-new migration in the same deploy that first
// introduces the runner — on that deploy it would be baselined (recorded) rather
// than executed.
func Migrate(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS) error {
	if _, err := pool.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrateLockID); err != nil {
		return fmt.Errorf("advisory lock: %w", err)
	}
	defer pool.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, migrateLockID)

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename   TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	files, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	applied := map[string]bool{}
	rows, err := pool.Query(ctx, `SELECT filename FROM schema_migrations`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var f string
		if err := rows.Scan(&f); err != nil {
			rows.Close()
			return err
		}
		applied[f] = true
	}
	rows.Close()

	// Baseline a pre-existing schema on the very first run.
	if len(applied) == 0 {
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT to_regclass('public.users') IS NOT NULL`).Scan(&exists); err != nil {
			return err
		}
		if exists {
			for _, f := range files {
				if _, err := pool.Exec(ctx,
					`INSERT INTO schema_migrations(filename) VALUES($1) ON CONFLICT DO NOTHING`, f); err != nil {
					return err
				}
			}
			log.Printf("migrate: baselined %d existing migration(s) on pre-existing schema", len(files))
			return nil
		}
	}

	count := 0
	for _, f := range files {
		if applied[f] {
			continue
		}
		sqlText, err := fs.ReadFile(fsys, f)
		if err != nil {
			return err
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(sqlText)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply %s: %w", f, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations(filename) VALUES($1)`, f); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record %s: %w", f, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", f, err)
		}
		log.Printf("migrate: applied %s", f)
		count++
	}
	if count == 0 {
		log.Printf("migrate: schema up to date (%d migration(s))", len(files))
	}
	return nil
}
