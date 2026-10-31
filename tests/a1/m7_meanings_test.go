package a1_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/sem"
)

// Semantic requirements are independent of the implementation of source
// analysis. English article evidence is a property of each referring occurrence.
type m7Spec struct {
	predicate, subject, object, tense string
	weather, firstPerson              bool
}

var m7Requirements = []m7Spec{
	{predicate: "sleep", subject: "cat", tense: "past"},
	{predicate: "run", subject: "dog", tense: "past"},
	{predicate: "read", subject: "teacher", object: "book", tense: "past"},
	{predicate: "cook", subject: "mother", tense: "past"},
	{predicate: "rain", subject: "rain", tense: "past", weather: true},
	{predicate: "sleep", subject: "cat", tense: "future"},
	{predicate: "read", subject: "teacher", object: "book", tense: "future"},
	{predicate: "run", subject: "child", tense: "future"},
	{predicate: "cook", subject: "mother", tense: "future"},
	{predicate: "work", subject: "person", tense: "future", firstPerson: true},
}

func authoredM7Meaning(spec m7Spec, lang string) sem.Meaning {
	sub := sem.Node{ID: "authored.actor", Kind: sem.Referent, Concept: core.ID("entity." + spec.subject)}
	if spec.firstPerson {
		sub.Person = "first"
		sub.Number = "singular"
	} else if !spec.weather {
		sub.Number = "singular"
		if lang == "en" {
			sub.Definiteness = "indefinite"
		}
	}
	role := "agent"
	if spec.weather {
		role = "theme"
	}
	ref := "previous_day"
	if spec.tense == "future" {
		ref = "next_day"
	}
	event := sem.Node{ID: "authored.predicate", Kind: sem.Event, Concept: core.ID("predicate." + spec.predicate), Tense: spec.tense, TimeReference: ref, Arguments: []sem.Argument{{Role: role, Target: sub.ID}}}
	nodes := []sem.Node{sub, event}
	if spec.object != "" {
		obj := sem.Node{ID: "authored.patient", Kind: sem.Referent, Concept: core.ID("entity." + spec.object), Number: "singular"}
		if lang == "en" {
			obj.Definiteness = "indefinite"
		}
		event.Arguments = append(event.Arguments, sem.Argument{Role: "patient", Target: obj.ID})
		nodes[1] = event
		nodes = append(nodes, obj)
	}
	return sem.Meaning{Root: event.ID, Nodes: nodes}
}
func TestM7IndependentTemporalMeaningRequirements(t *testing.T) {
	c, _ := core.BaseA1()
	if len(m7Requirements) != 10 {
		t.Fatal("invalid temporal case count")
	}
	for i, spec := range m7Requirements {
		for _, lang := range []string{"pl", "en"} {
			if err := sem.Validate(authoredM7Meaning(spec, lang), c); err != nil {
				t.Errorf("%s/%s: %v", cases[90+i].ID, lang, err)
			}
		}
	}
}
