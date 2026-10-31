package a1_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/sem"
)

// These are independently authored typed-IR requirements, never obtained from
// a production analyzer. They deliberately contain no source sentence strings.
var m1MeaningRequirements = []struct{ id, kind, entity, property string }{
	{"A1-001", "classification", "cat", ""}, {"A1-002", "classification", "dog", ""},
	{"A1-003", "classification", "book", ""}, {"A1-004", "classification", "house", ""},
	{"A1-005", "classification", "table", ""}, {"A1-006", "classification", "chair", ""},
	{"A1-007", "classification", "teacher", ""}, {"A1-008", "classification", "student", ""},
	{"A1-009", "classification", "apple", ""}, {"A1-010", "classification", "window", ""},
	{"A1-011", "property", "cat", "black"}, {"A1-012", "property", "dog", "small"},
	{"A1-013", "property", "book", "new"}, {"A1-014", "property", "bag", "big"},
	{"A1-015", "property", "window", "open"}, {"A1-016", "property", "apple", "red"},
	{"A1-017", "property", "house", "big"}, {"A1-018", "property", "school", "new"},
	{"A1-019", "property", "chair", "small"}, {"A1-020", "property", "coffee", "hot"},
}

func authoredM1Meaning(kind, entity, property string) sem.Meaning {
	ref := sem.Node{ID: "author.ref", Kind: sem.Referent, Reference: context.Proximal}
	predicate := sem.Node{ID: "author.claim", Kind: sem.State}
	switch kind {
	case "classification":
		ref.Concept = "entity.unspecified"
		predicate.Concept = "predicate.is_a"
		predicate.Value = core.ID("entity." + entity)
		predicate.Arguments = []sem.Argument{{Role: "member", Target: ref.ID}}
	case "property":
		ref.Concept = core.ID("entity." + entity)
		predicate.Concept = "predicate.has_property"
		predicate.Value = core.ID("property." + property)
		predicate.Arguments = []sem.Argument{{Role: "bearer", Target: ref.ID}}
	}
	// Intentionally different local IDs and node ordering than the parser.
	return sem.Meaning{Root: predicate.ID, Nodes: []sem.Node{predicate, ref}}
}

func TestM1IndependentMeaningRequirements(t *testing.T) {
	c, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	if len(m1MeaningRequirements) != 20 {
		t.Fatal("expected exactly 20 independent M1 meanings")
	}
	for _, tc := range m1MeaningRequirements {
		t.Run(tc.id, func(t *testing.T) {
			m := authoredM1Meaning(tc.kind, tc.entity, tc.property)
			if err := sem.Validate(m, c); err != nil {
				t.Fatal(err)
			}
		})
	}
}
