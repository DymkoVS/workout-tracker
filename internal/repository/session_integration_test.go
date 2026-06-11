package repository

import (
	"context"
	"testing"
	"time"

	"workout-tracker/internal/session"
)

// TestSessionTouch: скользящая сессия — Touch продлевает TTL только если с
// прошлого продления прошло больше суток, и не трогает истёкшие сессии.
func TestSessionTouch(t *testing.T) {
	ctx := context.Background()
	store := session.NewStore(testPool)
	uid := mkUser(t)

	sid, err := store.Create(ctx, uid)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Свежая сессия (expires_at = now+7d) — продлевать рано.
	if store.Touch(ctx, sid) {
		t.Error("Touch on fresh session should be a no-op")
	}

	// Состарим сессию: осталось 5 дней (прошло больше суток с продления).
	if _, err := testPool.Exec(ctx,
		`UPDATE sessions SET expires_at = NOW() + interval '5 days' WHERE id = $1`, sid); err != nil {
		t.Fatalf("age session: %v", err)
	}
	if !store.Touch(ctx, sid) {
		t.Fatal("Touch on aged session should extend it")
	}
	var expires time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT expires_at FROM sessions WHERE id = $1`, sid).Scan(&expires); err != nil {
		t.Fatalf("read expires_at: %v", err)
	}
	if remaining := time.Until(expires); remaining < 6*24*time.Hour {
		t.Errorf("after Touch remaining = %v, want ~7d", remaining)
	}

	// Истёкшая сессия не воскрешается.
	if _, err := testPool.Exec(ctx,
		`UPDATE sessions SET expires_at = NOW() - interval '1 hour' WHERE id = $1`, sid); err != nil {
		t.Fatalf("expire session: %v", err)
	}
	if store.Touch(ctx, sid) {
		t.Error("Touch must not resurrect an expired session")
	}
	if _, err := store.GetUserID(ctx, sid); err == nil {
		t.Error("GetUserID should fail for expired session")
	}
}
