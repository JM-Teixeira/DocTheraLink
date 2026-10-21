package patient

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestValidNIF(t *testing.T) {
	valid := []string{"123456789", "501964843", "100000010", "100000002"}
	invalid := []string{"123456780", "12345678", "1234567890", "12345678a", ""}
	for _, n := range valid {
		if !ValidNIF(n) {
			t.Errorf("%s should be valid", n)
		}
	}
	for _, n := range invalid {
		if ValidNIF(n) {
			t.Errorf("%s should be invalid", n)
		}
	}
}

func TestIdentifierNormalisation(t *testing.T) {
	v, err := identifier("nif", strp(" 123 456 789 "))
	if err != nil || v == nil || *v != "123456789" {
		t.Fatalf("nif with spaces: %v %v", v, err)
	}
	if v, err := identifier("sns_number", strp("123.456.789")); err != nil || *v != "123456789" {
		t.Fatalf("sns with dots: %v %v", v, err)
	}
	if v, err := identifier("nif", strp("  ")); err != nil || v != nil {
		t.Errorf("blank clears: %v %v", v, err)
	}
	var verr *ValidationError
	if _, err := identifier("nif", strp("123456780")); !errors.As(err, &verr) || verr.Field != "nif" {
		t.Errorf("bad check digit: %v", err)
	}
	if _, err := identifier("sns_number", strp("12345")); !errors.As(err, &verr) || verr.Field != "sns_number" {
		t.Errorf("short sns: %v", err)
	}
}

func TestSimilarNames(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"Maria João Silva", "maria silva", true},
		{"JOÃO PEREIRA", "Joao Pereira", true},
		{"Maria Silva", "Maria Santos", false},
		{"Ana", "ana", true},
		{"Ana", "Ana Costa", false},
		{"", "Ana", false},
	}
	for _, c := range cases {
		if got := similarNames(c.a, c.b); got != c.want {
			t.Errorf("similarNames(%q, %q) = %v", c.a, c.b, got)
		}
	}
}

func TestPhoneKey(t *testing.T) {
	if phoneKey(strp("+351 912 345 678")) != phoneKey(strp("912-345-678")) {
		t.Error("country code and separators should not matter")
	}
	if phoneKey(strp("1234")) != "" || phoneKey(nil) != "" {
		t.Error("short or missing numbers are not compared")
	}
}

func TestCreateDuplicateRules(t *testing.T) {
	ctx := context.Background()
	s, _ := newSvc()
	existing, err := s.Create(ctx, nurse, CreateRequest{
		Name: "Maria João Silva", DateOfBirth: strp("1980-05-17"), PhoneNumber: strp("+351 912 345 678"),
		NIF: strp("123456789"), SNSNumber: strp("987654321"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if *existing.NIF != "123456789" || *existing.SNSNumber != "987654321" {
		t.Fatalf("identifiers not stored: %+v", existing)
	}

	cases := []struct {
		name     string
		req      CreateRequest
		reason   string
		blocking bool
	}{
		{"same NIF, different name", CreateRequest{Name: "Other Person", NIF: strp("123 456 789")}, MatchNIF, true},
		{"same SNS number", CreateRequest{Name: "Other Person", SNSNumber: strp("987654321")}, MatchSNS, true},
		{"same NIF even when confirmed", CreateRequest{Name: "Other", NIF: strp("123456789"), AllowDuplicate: true}, MatchNIF, true},
		{"similar name and birth date", CreateRequest{Name: "maria silva", DateOfBirth: strp("1980-05-17")}, MatchNameBirth, false},
		{"similar name and phone", CreateRequest{Name: "Maria Silva", PhoneNumber: strp("912345678")}, MatchNamePhone, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.Create(ctx, receptionist, tc.req)
			var dup *DuplicateError
			if !errors.As(err, &dup) {
				t.Fatalf("want DuplicateError, got %v", err)
			}
			if len(dup.Matches) != 1 || dup.Matches[0].ID != existing.ID || dup.Matches[0].Reason != tc.reason || dup.Blocking() != tc.blocking {
				t.Fatalf("matches: %+v", dup.Matches)
			}
		})
	}

	// A name-based match can be confirmed as a different person
	if _, err := s.Create(ctx, nurse, CreateRequest{Name: "Maria Silva", DateOfBirth: strp("1980-05-17"), AllowDuplicate: true}); err != nil {
		t.Fatalf("confirmed duplicate: %v", err)
	}
	// Same birthday but different name, or same name without other data, is not a duplicate
	for _, req := range []CreateRequest{
		{Name: "Rui Costa", DateOfBirth: strp("1980-05-17")},
		{Name: "Maria João Silva"},
	} {
		if _, err := s.Create(ctx, nurse, req); err != nil {
			t.Errorf("%s: unexpected %v", req.Name, err)
		}
	}
}

func TestUpdateIdentifierRules(t *testing.T) {
	ctx := context.Background()
	s, repo := newSvc()
	a, _ := s.Create(ctx, nurse, CreateRequest{Name: "Ana Costa", NIF: strp("123456789")})
	b, _ := s.Create(ctx, nurse, CreateRequest{Name: "Rui Lopes"})

	var dup *DuplicateError
	if _, err := s.Update(ctx, nurse, b.ID, map[string]any{"nif": "123456789"}); !errors.As(err, &dup) || !dup.Blocking() {
		t.Fatalf("taking another patient's NIF: %v", err)
	}
	if _, err := s.Update(ctx, nurse, a.ID, map[string]any{"nif": "123 456 789", "name": "Ana C."}); err != nil {
		t.Fatalf("re-saving own NIF: %v", err)
	}
	if repo.updates["nif"] != "123456789" {
		t.Errorf("NIF stored normalised: %v", repo.updates["nif"])
	}
	var verr *ValidationError
	if _, err := s.Update(ctx, nurse, b.ID, map[string]any{"nif": "111111111"}); !errors.As(err, &verr) {
		t.Errorf("invalid NIF: %v", err)
	}
	if _, err := s.Update(ctx, nurse, b.ID, map[string]any{"sns_number": ""}); err != nil || repo.updates["sns_number"] != nil {
		t.Errorf("blank clears: %v %v", err, repo.updates["sns_number"])
	}
}

func TestSameDayIgnoresTimezone(t *testing.T) {
	a := time.Date(1980, 5, 17, 0, 0, 0, 0, time.UTC)
	if !sameDay(a, a.In(time.FixedZone("x", 3600))) {
		t.Error("same instant in another zone is the same day")
	}
}
