package version

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"
)

func TestVersionIsSemverAndMatchesFrontend(t *testing.T) {
	if !regexp.MustCompile(`^0\.\d+\.\d+$`).MatchString(Version) {
		t.Fatalf("version %q must be 0.minor.patch", Version)
	}
	raw, err := os.ReadFile("../../../fe/package.json")
	if err != nil {
		t.Skipf("frontend not next to the backend: %v", err)
	}
	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &pkg); err != nil {
		t.Fatal(err)
	}
	if pkg.Version != Version {
		t.Fatalf("fe/package.json has %s, backend %s: run scripts/bump-version.sh", pkg.Version, Version)
	}
}
