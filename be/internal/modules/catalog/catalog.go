// Package catalog serves read-only reference data for the referral form:
// a curated ICD-10 subset relevant to physiotherapy and built-in referral
// templates for common conditions, both in English and Portuguese.
//
// The data is embedded (icd10.json, templates.json) and validated at start-up.
// Clinic-specific templates stored in the database are a later step (TODO).
package catalog

import (
	"embed"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"klyniva/internal/textutil"
)

//go:embed icd10.json templates.json
var files embed.FS

// Supported languages; anything else falls back to English.
const (
	LangEN = "en"
	LangPT = "pt"
)

// ICDResult is one code as returned by the API, in the requested language.
type ICDResult struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// Template pre-fills the referral form, in the requested language.
type Template struct {
	Key                 string `json:"key"`
	Name                string `json:"name"`
	ICDCode             string `json:"icd_code"`
	Diagnosis           string `json:"diagnosis"`
	Reason              string `json:"reason"`
	Goals               string `json:"goals"`
	Precautions         string `json:"precautions"`
	Priority            string `json:"priority"`
	RecommendedSessions int    `json:"recommended_sessions"`
}

type code struct {
	Code string `json:"code"`
	EN   string `json:"en"`
	PT   string `json:"pt"`
}

type templateText struct {
	Name        string `json:"name"`
	Diagnosis   string `json:"diagnosis"`
	Reason      string `json:"reason"`
	Goals       string `json:"goals"`
	Precautions string `json:"precautions"`
}

type template struct {
	Key                 string                  `json:"key"`
	ICDCode             string                  `json:"icd_code"`
	Priority            string                  `json:"priority"`
	RecommendedSessions int                     `json:"recommended_sessions"`
	Text                map[string]templateText `json:"text"`
}

// Catalog holds the loaded reference data. It is immutable and safe for
// concurrent use.
type Catalog struct {
	codes     []code
	keys      []string // folded code without dot, e.g. "m545"
	words     [][]string
	templates []template
}

var codeFormat = regexp.MustCompile(`^[A-Z][0-9]{2}(\.[0-9A-Z]{1,2})?$`)

// Load parses and validates the embedded data.
func Load() (*Catalog, error) {
	c := &Catalog{}
	if err := readJSON("icd10.json", &c.codes); err != nil {
		return nil, err
	}
	if err := readJSON("templates.json", &c.templates); err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	for _, k := range c.codes {
		if !codeFormat.MatchString(k.Code) || k.EN == "" || k.PT == "" {
			return nil, fmt.Errorf("catalog: invalid ICD entry %+v", k)
		}
		if seen[k.Code] {
			return nil, fmt.Errorf("catalog: duplicate ICD code %s", k.Code)
		}
		seen[k.Code] = true
		c.keys = append(c.keys, codeKey(k.Code))
		c.words = append(c.words, strings.FieldsFunc(textutil.Fold(k.EN+" "+k.PT), notAlnum))
	}

	keys := map[string]bool{}
	for _, t := range c.templates {
		if t.Key == "" || keys[t.Key] {
			return nil, fmt.Errorf("catalog: missing or duplicate template key %q", t.Key)
		}
		keys[t.Key] = true
		if !seen[t.ICDCode] {
			return nil, fmt.Errorf("catalog: template %s uses unknown ICD code %s", t.Key, t.ICDCode)
		}
		if t.Priority != "routine" && t.Priority != "urgent" {
			return nil, fmt.Errorf("catalog: template %s has invalid priority %q", t.Key, t.Priority)
		}
		for _, lang := range []string{LangEN, LangPT} {
			tx, ok := t.Text[lang]
			if !ok || tx.Name == "" || tx.Diagnosis == "" || tx.Reason == "" {
				return nil, fmt.Errorf("catalog: template %s lacks %s text", t.Key, lang)
			}
		}
	}
	return c, nil
}

// MustLoad is Load for start-up code: the data is embedded, so an error is a
// programming mistake caught by the tests.
func MustLoad() *Catalog {
	c, err := Load()
	if err != nil {
		panic(err)
	}
	return c
}

// Lang normalises a language tag ("pt-PT", "pt", "en-GB") to LangPT or LangEN.
func Lang(tag string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(tag)), "pt") {
		return LangPT
	}
	return LangEN
}

// Search finds codes by code ("m54", "M54.5", "m545") or by words of the
// description in either language, accent-insensitive ("lombalgia",
// "low back", "ombro"). Results come in the requested language, best match
// first: exact code, code prefix, word prefixes, then substrings.
func (c *Catalog) Search(query, lang string, limit int) []ICDResult {
	q := textutil.Fold(strings.TrimSpace(query))
	qKey := codeKey(query)
	terms := strings.FieldsFunc(q, notAlnum)

	type hit struct{ i, rank int }
	var hits []hit
	for i := range c.codes {
		rank := -1
		switch {
		case len(terms) == 0:
			rank = 4
		case qKey != "" && c.keys[i] == qKey:
			rank = 0
		case qKey != "" && strings.HasPrefix(c.keys[i], qKey):
			rank = 1
		case all(terms, func(t string) bool { return anyWord(c.words[i], strings.HasPrefix, t) }):
			rank = 2
		case all(terms, func(t string) bool { return anyWord(c.words[i], strings.Contains, t) }):
			rank = 3
		}
		if rank >= 0 {
			hits = append(hits, hit{i, rank})
		}
	}
	sort.SliceStable(hits, func(a, b int) bool { return hits[a].rank < hits[b].rank })

	out := make([]ICDResult, 0, min(limit, len(hits)))
	for _, h := range hits {
		if len(out) == limit {
			break
		}
		out = append(out, ICDResult{Code: c.codes[h.i].Code, Description: c.describe(h.i, lang)})
	}
	return out
}

// Templates returns the built-in referral templates in the requested language.
func (c *Catalog) Templates(lang string) []Template {
	out := make([]Template, len(c.templates))
	for i, t := range c.templates {
		tx := t.Text[Lang(lang)]
		out[i] = Template{
			Key: t.Key, Name: tx.Name, ICDCode: t.ICDCode, Diagnosis: tx.Diagnosis,
			Reason: tx.Reason, Goals: tx.Goals, Precautions: tx.Precautions,
			Priority: t.Priority, RecommendedSessions: t.RecommendedSessions,
		}
	}
	return out
}

func (c *Catalog) describe(i int, lang string) string {
	if Lang(lang) == LangPT {
		return c.codes[i].PT
	}
	return c.codes[i].EN
}

func readJSON(name string, v any) error {
	b, err := files.ReadFile(name)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("catalog: %s: %w", name, err)
	}
	return nil
}

// codeKey turns "M54.5" / "m 54.5" / "M545" into "m545"; "" if the text
// cannot be a code (codes start with a letter followed by a digit).
func codeKey(s string) string {
	k := strings.ToLower(strings.NewReplacer(".", "", " ", "").Replace(strings.TrimSpace(s)))
	if len(k) < 2 || k[0] < 'a' || k[0] > 'z' || k[1] < '0' || k[1] > '9' {
		return ""
	}
	return k
}

func notAlnum(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }

func all(terms []string, ok func(string) bool) bool {
	for _, t := range terms {
		if !ok(t) {
			return false
		}
	}
	return true
}

func anyWord(words []string, match func(string, string) bool, term string) bool {
	for _, w := range words {
		if match(w, term) {
			return true
		}
	}
	return false
}
