package a1_test

import (
	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/sem"
	"testing"
)

type m6Spec struct {
	kind, predicate, subject, object, person, number, property, role string
	prox, subjectDef, objDef, mass                                   bool
}

var m6Requirements = []m6Spec{
	{kind: "polar", predicate: "classification", object: "cat", prox: true},
	{kind: "polar", predicate: "run", subject: "dog", prox: true},
	{kind: "polar", predicate: "have", subject: "person", person: "second", number: "singular", object: "dog", objDef: false},
	{kind: "polar", predicate: "read", subject: "teacher", object: "book", subjectDef: true},
	{kind: "polar", predicate: "property", subject: "book", property: "new", prox: true},
	{kind: "polar", predicate: "have", subject: "person", person: "second", number: "singular", object: "water", mass: true},
	{kind: "who", predicate: "read", subject: "unspecified", object: "book", role: "agent"},
	{kind: "what", predicate: "read", subject: "teacher", object: "unspecified", subjectDef: true, role: "patient"},
	{kind: "where", predicate: "location", subject: "cat", object: "unspecified", subjectDef: true, role: "site"},
	{kind: "how_many", predicate: "have", subject: "person", person: "second", number: "singular", object: "book", role: "theme"},
}

func authoredM6Meaning(s m6Spec, lang string) sem.Meaning {
	var event sem.Node
	var refs []sem.Node
	if s.predicate == "classification" {
		m := authoredM1Meaning("classification", s.object, "")
		event = m.Nodes[0]
		refs = append(refs, m.Nodes[1])
	} else if s.predicate == "property" {
		m := authoredM1Meaning("property", s.subject, s.property)
		event = m.Nodes[0]
		refs = append(refs, m.Nodes[1])
	} else {
		sub := sem.Node{ID: "qtest.subject", Kind: sem.Referent, Concept: core.ID("entity." + s.subject)}
		if s.subject == "unspecified" {
			sub.Number = "singular"
		} else if s.person != "" {
			sub.Person = s.person
			sub.Number = s.number
		} else {
			sub.Number = "singular"
			if lang == "en" && s.subjectDef {
				sub.Definiteness = "definite"
			}
		}
		if s.prox {
			sub.Reference = context.Proximal
		}
		refs = append(refs, sub)
		if s.predicate == "location" {
			site := sem.Node{ID: "qtest.site", Kind: sem.Referent, Concept: "entity.unspecified"}
			refs = append(refs, site)
			event = sem.Node{ID: "qtest.place", Kind: sem.State, Concept: "relation.located", Arguments: []sem.Argument{{Role: "located", Target: sub.ID}, {Role: "site", Target: site.ID}}}
		} else {
			actorRole, objectRole := "agent", "patient"
			if s.predicate == "have" {
				actorRole, objectRole = "holder", "theme"
			}
			event = sem.Node{ID: "qtest.event", Kind: sem.Event, Concept: core.ID("predicate." + s.predicate), Tense: "present", Arguments: []sem.Argument{{Role: actorRole, Target: sub.ID}}}
			if s.object != "" {
				obj := sem.Node{ID: "qtest.object", Kind: sem.Referent, Concept: core.ID("entity." + s.object)}
				if s.kind == "how_many" {
					obj.Number = "plural"
				} else if s.object != "unspecified" && !s.mass {
					obj.Number = "singular"
				}
				if lang == "en" && !s.mass && s.object != "unspecified" && s.kind != "how_many" {
					if s.objDef {
						obj.Definiteness = "definite"
					} else {
						obj.Definiteness = "indefinite"
					}
				}
				refs = append(refs, obj)
				event.Arguments = append(event.Arguments, sem.Argument{Role: objectRole, Target: obj.ID})
			}
		}
	}
	scope := sem.Node{ID: "qtest.scope", Kind: sem.Scope, Operator: sem.Ask, QuestionKind: s.kind, QuestionRole: s.role, Content: event.ID}
	nodes := append([]sem.Node{scope, event}, refs...)
	return sem.Meaning{Root: scope.ID, Nodes: nodes}
}
func TestM6IndependentQuestionMeanings(t *testing.T) {
	c, _ := core.BaseA1()
	if len(m6Requirements) != 10 {
		t.Fatal("unexpected question count")
	}
	for i, s := range m6Requirements {
		for _, lang := range []string{"pl", "en"} {
			if err := sem.Validate(authoredM6Meaning(s, lang), c); err != nil {
				t.Errorf("%s/%s: %v", cases[60+i].ID, lang, err)
			}
		}
	}
}
