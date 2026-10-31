package a1_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/sem"
)

// Author-written predicate/argument requirements: no production parser is used
// to obtain an expected IR, and the source string is never a grammar rule.
type m3Spec struct {
	id, predicate, subject, object, person, number string
	mass                                           bool
}

var m3Requirements = []m3Spec{
	{"A1-051", "sleep", "cat", "", "", "", false},
	{"A1-052", "run", "dog", "", "", "", false},
	{"A1-053", "read", "teacher", "book", "", "", false},
	{"A1-054", "eat", "boy", "apple", "", "", false},
	{"A1-055", "drink", "girl", "water", "", "", true},
	{"A1-056", "fly", "bird", "", "", "", false},
	{"A1-057", "open", "mother", "window", "", "", false},
	{"A1-058", "close", "father", "window", "", "", false},
	{"A1-059", "have", "person", "car", "first", "singular", false},
	{"A1-060", "have", "person", "water", "first", "plural", true},
}

func authoredM3Meaning(spec m3Spec, lang string) sem.Meaning {
	actor := sem.Node{ID: "source-independent.actor", Kind: sem.Referent, Concept: core.ID("entity." + spec.subject)}
	if spec.person != "" {
		actor.Person, actor.Number = spec.person, spec.number
	} else {
		actor.Number = "singular"
		if lang == "en" {
			actor.Definiteness = "indefinite"
		}
	}
	agentRole, objectRole := "agent", "patient"
	if spec.predicate == "have" {
		agentRole, objectRole = "holder", "theme"
	}
	event := sem.Node{ID: "source-independent.event", Kind: sem.Event, Concept: core.ID("predicate." + spec.predicate), Tense: "present", Arguments: []sem.Argument{{Role: agentRole, Target: actor.ID}}}
	scope := sem.Node{ID: "source-independent.scope", Kind: sem.Scope, Operator: sem.Not, Content: event.ID}
	nodes := []sem.Node{actor, scope, event}
	if spec.object != "" {
		object := sem.Node{ID: "source-independent.object", Kind: sem.Referent, Concept: core.ID("entity." + spec.object)}
		if !spec.mass {
			object.Number = "singular"
		}
		if lang == "en" && !spec.mass {
			object.Definiteness = "indefinite"
		}
		event.Arguments = append(event.Arguments, sem.Argument{Role: objectRole, Target: object.ID})
		nodes[2] = event
		nodes = append(nodes, object)
	}
	return sem.Meaning{Root: scope.ID, Nodes: nodes}
}

func TestM3IndependentMeaningRequirements(t *testing.T) {
	c, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	if len(m3Requirements) != 10 {
		t.Fatalf("expected 10 negative requirements, got %d", len(m3Requirements))
	}
	for i, s := range m3Requirements {
		if s.id != cases[50+i].ID {
			t.Fatalf("unexpected record %s", s.id)
		}
		for _, lang := range []string{"pl", "en"} {
			if err := sem.Validate(authoredM3Meaning(s, lang), c); err != nil {
				t.Errorf("%s %s: %v", s.id, lang, err)
			}
		}
	}
}
