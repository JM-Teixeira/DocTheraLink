package textutil

import (
	"slices"
	"testing"
)

func TestFold(t *testing.T) {
	for in, want := range map[string]string{"Fasceíte": "fasceite", "JOÃO Conceição": "joao conceicao", "Ünïcode": "unicode", "": ""} {
		if got := Fold(in); got != want {
			t.Errorf("Fold(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWordsAndDigits(t *testing.T) {
	if got := Words("  Maria-João  Silva "); !slices.Equal(got, []string{"maria", "joao", "silva"}) {
		t.Errorf("Words = %v", got)
	}
	if got := Digits("+351 912-345 678"); got != "351912345678" {
		t.Errorf("Digits = %q", got)
	}
}
