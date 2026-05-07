package store_test

import (
	"context"
	"os"
	"testing"
	"time"

	wedding "github.com/notxt/wedding-website"
	"github.com/notxt/wedding-website/internal/store"
)

func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s, err := store.Open(ctx, dbURL)
	if err != nil {
		t.Skipf("postgres unreachable: %v", err)
	}
	t.Cleanup(s.Close)
	if err := s.Migrate(ctx, wedding.Migrations, "migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := s.Pool.Exec(ctx, "TRUNCATE rsvps RESTART IDENTITY"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return s
}

func strp(s string) *string { return &s }

func TestRSVPRoundTrip(t *testing.T) {
	s := openTestStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cases := []struct {
		name string
		in   store.RSVP
	}{
		{
			name: "attending with meal and allergies",
			in: store.RSVP{
				Attending:      true,
				PartyNames:     "Alice Smith, Bob Smith",
				MealChoice:     strp("chicken"),
				DairyAllergy:   true,
				GlutenAllergy:  false,
				OtherAllergies: strp("shellfish"),
				RemoteAddr:     strp("10.0.0.1:1234"),
			},
		},
		{
			name: "not attending — nullable fields stay NULL",
			in: store.RSVP{
				Attending:  false,
				PartyNames: "Carol",
			},
		},
		{
			name: "attending vegetarian no allergies",
			in: store.RSVP{
				Attending:     true,
				PartyNames:    "Dave",
				MealChoice:    strp("vegetarian"),
				GlutenAllergy: true,
			},
		},
	}

	idsByName := map[string]int64{}
	for _, c := range cases {
		id, err := s.InsertRSVP(ctx, c.in)
		if err != nil {
			t.Fatalf("%s: insert: %v", c.name, err)
		}
		if id == 0 {
			t.Errorf("%s: expected nonzero id", c.name)
		}
		idsByName[c.in.PartyNames] = id
	}

	list, err := s.ListRSVPs(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if got, want := len(list), len(cases); got != want {
		t.Fatalf("list length: got %d, want %d", got, want)
	}

	byName := make(map[string]store.RSVP, len(list))
	for _, r := range list {
		byName[r.PartyNames] = r
	}

	for _, c := range cases {
		got, ok := byName[c.in.PartyNames]
		if !ok {
			t.Errorf("%s: row not found in list", c.name)
			continue
		}
		if got.Attending != c.in.Attending {
			t.Errorf("%s: Attending got %v want %v", c.name, got.Attending, c.in.Attending)
		}
		if !ptrEqual(got.MealChoice, c.in.MealChoice) {
			t.Errorf("%s: MealChoice got %v want %v", c.name, deref(got.MealChoice), deref(c.in.MealChoice))
		}
		if got.DairyAllergy != c.in.DairyAllergy {
			t.Errorf("%s: DairyAllergy got %v want %v", c.name, got.DairyAllergy, c.in.DairyAllergy)
		}
		if got.GlutenAllergy != c.in.GlutenAllergy {
			t.Errorf("%s: GlutenAllergy got %v want %v", c.name, got.GlutenAllergy, c.in.GlutenAllergy)
		}
		if !ptrEqual(got.OtherAllergies, c.in.OtherAllergies) {
			t.Errorf("%s: OtherAllergies got %v want %v", c.name, deref(got.OtherAllergies), deref(c.in.OtherAllergies))
		}
		if !ptrEqual(got.RemoteAddr, c.in.RemoteAddr) {
			t.Errorf("%s: RemoteAddr got %v want %v", c.name, deref(got.RemoteAddr), deref(c.in.RemoteAddr))
		}
		if got.SubmittedAt.IsZero() {
			t.Errorf("%s: SubmittedAt should be populated by default", c.name)
		}
		if got.ID != idsByName[c.in.PartyNames] {
			t.Errorf("%s: ID round-trip mismatch: list %d, insert %d", c.name, got.ID, idsByName[c.in.PartyNames])
		}
	}

	// ListRSVPs contract: rows come back ordered by submitted_at DESC.
	for i := 1; i < len(list); i++ {
		if list[i-1].SubmittedAt.Before(list[i].SubmittedAt) {
			t.Errorf("list not ordered DESC at index %d: %v before %v", i, list[i-1].SubmittedAt, list[i].SubmittedAt)
		}
	}
}

func ptrEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func deref(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}
