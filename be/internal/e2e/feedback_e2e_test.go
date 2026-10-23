//go:build e2e

package e2e

import (
	"testing"
	"time"
)

// TestSuggestions: anyone sends; only admins list all and answer; the
// author sees the answer.
func TestSuggestions(t *testing.T) {
	nurse, _ := login(t, "cm.teresa.gomes")
	admin, _ := login(t, "admin.user")
	text := "Sugestão e2e " + time.Now().Format("150405.000")
	var s map[string]any
	expect(t, "send", nurse.call("POST", "/suggestions/", map[string]any{"kind": "idea", "text": text}, &s), 201)
	expect(t, "too short", nurse.call("POST", "/suggestions/", map[string]any{"kind": "idea", "text": "ok"}, nil), 400)
	expect(t, "nurse lists all", nurse.call("GET", "/suggestions/", nil, nil), 403)
	var count map[string]int
	expect(t, "count", admin.call("GET", "/suggestions/count/", nil, &count), 200)
	if count["new"] < 1 {
		t.Errorf("new count: %v", count)
	}
	expect(t, "answer", admin.call("PATCH", "/suggestions/"+s["id"].(string)+"/", map[string]any{"status": "done", "admin_note": "Feito."}, nil), 200)
	var mine []map[string]any
	expect(t, "mine", nurse.call("GET", "/suggestions/mine/", nil, &mine), 200)
	found := false
	for _, m := range mine {
		if m["text"] == text {
			found = m["status"] == "done" && m["admin_note"] == "Feito."
		}
	}
	if !found {
		t.Errorf("author does not see the answer: %v", mine)
	}
	expect(t, "nurse answers", nurse.call("PATCH", "/suggestions/"+s["id"].(string)+"/", map[string]any{"status": "read"}, nil), 403)
}
