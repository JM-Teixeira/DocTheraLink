package clinic

import (
	"context"
	"testing"
)

func TestSameClinic(t *testing.T) {
	a, b := "clinic-a", "clinic-b"
	cases := []struct {
		x, y *string
		want bool
	}{
		{&a, &a, true},
		{&a, &b, false},
		{nil, &a, true},
		{&a, nil, true},
		{nil, nil, true},
	}
	for _, c := range cases {
		if got := SameClinic(c.x, c.y); got != c.want {
			t.Errorf("SameClinic(%v,%v) = %v, want %v", c.x, c.y, got, c.want)
		}
	}
}

func TestStaticResolver(t *testing.T) {
	r := Static{1: "clinic-a"}
	if c, _ := r.HomeClinic(context.Background(), 1); c == nil || *c != "clinic-a" {
		t.Fatalf("user 1: %v", c)
	}
	if c, _ := r.HomeClinic(context.Background(), 2); c != nil {
		t.Fatalf("user 2 should have no clinic, got %v", *c)
	}
}
