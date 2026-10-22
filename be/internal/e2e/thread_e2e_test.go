//go:build e2e

package e2e

import (
	"testing"
	"time"
)

// TestQuestionThread: a question waits (badge + filter) for the other side
// until they answer that question; progress notes keep the pain score.
func TestQuestionThread(t *testing.T) {
	doctor, _ := login(t, "doctor")
	sarah, _ := login(t, "physiotherapist")
	var p, ref map[string]any
	expect(t, "patient", doctor.call("POST", "/patient/", map[string]any{"name": "Pergunta " + time.Now().Format("150405.000")}, &p), 201)
	expect(t, "referral", doctor.call("POST", "/referrals/", map[string]any{
		"patient_id": p["id"], "diagnosis": "Gonalgia", "reason": "Fortalecimento", "therapist_id": userIDOf(sarah),
	}, &ref), 201)
	rid := ref["id"].(string)
	expect(t, "accept", sarah.call("POST", "/referrals/"+rid+"/accept/", map[string]any{}, nil), 200)

	var q map[string]any
	expect(t, "doctor asks", doctor.call("POST", "/referrals/"+rid+"/notes/", map[string]any{"kind": "question", "body": "Há derrame?"}, &q), 201)
	openFor := func(c *client) float64 {
		var page struct {
			Results []map[string]any `json:"results"`
		}
		expect(t, "waiting filter", c.call("GET", "/referrals/?open_questions=true&limit=100", nil, &page), 200)
		for _, r := range page.Results {
			if r["id"] == rid {
				return r["open_questions"].(float64)
			}
		}
		return 0
	}
	if openFor(sarah) != 1 || openFor(doctor) != 0 {
		t.Fatalf("waiting for Sarah only: sarah %v, doctor %v", openFor(sarah), openFor(doctor))
	}
	expect(t, "reply without its question", sarah.call("POST", "/referrals/"+rid+"/notes/", map[string]any{"kind": "reply", "body": "Sim"}, nil), 400)
	expect(t, "reply", sarah.call("POST", "/referrals/"+rid+"/notes/", map[string]any{"kind": "reply", "body": "Ligeiro", "reply_to": q["id"]}, nil), 201)
	if openFor(sarah) != 0 {
		t.Error("answered question still waiting")
	}
	var pn map[string]any
	expect(t, "progress with pain", sarah.call("POST", "/referrals/"+rid+"/notes/", map[string]any{"kind": "progress", "body": "Melhor", "pain": 3}, &pn), 201)
	expect(t, "pain out of range", sarah.call("POST", "/referrals/"+rid+"/notes/", map[string]any{"kind": "progress", "body": "x", "pain": 11}, nil), 400)
	var notes []map[string]any
	expect(t, "thread", doctor.call("GET", "/referrals/"+rid+"/notes/", nil, &notes), 200)
	for _, n := range notes {
		if n["id"] == q["id"] && n["answered"] != true {
			t.Error("question not marked answered")
		}
		if n["id"] == pn["id"] && n["pain"] != 3.0 {
			t.Errorf("pain: %v", n["pain"])
		}
	}
}
