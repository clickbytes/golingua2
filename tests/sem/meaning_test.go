package sem_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/sem"
)

func testCatalog(t *testing.T) *core.Catalog {
	t.Helper()
	c, err := core.NewCatalog([]core.Definition{
		{ID: "entity.cat", Kind: core.Entity},
		{ID: "entity.dog", Kind: core.Entity},
		{ID: "predicate.related", Kind: core.Predicate, ValueKind: core.Entity, Roles: []core.Role{{Name: "left", Target: core.Entity}, {Name: "right", Target: core.Entity}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func twoCats(shared bool) sem.Meaning {
	a := sem.Node{ID: "cat-a", Kind: sem.Referent, Concept: "entity.cat", Reference: context.Unresolved}
	b := sem.Node{ID: "cat-b", Kind: sem.Referent, Concept: "entity.cat", Reference: context.Unresolved}
	target := b.ID
	if shared {
		target = a.ID
	}
	predicate := sem.Node{ID: "relation", Kind: sem.State, Concept: "predicate.related", Value: "entity.dog", Arguments: []sem.Argument{{Role: "left", Target: a.ID}, {Role: "right", Target: target}}}
	if shared {
		return sem.Meaning{Root: predicate.ID, Nodes: []sem.Node{a, predicate}}
	}
	return sem.Meaning{Root: predicate.ID, Nodes: []sem.Node{b, predicate, a}}
}

func TestMeaningPreservesRepeatedOccurrenceIdentity(t *testing.T) {
	c := testCatalog(t)
	distinct, shared := twoCats(false), twoCats(true)
	if err := sem.Validate(distinct, c); err != nil {
		t.Fatal(err)
	}
	if err := sem.Validate(shared, c); err != nil {
		t.Fatal(err)
	}
	same, err := sem.Equivalent(distinct, shared, c)
	if err != nil || same {
		t.Fatalf("two cats incorrectly collapsed: same=%v error=%v", same, err)
	}
	renamed := distinct.Clone()
	renamed.Root = "claim"
	renamed.Nodes[0].ID = "first"
	renamed.Nodes[1].ID = "claim"
	renamed.Nodes[1].Arguments = []sem.Argument{{Role: "right", Target: "first"}, {Role: "left", Target: "second"}}
	renamed.Nodes[2].ID = "second"
	same, err = sem.Equivalent(distinct, renamed, c)
	if err != nil || !same {
		t.Fatalf("local ID renaming changed meaning: same=%v error=%v", same, err)
	}
}

func TestMeaningRefusesIncompleteAndWrongType(t *testing.T) {
	c := testCatalog(t)
	for _, bad := range []sem.Meaning{
		{Root: "x", Nodes: []sem.Node{{ID: "x", Kind: sem.State, Concept: "predicate.related", Value: "entity.dog"}}},
		{Root: "relation", Nodes: []sem.Node{{ID: "relation", Kind: sem.State, Concept: "predicate.related", Value: "entity.dog", Arguments: []sem.Argument{{Role: "left", Target: "missing"}, {Role: "right", Target: "missing"}}}}},
		{Root: "relation", Nodes: []sem.Node{{ID: "relation", Kind: sem.State, Concept: "predicate.related", Value: "predicate.related", Arguments: []sem.Argument{{Role: "left", Target: "x"}, {Role: "right", Target: "y"}}}}},
	} {
		if err := sem.Validate(bad, c); err == nil {
			t.Fatalf("invalid graph was accepted: %+v", bad)
		}
	}
}
