// Package textutil holds small text helpers shared by modules.
package textutil

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Fold lower-cases and strips accents ("Fasceíte" → "fasceite", "JOÃO" → "joao").
func Fold(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, err := transform.String(t, s)
	if err != nil {
		out = s
	}
	return strings.ToLower(out)
}

// Words splits folded text into letter/digit runs ("Maria-João Silva" → maria, joao, silva).
func Words(s string) []string {
	return strings.FieldsFunc(Fold(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

// Digits keeps only ASCII digits ("+351 912-345 678" → "351912345678").
func Digits(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
}
