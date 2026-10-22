//go:build e2e

package e2e

import (
	"fmt"
	"net/url"
	"testing"
)

// TestStockAccessByFloorAndCategory: in Climutua, Inês may change the
// floor-1 stock, Mariana the hygiene items of floors 2 and 3, Pedro nothing;
// the coordinator gives and takes these permissions.
func TestStockAccessByFloorAndCategory(t *testing.T) {
	coord, _ := login(t, "cm.coordenacao")
	ines, _ := login(t, "cm.fisio1")
	mariana, _ := login(t, "cm.fisio5")
	pedro, _ := login(t, "cm.fisio2")
	other, _ := login(t, "coordinator")

	item := func(name string) map[string]any {
		var items []map[string]any
		expect(t, "find "+name, coord.call("GET", "/inventory/items/?search="+url.QueryEscape(name), nil, &items), 200)
		if len(items) != 1 {
			t.Fatalf("%s: %d items", name, len(items))
		}
		return items[0]
	}
	caps, tens, antiseptic := item("Toucas de natação"), item("Elétrodos para TENS"), item("Solução antisséptica")
	move := func(c *client, it map[string]any, kind string) int {
		return c.call("POST", fmt.Sprintf("/inventory/items/%s/movements/", it["id"]), map[string]any{"kind": kind, "quantity": 1}, nil)
	}

	expect(t, "Inês uses floor-1 caps", move(ines, caps, "use"), 200)
	expect(t, "Inês restocks them", move(ines, caps, "restock"), 200)
	expect(t, "Inês cannot touch floor-2 TENS", move(ines, tens, "use"), 403)
	expect(t, "Mariana restocks floor-2 hygiene", move(mariana, antiseptic, "restock"), 200)
	expect(t, "Mariana cannot touch floor-1 caps", move(mariana, caps, "use"), 403)
	expect(t, "Pedro has no permission", move(pedro, caps, "use"), 403)

	// The list says what each one may change
	var list []map[string]any
	expect(t, "Pedro still sees the stock", pedro.call("GET", "/inventory/items/", nil, &list), 200)
	for _, it := range list {
		if it["can_change"] == true {
			t.Fatalf("Pedro can change %v", it["name"])
		}
	}

	var pedroMe map[string]any
	pedro.call("GET", "/users/getcurrentuser/", nil, &pedroMe)
	path := fmt.Sprintf("/inventory/grants/%v/", pedroMe["id"])
	rules := map[string]any{"rules": []map[string]any{{"floors": []int{2}, "categories": []string{}}}}
	expect(t, "physios do not give permissions", ines.call("PUT", path, rules, nil), 403)
	expect(t, "another clinic's coordinator", other.call("PUT", path, rules, nil), 404)
	var grants []map[string]any
	expect(t, "coordinator gives Pedro floor 2", coord.call("PUT", path, rules, &grants), 200)
	if len(grants) != 1 {
		t.Fatalf("grants: %v", grants)
	}
	defer coord.call("PUT", path, map[string]any{"rules": []any{}}, nil)
	expect(t, "Pedro restocks floor-2 TENS", move(pedro, tens, "restock"), 200)
	expect(t, "but not floor-1 caps", move(pedro, caps, "use"), 403)
	var all []map[string]any
	expect(t, "coordinator lists permissions", coord.call("GET", "/inventory/grants/", nil, &all), 200)
	if len(all) < 3 {
		t.Errorf("clinic grants: %d", len(all))
	}
	expect(t, "coordinator takes it away", coord.call("PUT", path, map[string]any{"rules": []any{}}, nil), 200)
	expect(t, "Pedro is read-only again", move(pedro, tens, "restock"), 403)
}

// TestInventoryClinicFilter: admins look at one clinic at a time if they want.
func TestInventoryClinicFilter(t *testing.T) {
	admin, _ := login(t, "stateadmin")
	physio, _ := login(t, "physiotherapist")
	const climutua = "6b1f3c1e-5a8e-4c1d-9a0e-000000000002"
	var all, one []map[string]any
	expect(t, "all clinics", admin.call("GET", "/inventory/items/", nil, &all), 200)
	expect(t, "Climutua only", admin.call("GET", "/inventory/items/?clinic="+climutua, nil, &one), 200)
	if len(one) == 0 || len(one) >= len(all) {
		t.Fatalf("filter: %d of %d", len(one), len(all))
	}
	for _, it := range one {
		if it["clinic_id"] != climutua {
			t.Fatalf("other clinic leaked: %v", it["clinic_id"])
		}
	}
	expect(t, "staff cannot peek at another clinic", physio.call("GET", "/inventory/items/?clinic="+climutua, nil, nil), 403)
}
