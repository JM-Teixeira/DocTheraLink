// Package usernames builds sign-in names from a person's name and clinic:
//
//	<clinic code>.<first name>.<last surname>     cm.joao.teixeira
//	<first name>.<last surname>                   (organisation admins, no clinic)
//
// Lower case, no accents, letters and digits only in each part. When the
// name is taken a number is added: cm.joao.teixeira2, cm.joao.teixeira3 …
// Using the full first name and the last surname (how people are usually
// called in Portugal) keeps names readable; the number settles clashes such
// as two João Teixeira, which initials would not.
package usernames

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Max is the longest generated username (the column allows 150).
const Max = 60

var codeRe = regexp.MustCompile(`^[a-z]{2,4}$`)

// ValidCode reports whether s is a clinic code: 2 to 4 lower-case letters.
func ValidCode(s string) bool { return codeRe.MatchString(s) }

// part folds a name to lower-case ASCII letters and digits ("Inês" → "ines").
func part(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	folded, _, _ := transform.String(t, s)
	var b strings.Builder
	for _, r := range strings.ToLower(folded) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Base is the username before any clash number: code, first word of the
// first name, last word of the last name (falls back to what is present).
func Base(code, first, last string) string {
	f := strings.Fields(first)
	l := strings.Fields(last)
	var parts []string
	if c := part(code); c != "" {
		parts = append(parts, c)
	}
	if len(f) > 0 {
		if p := part(f[0]); p != "" {
			parts = append(parts, p)
		}
	}
	if len(l) > 0 {
		if p := part(l[len(l)-1]); p != "" {
			parts = append(parts, p)
		}
	}
	base := strings.Join(parts, ".")
	if len(base) > Max-3 {
		base = strings.TrimRight(base[:Max-3], ".")
	}
	return base
}

// Next returns the first free username for the person: Base, then Base2, Base3 …
// taken reports whether a username already exists.
func Next(code, first, last string, taken func(string) (bool, error)) (string, error) {
	base := Base(code, first, last)
	if base == "" || !strings.Contains(base, ".") && code == "" && len(base) < 3 {
		return "", fmt.Errorf("a first and last name are needed")
	}
	for n := 1; n < 1000; n++ {
		name := base
		if n > 1 {
			name = fmt.Sprintf("%s%d", base, n)
		}
		used, err := taken(name)
		if err != nil {
			return "", err
		}
		if !used {
			return name, nil
		}
	}
	return "", fmt.Errorf("no free username for %s", base)
}
