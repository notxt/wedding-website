package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func mustReq(remoteAddr, xff string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/rsvp", nil)
	r.RemoteAddr = remoteAddr
	if xff != "" {
		r.Header.Set("X-Forwarded-For", xff)
	}
	return r
}

func TestValidateRSVP(t *testing.T) {
	t.Run("attending with meal and allergies", func(t *testing.T) {
		form := url.Values{
			"attending":       {"yes"},
			"party_names":     {"  Alice Smith, Bob Smith  "},
			"meal_choice":     {"chicken"},
			"dairy_allergy":   {"yes"},
			"gluten_allergy":  {"no"},
			"other_allergies": {"  shellfish  "},
		}
		r, errs := validateRSVP(form)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if !r.Attending {
			t.Errorf("Attending: got false, want true")
		}
		if r.PartyNames != "Alice Smith, Bob Smith" {
			t.Errorf("PartyNames: got %q, want trimmed value", r.PartyNames)
		}
		if r.MealChoice == nil || *r.MealChoice != "chicken" {
			t.Errorf("MealChoice: got %v, want chicken", r.MealChoice)
		}
		if !r.DairyAllergy {
			t.Errorf("DairyAllergy: got false, want true")
		}
		if r.GlutenAllergy {
			t.Errorf("GlutenAllergy: got true, want false")
		}
		if r.OtherAllergies == nil || *r.OtherAllergies != "shellfish" {
			t.Errorf("OtherAllergies: got %v, want shellfish trimmed", r.OtherAllergies)
		}
	})

	t.Run("not attending — meal/allergy fields ignored", func(t *testing.T) {
		form := url.Values{
			"attending":       {"no"},
			"party_names":     {"Carol"},
			"meal_choice":     {"chicken"},
			"dairy_allergy":   {"yes"},
			"gluten_allergy":  {"yes"},
			"other_allergies": {"peanuts"},
		}
		r, errs := validateRSVP(form)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if r.Attending {
			t.Errorf("Attending: got true, want false")
		}
		if r.MealChoice != nil {
			t.Errorf("MealChoice: got %v, want nil", *r.MealChoice)
		}
		if r.DairyAllergy {
			t.Errorf("DairyAllergy: got true, want false (ignored when not attending)")
		}
		if r.GlutenAllergy {
			t.Errorf("GlutenAllergy: got true, want false (ignored when not attending)")
		}
		if r.OtherAllergies != nil {
			t.Errorf("OtherAllergies: got %v, want nil (ignored when not attending)", *r.OtherAllergies)
		}
	})

	t.Run("missing attending", func(t *testing.T) {
		form := url.Values{"party_names": {"Dave"}}
		_, errs := validateRSVP(form)
		if _, ok := errs["attending"]; !ok {
			t.Errorf("expected error on attending, got %v", errs)
		}
	})

	t.Run("blank party_names", func(t *testing.T) {
		form := url.Values{
			"attending":   {"no"},
			"party_names": {"   "},
		}
		_, errs := validateRSVP(form)
		if _, ok := errs["party_names"]; !ok {
			t.Errorf("expected error on party_names, got %v", errs)
		}
	})

	t.Run("attending requires meal", func(t *testing.T) {
		form := url.Values{
			"attending":   {"yes"},
			"party_names": {"Eve"},
		}
		_, errs := validateRSVP(form)
		if _, ok := errs["meal_choice"]; !ok {
			t.Errorf("expected error on meal_choice when attending, got %v", errs)
		}
	})

	t.Run("invalid meal value rejected", func(t *testing.T) {
		form := url.Values{
			"attending":   {"yes"},
			"party_names": {"Frank"},
			"meal_choice": {"steak"},
		}
		_, errs := validateRSVP(form)
		if _, ok := errs["meal_choice"]; !ok {
			t.Errorf("expected error on meal_choice for invalid value, got %v", errs)
		}
	})

	t.Run("attending with empty other_allergies leaves nil", func(t *testing.T) {
		form := url.Values{
			"attending":       {"yes"},
			"party_names":     {"Grace"},
			"meal_choice":     {"vegetarian"},
			"other_allergies": {"   "},
		}
		r, errs := validateRSVP(form)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if r.OtherAllergies != nil {
			t.Errorf("OtherAllergies: got %v, want nil for whitespace-only", *r.OtherAllergies)
		}
	})
}

func TestClientIP(t *testing.T) {
	t.Run("falls back to RemoteAddr", func(t *testing.T) {
		r := mustReq("10.0.0.1:1234", "")
		if got := clientIP(r); got != "10.0.0.1:1234" {
			t.Errorf("got %q, want RemoteAddr", got)
		}
	})

	t.Run("uses XFF first hop when set", func(t *testing.T) {
		r := mustReq("10.0.0.1:1234", "203.0.113.5, 198.51.100.1")
		if got := clientIP(r); got != "203.0.113.5" {
			t.Errorf("got %q, want first XFF hop", got)
		}
	})

	t.Run("XFF without comma", func(t *testing.T) {
		r := mustReq("10.0.0.1:1234", "203.0.113.5")
		if got := clientIP(r); got != "203.0.113.5" {
			t.Errorf("got %q, want sole XFF value", got)
		}
	})
}
