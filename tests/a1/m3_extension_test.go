package a1_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/engine"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/sem"
	"github.com/clickbytes/golingua2/internal/status"
	"github.com/clickbytes/golingua2/languages/en"
	"github.com/clickbytes/golingua2/languages/pl"
)

// New predicate+lexeme declarations exercise the SAME generic negative
// construction as the 10 A1 examples. No new sentence grammar is installed.
func TestM3NegativeNewPredicateFromDataOnly(t *testing.T) {
	base, err := core.BaseA1()
	if err != nil {
		t.Fatal(err)
	}
	definitions := append(base.Definitions(), core.Definition{ID: "predicate.observe", Kind: core.Predicate, Roles: []core.Role{{Name: "agent", Target: core.Entity}, {Name: "patient", Target: core.Entity}}})
	catalog, err := core.NewCatalog(definitions)
	if err != nil {
		t.Fatal(err)
	}
	polish, english := pl.Definition(), en.Definition()
	polish.Lexemes = append(polish.Lexemes, lexicon.Entry{ID: "pl.verb.observe", Language: "pl", POS: lexicon.Verb, Lemma: "obserwować", ClassID: "pl.verb.present-e", Stems: map[string]string{"root": "obserwuj"}})
	polish.Senses = append(polish.Senses, lexicon.Sense{ID: "pl.sense.verb.observe", LexemeID: "pl.verb.observe", Concept: "predicate.observe", FrameIDs: []string{"pl.frame.acc"}})
	english.Lexemes = append(english.Lexemes, lexicon.Entry{ID: "en.verb.observe", Language: "en", POS: lexicon.Verb, Lemma: "observe", ClassID: "en.verb.present-s"})
	english.Senses = append(english.Senses, lexicon.Sense{ID: "en.sense.verb.observe", LexemeID: "en.verb.observe", Concept: "predicate.observe", FrameIDs: []string{"en.frame.acc"}})
	e, err := engine.New(catalog, polish, english)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ source, target, input, want string }{
		{"pl", "en", "Nauczyciel nie obserwuje psa.", "A teacher does not observe a dog."},
		{"en", "pl", "A teacher does not observe a dog.", "Nauczyciel nie obserwuje psa."},
	} {
		t.Run(tc.source+"-"+tc.target, func(t *testing.T) {
			r := e.Translate(tc.source, tc.target, tc.input, context.Request{})
			if r.Outcome != status.OK && r.Outcome != status.Ambiguous {
				t.Fatalf("missing negative clause: %+v", r)
			}
			found := false
			for _, candidate := range r.Candidates {
				if candidate.Text == tc.want {
					found = true
				}
				root := candidate.Meaning.Root
				var scope sem.Node
				for _, n := range candidate.Meaning.Nodes {
					if n.ID == root {
						scope = n
						break
					}
				}
				if scope.Kind != sem.Scope || scope.Operator != sem.Not {
					t.Fatalf("lost Not scope: %+v", candidate)
				}
			}
			if !found {
				t.Fatalf("missing %q: %+v", tc.want, r)
			}
		})
	}
}

func TestM3NeutralYouRetainsTwoReadingsAndExplicitContext(t *testing.T) {
	e, _ := newEngine(t)
	ambiguous := e.Analyze("en", "You do not have a car.", context.Request{})
	if ambiguous.Outcome != status.Ambiguous || len(ambiguous.Candidates) != 2 {
		t.Fatalf("English you guessed singular/plural: %+v", ambiguous)
	}
	for _, choice := range ambiguous.Candidates {
		if len(choice.Assumptions) != 1 || choice.Assumptions[0].Field != "addressee_number" {
			t.Fatalf("implicit addressee choice: %+v", choice)
		}
	}
	for _, test := range []struct{ number, want string }{{"singular", "Nie masz samochodu."}, {"plural", "Nie macie samochodu."}} {
		r := e.Translate("en", "pl", "You do not have a car.", context.Request{AddresseeNumber: test.number})
		if r.Outcome != status.OK || len(r.Candidates) != 1 || r.Candidates[0].Text != test.want {
			t.Fatalf("wrong addressee number %s: %+v", test.number, r)
		}
	}
}

func TestM3NegationIsLocalAndNotSilent(t *testing.T) {
	e, c := newEngine(t)
	positive := e.Analyze("en", "A dog runs.", context.Request{})
	negative := e.Analyze("en", "A dog does not run.", context.Request{})
	if len(positive.Candidates) != 1 || len(negative.Candidates) != 1 {
		t.Fatalf("missing positive or negative: %+v %+v", positive, negative)
	}
	eq, err := sem.Equivalent(positive.Candidates[0].Meaning, negative.Candidates[0].Meaning, c)
	if err != nil || eq {
		t.Fatalf("negation lost its semantic scope: %v", err)
	}
	target := e.Realize("pl", negative.Candidates[0].Meaning)
	if len(target.Candidates) == 0 || target.Candidates[0].Text != "Pies nie biega." {
		t.Fatalf("Polish target lost event Not: %+v", target)
	}
	for _, bad := range []struct{ language, text string }{{"pl", "Nauczyciel nie czyta książkę."}, {"en", "A dog does not runs."}} {
		r := e.Analyze(bad.language, bad.text, context.Request{})
		if r.Outcome == status.OK || r.Outcome == status.Ambiguous || len(r.Candidates) != 0 {
			t.Fatalf("invalid negative accepted: %+v", r)
		}
	}
}
