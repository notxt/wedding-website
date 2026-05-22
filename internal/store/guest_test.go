package store_test

import (
	"context"
	"testing"
	"time"
)

func TestGuestLookup(t *testing.T) {
	s := openTestStore(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := insertGuest(t, ctx, s, "becca@example.com", "Becca", "Boragine", true)

	g, found, err := s.GetGuestByEmail(ctx, "becca@example.com")
	if err != nil || !found {
		t.Fatalf("by email: found=%v err=%v", found, err)
	}
	if g.ID != id || g.FirstName != "Becca" || !g.PlusOneAllowed {
		t.Errorf("unexpected guest: %+v", g)
	}

	if _, found, err := s.GetGuestByEmail(ctx, "stranger@example.com"); err != nil || found {
		t.Errorf("expected not found, got found=%v err=%v", found, err)
	}

	g2, found, err := s.GetGuestByID(ctx, id)
	if err != nil || !found || g2.Email != "becca@example.com" {
		t.Errorf("by id: %+v found=%v err=%v", g2, found, err)
	}

	if _, found, err := s.GetGuestByID(ctx, 999999); err != nil || found {
		t.Errorf("expected not found by id, got found=%v err=%v", found, err)
	}
}
