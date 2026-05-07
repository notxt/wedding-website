package store

import (
	"context"
	"fmt"
	"time"
)

type RSVP struct {
	ID             int64
	SubmittedAt    time.Time
	Attending      bool
	PartyNames     string
	MealChoice     *string
	DairyAllergy   bool
	GlutenAllergy  bool
	OtherAllergies *string
	RemoteAddr     *string
}

func (s *Store) InsertRSVP(ctx context.Context, r RSVP) (int64, error) {
	const q = `
		INSERT INTO rsvps
			(attending, party_names, meal_choice, dairy_allergy, gluten_allergy, other_allergies, remote_addr)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	var id int64
	if err := s.Pool.QueryRow(ctx, q,
		r.Attending, r.PartyNames, r.MealChoice,
		r.DairyAllergy, r.GlutenAllergy, r.OtherAllergies,
		r.RemoteAddr,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("insert rsvp: %w", err)
	}
	return id, nil
}

func (s *Store) ListRSVPs(ctx context.Context) ([]RSVP, error) {
	const q = `
		SELECT id, submitted_at, attending, party_names, meal_choice,
		       dairy_allergy, gluten_allergy, other_allergies, remote_addr
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
			&r.DairyAllergy, &r.GlutenAllergy, &r.OtherAllergies, &r.RemoteAddr,
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
