package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type RSVP struct {
	ID                   int64
	SubmittedAt          time.Time
	GuestID              int64
	Attending            bool
	PartyNames           string
	MealChoice           *string
	DairyAllergy         bool
	GlutenAllergy        bool
	OtherAllergies       *string
	PlusOneAttending     bool
	PlusOneName          *string
	PlusOneMealChoice    *string
	PlusOneDairyAllergy  bool
	PlusOneGlutenAllergy bool
	AnythingElse         *string
	RemoteAddr           *string
}

// InsertRSVP records a guest's RSVP. RSVPs are submit-once: if the guest already
// has one (enforced by the rsvps_guest_id_key unique index), the insert is a no-op
// and inserted is false. Corrections are handled out of band by the couple.
func (s *Store) InsertRSVP(ctx context.Context, r RSVP) (id int64, inserted bool, err error) {
	const q = `
		INSERT INTO rsvps
			(guest_id, attending, party_names, meal_choice, dairy_allergy, gluten_allergy,
			 other_allergies, plus_one_attending, plus_one_name, plus_one_meal_choice,
			 plus_one_dairy_allergy, plus_one_gluten_allergy, anything_else, remote_addr)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (guest_id) DO NOTHING
		RETURNING id`
	err = s.Pool.QueryRow(ctx, q,
		r.GuestID, r.Attending, r.PartyNames, r.MealChoice, r.DairyAllergy, r.GlutenAllergy,
		r.OtherAllergies, r.PlusOneAttending, r.PlusOneName, r.PlusOneMealChoice,
		r.PlusOneDairyAllergy, r.PlusOneGlutenAllergy, r.AnythingElse, r.RemoteAddr,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("insert rsvp: %w", err)
	}
	return id, true, nil
}

// GetRSVPByGuestID returns a guest's existing RSVP for pre-filling the form. The
// bool is false when the guest has not RSVP'd yet.
func (s *Store) GetRSVPByGuestID(ctx context.Context, guestID int64) (RSVP, bool, error) {
	const q = `
		SELECT id, submitted_at, guest_id, attending, party_names, meal_choice,
		       dairy_allergy, gluten_allergy, other_allergies,
		       plus_one_attending, plus_one_name, plus_one_meal_choice,
		       plus_one_dairy_allergy, plus_one_gluten_allergy, anything_else, remote_addr
		FROM rsvps
		WHERE guest_id = $1`
	var r RSVP
	err := s.Pool.QueryRow(ctx, q, guestID).Scan(
		&r.ID, &r.SubmittedAt, &r.GuestID, &r.Attending, &r.PartyNames, &r.MealChoice,
		&r.DairyAllergy, &r.GlutenAllergy, &r.OtherAllergies,
		&r.PlusOneAttending, &r.PlusOneName, &r.PlusOneMealChoice,
		&r.PlusOneDairyAllergy, &r.PlusOneGlutenAllergy, &r.AnythingElse, &r.RemoteAddr,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return RSVP{}, false, nil
	}
	if err != nil {
		return RSVP{}, false, fmt.Errorf("get rsvp by guest: %w", err)
	}
	return r, true, nil
}

func (s *Store) ListRSVPs(ctx context.Context) ([]RSVP, error) {
	const q = `
		SELECT id, submitted_at, attending, party_names, meal_choice,
		       dairy_allergy, gluten_allergy, other_allergies,
		       plus_one_attending, plus_one_name, plus_one_meal_choice,
		       plus_one_dairy_allergy, plus_one_gluten_allergy, anything_else, remote_addr
		FROM rsvps
		ORDER BY submitted_at DESC`
	rows, err := s.Pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list rsvps: %w", err)
	}
	defer rows.Close()
	var out []RSVP
	for rows.Next() {
		var r RSVP
		if err := rows.Scan(
			&r.ID, &r.SubmittedAt, &r.Attending, &r.PartyNames, &r.MealChoice,
			&r.DairyAllergy, &r.GlutenAllergy, &r.OtherAllergies,
			&r.PlusOneAttending, &r.PlusOneName, &r.PlusOneMealChoice,
			&r.PlusOneDairyAllergy, &r.PlusOneGlutenAllergy, &r.AnythingElse, &r.RemoteAddr,
		); err != nil {
			return nil, fmt.Errorf("scan rsvp: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rsvps: %w", err)
	}
	return out, nil
}
