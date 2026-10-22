//go:build e2e

package e2e

import (
	"testing"
	"time"
)

// TestTreatmentPlan: the physio writes the plan once the referral is
// accepted; every save is a version; the doctor reads it; another physio
// does not see it.
func TestTreatmentPlan(t *testing.T) {
	doctor, _ := login(t, "cm.helena.carvalho")
	joao, _ := login(t, "cm.joao.ribeiro")
	pedro, _ := login(t, "cm.pedro.almeida")
	desk, _ := login(t, "cm.sofia.martins")
	var p, ref map[string]any
	expect(t, "patient", desk.call("POST", "/patient/", map[string]any{"name": "Plano " + time.Now().Format("150405.000")}, &p), 201)
	expect(t, "referral", doctor.call("POST", "/referrals/", map[string]any{
		"patient_id": p["id"], "diagnosis": "Tendinopatia do supraespinhoso", "reason": "Dor no ombro", "therapist_id": userIDOf(joao),
	}, &ref), 201)
	path := "/referrals/" + ref["id"].(string) + "/plan/"
	plan := map[string]any{"items": []any{
		map[string]any{"activity": "Ultrassons", "minutes": 10, "equipment": "Máquina 2"},
		map[string]any{"activity": "Massagem", "minutes": 5},
	}}
	expect(t, "not before accepting", joao.call("PUT", path, plan, nil), 403)
	expect(t, "accept", joao.call("POST", "/referrals/"+ref["id"].(string)+"/accept/", map[string]any{}, nil), 200)
	expect(t, "save", joao.call("PUT", path, plan, nil), 200)
	plan["items"] = append(plan["items"].([]any), map[string]any{"activity": "Codman", "dose": "3x10"})
	var v struct {
		Current struct {
			Items []map[string]any `json:"items"`
		} `json:"current"`
		Versions []any `json:"versions"`
		CanEdit  bool  `json:"can_edit"`
	}
	expect(t, "second version", joao.call("PUT", path, plan, &v), 200)
	if len(v.Current.Items) != 3 || len(v.Versions) != 2 || !v.CanEdit {
		t.Errorf("after two saves: %+v", v)
	}
	expect(t, "doctor reads", doctor.call("GET", path, nil, &v), 200)
	if v.CanEdit || len(v.Current.Items) != 3 {
		t.Errorf("doctor view: %+v", v)
	}
	expect(t, "doctor cannot edit", doctor.call("PUT", path, plan, nil), 403)
	expect(t, "other physio does not see it", pedro.call("GET", path, nil, nil), 404)
	expect(t, "bad minutes", joao.call("PUT", path, map[string]any{"items": []any{map[string]any{"activity": "x", "minutes": 0}}}, nil), 400)
}
