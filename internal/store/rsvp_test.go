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
	if _, err := s.Pool.Exec(ctx, "TRUNCATE guests, rsvps RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return s
}

func strp(s string) *string { return &s }

func insertGuest(t *testing.T, ctx context.Context, s *store.Store, email, first, last string, plusOne bool) int64 {
	t.Helper()
	var id int64
	if err := s.Pool.QueryRow(ctx,
		`INSERT INTO guests (email, first_name, last_name, plus_one_allowed) VALUES ($1, $2, $3, $4) RETURNING id`,
		email, first, last, plusOne,
	).Scan(&id); err != nil {
		t.Fatalf("insert guest %s: %v", email, err)
	}
	return id
}

func TestRSVPRoundTrip(t *testing.T) {
	s := openTestStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cases := []struct {
		guest store.Guest
		in    store.RSVP
	}{
		{
			guest: store.Guest{Email: "alice@example.com", FirstName: "Alice", LastName: "Smith", PlusOneAllowed: true},
			in: store.RSVP{
				Attending:         true,
				PartyNames:        "Alice Smith + Bob",
				MealChoice:        strp("chicken"),
				DairyAllergy:      true,
				OtherAllergies:    strp("shellfish"),
				PlusOneAttending:  true,
				PlusOneName:       strp("Bob"),
				PlusOneMealChoice: strp("vegetarian"),
				RemoteAddr:        strp("10.0.0.1:1234"),
			},
		},
		{
			guest: store.Guest{Email: "carol@example.com", FirstName: "Carol", LastName: "Jones"},
			in:    store.RSVP{Attending: false, PartyNames: "Carol Jones"},
		},
		{
			guest: store.Guest{Email: "dave@example.com", FirstName: "Dave", LastName: "Lee"},
			in:    store.RSVP{Attending: true, PartyNames: "Dave Lee", MealChoice: strp("vegetarian"), GlutenAllergy: true},
		},
	}

	for i := range cases {
		gid := insertGuest(t, ctx, s, cases[i].guest.Email, cases[i].guest.FirstName, cases[i].guest.LastName, cases[i].guest.PlusOneAllowed)
		cases[i].in.GuestID = gid
		if _, err := s.UpsertRSVP(ctx, cases[i].in); err != nil {
			t.Fatalf("%s: upsert: %v", cases[i].guest.Email, err)
		}
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
			t.Errorf("%s: row not found", c.in.PartyNames)
			continue
		}
		if got.Attending != c.in.Attending {
			t.Errorf("%s: Attending got %v want %v", c.in.PartyNames, got.Attending, c.in.Attending)
		}
		if !ptrEqual(got.MealChoice, c.in.MealChoice) {
			t.Errorf("%s: MealChoice got %v want %v", c.in.PartyNames, deref(got.MealChoice), deref(c.in.MealChoice))
		}
		if got.PlusOneAttending != c.in.PlusOneAttending {
			t.Errorf("%s: PlusOneAttending got %v want %v", c.in.PartyNames, got.PlusOneAttending, c.in.PlusOneAttending)
		}
		if !ptrEqual(got.PlusOneName, c.in.PlusOneName) {
			t.Errorf("%s: PlusOneName got %v want %v", c.in.PartyNames, deref(got.PlusOneName), deref(c.in.PlusOneName))
		}
		if !ptrEqual(got.PlusOneMealChoice, c.in.PlusOneMealChoice) {
			t.Errorf("%s: PlusOneMealChoice got %v want %v", c.in.PartyNames, deref(got.PlusOneMealChoice), deref(c.in.PlusOneMealChoice))
		}
		if got.SubmittedAt.IsZero() {
			t.Errorf("%s: SubmittedAt should be populated", c.in.PartyNames)
		}
	}

	for i := 1; i < len(list); i++ {
		if list[i-1].SubmittedAt.Before(list[i].SubmittedAt) {
			t.Errorf("list not ordered DESC at index %d", i)
		}
	}
}

func TestRSVPUpsertOverwrites(t *testing.T) {
	s := openTestStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gid := insertGuest(t, ctx, s, "up@example.com", "Up", "Sert", true)

	id1, err := s.UpsertRSVP(ctx, store.RSVP{GuestID: gid, Attending: true, PartyNames: "Up Sert", MealChoice: strp("chicken")})
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	id2, err := s.UpsertRSVP(ctx, store.RSVP{GuestID: gid, Attending: false, PartyNames: "Up Sert"})
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	if id1 != id2 {
		t.Errorf("upsert created a new row instead of updating: %d vs %d", id1, id2)
	}

	got, ok, err := s.GetRSVPByGuestID(ctx, gid)
	if err != nil || !ok {
		t.Fatalf("get by guest: ok=%v err=%v", ok, err)
	}
	if got.Attending {
		t.Errorf("expected the second (not attending) answer to win")
	}

	var n int
	if err := s.Pool.QueryRow(ctx, "SELECT count(*) FROM rsvps WHERE guest_id = $1", gid).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 1 {
		t.Errorf("expected exactly 1 row for the guest, got %d", n)
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
