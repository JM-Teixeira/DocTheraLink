package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func codes(rs []ICDResult) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.Code
	}
	return out
}

func TestEmbeddedDataIsValid(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.codes) < 50 || len(c.templates) < 8 {
		t.Errorf("catalog unexpectedly small: %d codes, %d templates", len(c.codes), len(c.templates))
	}
	// Every template's code is findable by the picker
	for _, tpl := range c.Templates(LangEN) {
		if got := c.Search(tpl.ICDCode, LangEN, 1); len(got) != 1 || got[0].Code != tpl.ICDCode {
			t.Errorf("template %s: code %s not found first (%v)", tpl.Key, tpl.ICDCode, codes(got))
		}
	}
}

func TestSearch(t *testing.T) {
	c := MustLoad()
	cases := []struct {
		name, query, lang string
		first             string
	}{
		{"exact code", "M54.5", LangEN, "M54.5"},
		{"code without dot, lower case", "m545", LangEN, "M54.5"},
		{"English words", "low back", LangEN, "M54.5"},
		{"Portuguese word", "lombalgia", LangPT, "M54.5"},
		{"accent-insensitive", "fasceite", LangPT, "M72.2"},
		{"accented query", "fémur", LangPT, "S72.0"},
		{"word prefix", "frozen", LangEN, "M75.0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := c.Search(tc.query, tc.lang, 5)
			if len(got) == 0 || got[0].Code != tc.first {
				t.Fatalf("Search(%q) = %v, want %s first", tc.query, codes(got), tc.first)
			}
		})
	}

	// A code prefix lists the whole family, in code order
	got := codes(c.Search("M54", LangEN, 10))
	want := []string{"M54.2", "M54.3", "M54.4", "M54.5", "M54.6"}
	if len(got) != len(want) {
		t.Fatalf("M54 family = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("M54 family = %v, want %v", got, want)
		}
	}

	if got := c.Search("lombalgia", LangPT, 1); got[0].Description != "Lombalgia" {
		t.Errorf("Portuguese description: %v", got)
	}
	if got := c.Search("lombalgia", LangEN, 1); got[0].Description != "Low back pain" {
		t.Errorf("English description for a Portuguese query: %v", got)
	}
	if got := c.Search("xyzzy", LangEN, 5); len(got) != 0 {
		t.Errorf("no match expected: %v", got)
	}
	if got := c.Search("", LangEN, 3); len(got) != 3 {
		t.Errorf("empty query returns the first codes up to the limit: %d", len(got))
	}
}

func TestLang(t *testing.T) {
	for in, want := range map[string]string{"pt": LangPT, "pt-PT,pt;q=0.9": LangPT, "PT-br": LangPT, "en-GB": LangEN, "": LangEN, "fr": LangEN} {
		if got := Lang(in); got != want {
			t.Errorf("Lang(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mount(r.Group("/catalog"), NewHandler(MustLoad()))
	get := func(path, acceptLang string, out any) int {
		req := httptest.NewRequest("GET", path, nil)
		if acceptLang != "" {
			req.Header.Set("Accept-Language", acceptLang)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		_ = json.Unmarshal(w.Body.Bytes(), out)
		return w.Code
	}

	var res []ICDResult
	if code := get("/catalog/icd10/?search=ombro&limit=500", "pt-PT", &res); code != http.StatusOK || len(res) == 0 || len(res) > maxLimit {
		t.Fatalf("icd10: %d %v", code, res)
	}
	for _, r := range res {
		if !strings.Contains(fold(r.Description), "ombro") {
			t.Errorf("Accept-Language pt should give Portuguese descriptions: %v", r)
		}
	}

	var tpls []Template
	if code := get("/catalog/referral-templates/?lang=pt", "en", &tpls); code != http.StatusOK || len(tpls) == 0 {
		t.Fatalf("templates: %d", code)
	}
	if tpls[0].Name != "Lombalgia" || tpls[0].ICDCode != "M54.5" || tpls[0].RecommendedSessions != 10 {
		t.Errorf("?lang wins over Accept-Language: %+v", tpls[0])
	}
}
