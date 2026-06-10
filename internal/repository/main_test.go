package repository

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"workout-tracker/internal/db"
	"workout-tracker/internal/model"
	"workout-tracker/migrations"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testPool is a connection pool to a throwaway test database, set up in TestMain.
// If no database is reachable, the integration tests are skipped (not failed) so
// `go test ./...` still passes in environments without Postgres.
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://workout:workout_secret@localhost:5432/workout_tracker_test?sslmode=disable"
	}
	ctx := context.Background()

	if err := resetTestDB(ctx, dsn); err != nil {
		fmt.Printf("integration tests skipped (no test DB: %v)\n", err)
		os.Exit(0)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Printf("integration tests skipped (connect: %v)\n", err)
		os.Exit(0)
	}
	if err := db.Migrate(ctx, pool, migrations.FS); err != nil {
		fmt.Printf("integration tests: migrate failed: %v\n", err)
		os.Exit(1)
	}
	testPool = pool

	code := m.Run()
	pool.Close()
	os.Exit(code)
}

// resetTestDB drops and recreates the target database from the maintenance
// ("postgres") database, giving each run a clean schema for Migrate to build.
func resetTestDB(ctx context.Context, dsn string) error {
	u, err := url.Parse(dsn)
	if err != nil {
		return err
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		return fmt.Errorf("no database name in TEST_DATABASE_URL")
	}

	adminURL := *u
	adminURL.Path = "/postgres"
	admin, err := pgxpool.New(ctx, adminURL.String())
	if err != nil {
		return err
	}
	defer admin.Close()
	if err := admin.Ping(ctx); err != nil {
		return err
	}

	if _, err := admin.Exec(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS %q WITH (FORCE)`, dbName)); err != nil {
		return err
	}
	if _, err := admin.Exec(ctx, fmt.Sprintf(`CREATE DATABASE %q`, dbName)); err != nil {
		return err
	}
	return nil
}

// mkUser creates a unique throwaway user and returns its ID.
func mkUser(t *testing.T) uuid.UUID {
	t.Helper()
	users := NewUserRepository(testPool)
	u, err := users.Create(context.Background(), model.CreateUserInput{
		Login:    "test_" + uuid.NewString()[:8],
		Password: "pw_12345",
		FullName: "Test User",
		Role:     "client",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u.ID
}
