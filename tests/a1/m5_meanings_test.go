package a1_test

import (
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/sem"
	"testing"
)

type m5Spec struct {
	id                            string
	predicate                     string
	person, personNumber, gender  string
	subjectConcept, objectConcept string
	quantity                      int
	relation, site                string
}

var m5Requirements = []m5Spec{
	{id: "A1-081", predicate: "have", person: "first", personNumber: "singular", objectConcept: "book", quantity: 2},
	{id: "A1-082", predicate: "have", person: "first", personNumber: "singular", objectConcept: "apple", quantity: 3},
	{id: "A1-083", predicate: "see", person: "first", personNumber: "singular", objectConcept: "dog", quantity: 2},
	{id: "A1-084", relation: "on_support", site: "table", objectConcept: "apple", quantity: 4},
	{id: "A1-085", relation: "inside", site: "house", objectConcept: "cat", quantity: 2},
	{id: "A1-086", relation: "inside", site: "room", objectConcept: "window", quantity: 1},
	{id: "A1-087", predicate: "have", person: "first", personNumber: "singular", objectConcept: "book", quantity: 5},
	{id: "A1-088", predicate: "have", person: "third", personNumber: "singular", gender: "feminine", objectConcept: "bicycle", quantity: 2},
	{id: "A1-089", predicate: "have", person: "first", personNumber: "plural", objectConcept: "chair", quantity: 3},
	{id: "A1-090", relation: "on_support", site: "table", objectConcept: "mug", quantity: 2},
}

func authoredM5Meaning(s m5Spec, lang string) sem.Meaning {
	number := "plural"
	if s.quantity == 1 {
		number = "singular"
	}
	object := sem.Node{ID: "auth.m5.object", Kind: sem.Referent, Concept: core.ID("entity." + s.objectConcept), Number: number, Cardinality: s.quantity}
	if s.relation != "" {
		site := sem.Node{ID: "auth.m5.site", Kind: sem.Referent, Concept: core.ID("entity." + s.site), Number: "singular"}
		if lang == "en" {
			site.Definiteness = "definite"
		}
		rel := sem.Node{ID: "auth.m5.existential", Kind: sem.State, Concept: core.ID("relation." + s.relation), Existential: true, Arguments: []sem.Argument{{Role: "located", Target: object.ID}, {Role: "site", Target: site.ID}}}
		return sem.Meaning{Root: rel.ID, Nodes: []sem.Node{site, rel, object}}
	}
	sub := sem.Node{ID: "auth.m5.subject", Kind: sem.Referent, Concept: "entity.person", Person: s.person, Number: s.personNumber, PronounGender: s.gender}
	role := "agent"
	objRole := "patient"
	if s.predicate == "have" {
		role = "holder"
		objRole = "theme"
	}
	event := sem.Node{ID: "auth.m5.event", Kind: sem.Event, Concept: core.ID("predicate." + s.predicate), Tense: "present", Arguments: []sem.Argument{{Role: role, Target: sub.ID}, {Role: objRole, Target: object.ID}}}
	return sem.Meaning{Root: event.ID, Nodes: []sem.Node{object, sub, event}}
}
func TestM5IndependentMeaningRequirements(t *testing.T) {
	c, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	if len(m5Requirements) != 10 {
		t.Fatalf("wrong M5 requirements")
	}
	for i, s := range m5Requirements {
		if s.id != cases[80+i].ID {
			t.Fatal(s.id)
		}
		for _, lang := range []string{"pl", "en"} {
			if err := sem.Validate(authoredM5Meaning(s, lang), c); err != nil {
				t.Errorf("%s/%s: %v", s.id, lang, err)
			}
		}
	}
}
