package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/notxt/wedding-website/internal/store"
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
	single := store.Guest{FirstName: "Alice", LastName: "Smith"}
	withPlus := store.Guest{FirstName: "Becca", LastName: "Boragine", PlusOneAllowed: true}

	t.Run("attending with meal and allergies", func(t *testing.T) {
		form := url.Values{
			"attending":       {"yes"},
			"meal_choice":     {"chicken"},
			"dairy_allergy":   {"yes"},
			"gluten_allergy":  {"no"},
			"other_allergies": {"  shellfish  "},
		}
		r, _, errs := validateRSVP(form, single)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if !r.Attending {
			t.Errorf("Attending: got false, want true")
		}
		if r.PartyNames != "Alice Smith" {
			t.Errorf("PartyNames: got %q, want derived from guest", r.PartyNames)
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

	t.Run("not attending — meal/allergy/plus-one ignored", func(t *testing.T) {
		form := url.Values{
			"attending":          {"no"},
			"meal_choice":        {"chicken"},
			"plus_one_attending": {"yes"},
			"plus_one_name":      {"Spencer"},
		}
		r, _, errs := validateRSVP(form, withPlus)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if r.Attending {
			t.Errorf("Attending: got true, want false")
		}
		if r.MealChoice != nil {
			t.Errorf("MealChoice: got %v, want nil", *r.MealChoice)
		}
		if r.PlusOneAttending {
			t.Errorf("plus-one should be ignored when the primary is not attending")
		}
		if r.PartyNames != "Becca Boragine" {
			t.Errorf("PartyNames: got %q, want %q", r.PartyNames, "Becca Boragine")
		}
	})

	t.Run("missing attending", func(t *testing.T) {
		_, _, errs := validateRSVP(url.Values{}, single)
		if _, ok := errs["attending"]; !ok {
			t.Errorf("expected error on attending, got %v", errs)
		}
	})

	t.Run("attending requires meal", func(t *testing.T) {
		_, _, errs := validateRSVP(url.Values{"attending": {"yes"}}, single)
		if _, ok := errs["meal_choice"]; !ok {
			t.Errorf("expected error on meal_choice when attending, got %v", errs)
		}
	})

	t.Run("invalid meal value rejected", func(t *testing.T) {
		form := url.Values{"attending": {"yes"}, "meal_choice": {"steak"}}
		_, _, errs := validateRSVP(form, single)
		if _, ok := errs["meal_choice"]; !ok {
			t.Errorf("expected error on meal_choice for invalid value, got %v", errs)
		}
	})

	t.Run("attending with empty other_allergies leaves nil", func(t *testing.T) {
		form := url.Values{"attending": {"yes"}, "meal_choice": {"vegetarian"}, "other_allergies": {"   "}}
		r, _, errs := validateRSVP(form, single)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if r.OtherAllergies != nil {
			t.Errorf("OtherAllergies: got %v, want nil for whitespace-only", *r.OtherAllergies)
		}
	})

	t.Run("plus-one attending requires name and meal", func(t *testing.T) {
		form := url.Values{"attending": {"yes"}, "meal_choice": {"chicken"}, "plus_one_attending": {"yes"}}
		_, _, errs := validateRSVP(form, withPlus)
		if _, ok := errs["plus_one_name"]; !ok {
			t.Errorf("expected error on plus_one_name, got %v", errs)
		}
		if _, ok := errs["plus_one_meal_choice"]; !ok {
			t.Errorf("expected error on plus_one_meal_choice, got %v", errs)
		}
	})

	t.Run("plus-one happy path appends to party names", func(t *testing.T) {
		form := url.Values{
			"attending":               {"yes"},
			"meal_choice":             {"chicken"},
			"plus_one_attending":      {"yes"},
			"plus_one_name":           {"  Spencer  "},
			"plus_one_meal_choice":    {"vegetarian"},
			"plus_one_dairy_allergy":  {"yes"},
			"plus_one_gluten_allergy": {"no"},
		}
		r, _, errs := validateRSVP(form, withPlus)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if !r.PlusOneAttending {
			t.Errorf("PlusOneAttending: got false, want true")
		}
		if r.PlusOneName == nil || *r.PlusOneName != "Spencer" {
			t.Errorf("PlusOneName: got %v, want Spencer trimmed", r.PlusOneName)
		}
		if r.PlusOneMealChoice == nil || *r.PlusOneMealChoice != "vegetarian" {
			t.Errorf("PlusOneMealChoice: got %v, want vegetarian", r.PlusOneMealChoice)
		}
		if !r.PlusOneDairyAllergy {
			t.Errorf("PlusOneDairyAllergy: got false, want true")
		}
		if r.PlusOneGlutenAllergy {
			t.Errorf("PlusOneGlutenAllergy: got true, want false")
		}
		if r.PartyNames != "Becca Boragine + Spencer" {
			t.Errorf("PartyNames: got %q, want %q", r.PartyNames, "Becca Boragine + Spencer")
		}
	})

	t.Run("anything_else is captured, trimmed, and nil when blank", func(t *testing.T) {
		form := url.Values{
			"attending":     {"yes"},
			"meal_choice":   {"chicken"},
			"anything_else": {"  we'll arrive late  "},
		}
		r, _, errs := validateRSVP(form, single)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if r.AnythingElse == nil || *r.AnythingElse != "we'll arrive late" {
			t.Errorf("AnythingElse: got %v, want trimmed note", r.AnythingElse)
		}

		blank := url.Values{"attending": {"yes"}, "meal_choice": {"chicken"}, "anything_else": {"   "}}
		r2, _, _ := validateRSVP(blank, single)
		if r2.AnythingElse != nil {
			t.Errorf("AnythingElse: got %v, want nil for whitespace-only", *r2.AnythingElse)
		}
	})

	t.Run("plus-one allergies ignored when not bringing a guest", func(t *testing.T) {
		form := url.Values{
			"attending":               {"yes"},
			"meal_choice":             {"chicken"},
			"plus_one_attending":      {"no"},
			"plus_one_dairy_allergy":  {"yes"},
			"plus_one_gluten_allergy": {"yes"},
		}
		r, _, errs := validateRSVP(form, withPlus)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if r.PlusOneDairyAllergy || r.PlusOneGlutenAllergy {
			t.Errorf("plus-one allergies should be ignored when not bringing a guest")
		}
	})

	t.Run("plus-one ignored when guest not allowed one", func(t *testing.T) {
		form := url.Values{
			"attending":          {"yes"},
			"meal_choice":        {"chicken"},
			"plus_one_attending": {"yes"},
			"plus_one_name":      {"Nope"},
		}
		r, _, errs := validateRSVP(form, single)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if r.PlusOneAttending {
			t.Errorf("plus-one should be ignored for a guest without an allowance")
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
