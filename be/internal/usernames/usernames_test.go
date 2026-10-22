package usernames

import (
	"errors"
	"testing"
)

func TestBase(t *testing.T) {
	cases := []struct{ code, first, last, want string }{
		{"cm", "João", "Teixeira", "cm.joao.teixeira"},
		{"cm", "Inês Maria", "Moreira da Silva", "cm.ines.silva"},
		{"de", "Ana-Rita", "D'Ávila", "de.anarita.davila"},
		{"", "Paula", "Neves", "paula.neves"},
		{"cm", "  ", "Costa", "cm.costa"},
	}
	for _, c := range cases {
		if got := Base(c.code, c.first, c.last); got != c.want {
			t.Errorf("Base(%q,%q,%q) = %q, want %q", c.code, c.first, c.last, got, c.want)
		}
	}
}

func TestNextAddsNumberOnClash(t *testing.T) {
	used := map[string]bool{"cm.joao.teixeira": true, "cm.joao.teixeira2": true}
	taken := func(s string) (bool, error) { return used[s], nil }
	got, err := Next("cm", "João", "Teixeira", taken)
	if err != nil || got != "cm.joao.teixeira3" {
		t.Errorf("got %q %v", got, err)
	}
	if _, err := Next("", "", "", taken); err == nil {
		t.Error("empty names must fail")
	}
	boom := func(string) (bool, error) { return false, errors.New("db") }
	if _, err := Next("cm", "Ana", "Lopes", boom); err == nil {
		t.Error("lookup errors propagate")
	}
}

func TestValidCode(t *testing.T) {
	for s, want := range map[string]bool{"cm": true, "de": true, "abcd": true, "c": false, "CM": false, "c1": false, "abcde": false} {
		if ValidCode(s) != want {
			t.Errorf("ValidCode(%q) != %v", s, want)
		}
	}
}
