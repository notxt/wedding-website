package session_test

import (
	"strings"
	"testing"

	"github.com/notxt/wedding-website/internal/session"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	s := session.New("test-secret")
	v := s.Sign("ok")
	if !strings.HasPrefix(v, "ok.") {
		t.Fatalf("expected marker prefix, got %q", v)
	}
	if !s.Verify(v) {
		t.Errorf("Verify rejected its own signature")
	}
}

func TestVerifyRejectsTampering(t *testing.T) {
	s := session.New("test-secret")
	v := s.Sign("ok")
	cases := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"no separator", "okabcdef"},
		{"flipped marker", "no" + v[len("ok"):]},
		{"truncated signature", v[:len(v)-1]},
		{"junk signature", "ok.aaaaaaaa"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if s.Verify(c.value) {
				t.Errorf("Verify accepted %q", c.value)
			}
		})
	}
}

func TestVerifyRejectsDifferentSecret(t *testing.T) {
	a := session.New("secret-a")
	b := session.New("secret-b")
	if b.Verify(a.Sign("ok")) {
		t.Error("Verify accepted signature from a different secret")
	}
}
