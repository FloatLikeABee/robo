package handlers

import "testing"

func TestParseCoordinatorJSON(t *testing.T) {
	raw := `{"plan":"Check data","memory_patch":"User prefers metric units","sub_agents":["dorphdata","files","knowledge","extra"]}`
	got, ok := parseCoordinatorJSON(raw)
	if !ok {
		t.Fatal("parse failed")
	}
	if got.Plan != "Check data" {
		t.Fatalf("plan = %q", got.Plan)
	}
	if len(got.SubAgents) != 3 {
		t.Fatalf("sub_agents = %v", got.SubAgents)
	}
}

func TestIntersectSubAgents(t *testing.T) {
	kw := []string{"files", "dorphdata"}
	co := []string{"dorphdata", "notes"}
	got := intersectSubAgents(kw, co)
	if len(got) != 1 || got[0] != "dorphdata" {
		t.Fatalf("got %v", got)
	}
}
