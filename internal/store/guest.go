package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Guest is one entry in the invite list. The email is the RSVP gate key and is
// stored normalized (lower-cased, trimmed); callers must normalize before lookup.
type Guest struct {
	ID             int64
	Email          string
	FirstName      string
	LastName       string
	PlusOneAllowed bool
	PlusOneName    *string
}

const guestColumns = `id, email, first_name, last_name, plus_one_allowed, plus_one_name`

// GetGuestByEmail looks up a guest by exact (already-normalized) email. The bool
// is false when no guest matches — i.e. the email is not on the list.
func (s *Store) GetGuestByEmail(ctx context.Context, email string) (Guest, bool, error) {
	const q = `SELECT ` + guestColumns + ` FROM guests WHERE email = $1`
	return scanGuest(s.Pool.QueryRow(ctx, q, email), "get guest by email")
}

// GetGuestByID looks up a guest by primary key (the value carried in the signed cookie).
func (s *Store) GetGuestByID(ctx context.Context, id int64) (Guest, bool, error) {
	const q = `SELECT ` + guestColumns + ` FROM guests WHERE id = $1`
	return scanGuest(s.Pool.QueryRow(ctx, q, id), "get guest by id")
}

func scanGuest(row pgx.Row, op string) (Guest, bool, error) {
	var g Guest
	err := row.Scan(&g.ID, &g.Email, &g.FirstName, &g.LastName, &g.PlusOneAllowed, &g.PlusOneName)
	if errors.Is(err, pgx.ErrNoRows) {
		return Guest{}, false, nil
	}
	if err != nil {
		return Guest{}, false, fmt.Errorf("%s: %w", op, err)
	}
	return g, true, nil
}
