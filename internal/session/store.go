package session

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	cookieName = "session_id"
	sessionTTL = 7 * 24 * time.Hour
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	id := uuid.New()
	expiresAt := time.Now().Add(sessionTTL)
	_, err := s.db.Exec(ctx,
		`INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)`,
		id, userID, expiresAt,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create session: %w", err)
	}
	return id, nil
}

func (s *Store) GetUserID(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	var userID uuid.UUID
	err := s.db.QueryRow(ctx,
		`SELECT user_id FROM sessions WHERE id = $1 AND expires_at > NOW()`,
		sessionID,
	).Scan(&userID)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func (s *Store) Delete(ctx context.Context, sessionID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, sessionID)
	return err
}

func (s *Store) DeleteExpired(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `DELETE FROM sessions WHERE expires_at < NOW()`)
	return err
}

func SetCookie(w http.ResponseWriter, sessionID uuid.UUID) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    sessionID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
}

func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func ReadCookie(r *http.Request) (uuid.UUID, error) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(c.Value)
}
