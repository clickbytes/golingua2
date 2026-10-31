package sem_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/sem"
)

func TestM4StateAndEventMeaningKindsRemainDistinct(t *testing.T) {
	catalog, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	a := sem.Node{ID: "a", Kind: sem.Referent, Concept: "entity.cat", Number: "singular"}
	b := sem.Node{ID: "b", Kind: sem.Referent, Concept: "entity.house", Number: "singular"}
	relation := sem.Node{ID: "p", Kind: sem.State, Concept: "relation.inside", Arguments: []sem.Argument{{Role: "located", Target: "a"}, {Role: "site", Target: "b"}}}
	good := sem.Meaning{Root: "p", Nodes: []sem.Node{relation, a, b}}
	if err := sem.Validate(good, catalog); err != nil {
		t.Fatal(err)
	}
	for _, m := range []sem.Meaning{
		{Root: "p", Nodes: []sem.Node{{ID: "p", Kind: sem.Event, Concept: "relation.inside", Tense: "present", Arguments: relation.Arguments}, a, b}},
		{Root: "p", Nodes: []sem.Node{{ID: "p", Kind: sem.State, Concept: "predicate.see", Arguments: []sem.Argument{{Role: "agent", Target: "a"}, {Role: "patient", Target: "b"}}}, a, b}},
		{Root: "p", Nodes: []sem.Node{{ID: "p", Kind: sem.State, Concept: "relation.inside", Value: "property.black", Arguments: relation.Arguments}, a, b}},
		{Root: "p", Nodes: []sem.Node{{ID: "p", Kind: sem.State, Concept: "relation.inside", Tense: "present", Arguments: relation.Arguments}, a, b}},
	} {
		if err := sem.Validate(m, catalog); err == nil {
			t.Fatal("event/state kinds conflated")
		}
	}
	swapped := good.Clone()
	swapped.Nodes[0].Arguments[0].Target = "b"
	swapped.Nodes[0].Arguments[1].Target = "a"
	if err := sem.Validate(swapped, catalog); err != nil {
		t.Fatal(err)
	}
	same, err := sem.Equivalent(good, swapped, catalog)
	if err != nil || same {
		t.Fatalf("location role swap must change meaning, err=%v", err)
	}
	repeated := good.Clone()
	repeated.Nodes[0].Arguments[1].Target = "a"
	repeated.Nodes = repeated.Nodes[:2] // orphan house removed, self-location remains logically distinct
	if err := sem.Validate(repeated, catalog); err != nil {
		t.Fatal(err)
	}
	same, err = sem.Equivalent(good, repeated, catalog)
	if err != nil || same {
		t.Fatalf("shared referent was conflated with two referents, err=%v", err)
	}
}
