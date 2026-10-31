package a1_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/sem"
)

// Independently authored spatial relation requirements, not reconstructed
// from the output of either language parser.
type m4Spec struct {
	id, relation, located, site string
	bareEnglishSite             bool
}

var m4Requirements = []m4Spec{
	{"A1-071", "inside", "cat", "house", false},
	{"A1-072", "on_support", "book", "table", false},
	{"A1-073", "below", "dog", "table", false},
	{"A1-074", "at_site", "student", "school", true},
	{"A1-075", "inside", "mother", "kitchen", false},
	{"A1-076", "in_front", "bicycle", "house", false},
	{"A1-077", "behind", "car", "house", false},
	{"A1-078", "adjacent", "child", "mother", false},
	{"A1-079", "at_site", "bird", "tree", false},
	{"A1-080", "inside", "table", "room", false},
}

func authoredM4Meaning(s m4Spec, language string) sem.Meaning {
	located := sem.Node{ID: "independent.located", Kind: sem.Referent, Concept: core.ID("entity." + s.located), Number: "singular"}
	site := sem.Node{ID: "independent.site", Kind: sem.Referent, Concept: core.ID("entity." + s.site), Number: "singular"}
	if language == "en" {
		located.Definiteness = "definite"
		if !s.bareEnglishSite {
			site.Definiteness = "definite"
		}
	}
	relation := sem.Node{ID: "independent.spatial-relation", Kind: sem.State, Concept: core.ID("relation." + s.relation), Arguments: []sem.Argument{
		{Role: "located", Target: located.ID}, {Role: "site", Target: site.ID},
	}}
	return sem.Meaning{Root: relation.ID, Nodes: []sem.Node{site, relation, located}}
}

func TestM4IndependentSpatialMeaning(t *testing.T) {
	c, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	if len(m4Requirements) != 10 {
		t.Fatalf("want 10 location requirements: %d", len(m4Requirements))
	}
	for i, s := range m4Requirements {
		if cases[70+i].ID != s.id {
			t.Fatalf("mismatch case ID %q", s.id)
		}
		for _, language := range []string{"pl", "en"} {
			if err := sem.Validate(authoredM4Meaning(s, language), c); err != nil {
				t.Errorf("%s/%s: %v", s.id, language, err)
			}
		}
	}
}
