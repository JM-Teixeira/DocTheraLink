package patient

import (
	"strings"

	"klyniva/internal/textutil"
)

// Portuguese identifiers.
//
//	NIF  - Número de Identificação Fiscal: 9 digits, the last is a mod-11
//	       check digit (weights 9..2 over the first eight).
//	SNS  - Número de utente do Serviço Nacional de Saúde: 9 digits. Its check
//	       rule is not published, so only the format is validated.
//
// Both are stored as 9 plain digits; spaces, dots and dashes typed by users
// are removed first.

// normalizeID strips separators users commonly type ("123 456 789", "123.456.789").
func normalizeID(s string) string {
	return strings.NewReplacer(" ", "", ".", "", "-", "").Replace(strings.TrimSpace(s))
}

func nineDigits(s string) bool {
	return len(s) == 9 && textutil.Digits(s) == s
}

// ValidNIF reports whether a normalised NIF has a correct check digit.
func ValidNIF(nif string) bool {
	if !nineDigits(nif) {
		return false
	}
	sum := 0
	for i := 0; i < 8; i++ {
		sum += int(nif[i]-'0') * (9 - i)
	}
	check := 11 - sum%11
	if check >= 10 {
		check = 0
	}
	return int(nif[8]-'0') == check
}

// ValidSNS reports whether a normalised SNS number has the right format.
func ValidSNS(sns string) bool { return nineDigits(sns) }

// identifier normalises and validates an optional identifier field; empty
// input clears it (nil).
func identifier(field string, v *string) (*string, error) {
	if v == nil {
		return nil, nil
	}
	n := normalizeID(*v)
	if n == "" {
		return nil, nil
	}
	valid := ValidSNS(n)
	if field == "nif" {
		valid = ValidNIF(n)
	}
	if !valid {
		return nil, &ValidationError{field, "is not a valid number"}
	}
	return &n, nil
}

// similarNames treats two names as the same person's when, ignoring case and
// accents, they share first and last name ("Maria João Silva" ~ "maria silva").
// Single-word names must match exactly.
func similarNames(a, b string) bool {
	wa, wb := textutil.Words(a), textutil.Words(b)
	if len(wa) == 0 || len(wb) == 0 {
		return false
	}
	if len(wa) == 1 || len(wb) == 1 {
		return strings.Join(wa, " ") == strings.Join(wb, " ")
	}
	return wa[0] == wb[0] && wa[len(wa)-1] == wb[len(wb)-1]
}

// phoneKey compares phone numbers by their last 9 digits, so "+351 912 345 678"
// and "912345678" match. Shorter numbers are not compared.
func phoneKey(s *string) string {
	if s == nil {
		return ""
	}
	d := textutil.Digits(*s)
	if len(d) < 9 {
		return ""
	}
	return d[len(d)-9:]
}
