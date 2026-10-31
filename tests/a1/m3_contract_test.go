package a1_test

import (
	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/sem"
	"github.com/clickbytes/golingua2/internal/status"
	"testing"
)

func TestM3SixIndependentFlows(t *testing.T) {
	e, c := newEngine(t)
	for i, tc := range cases[50:60] {
		s := m3Requirements[i]
		if s.id != tc.ID {
			t.Fatalf("mismatch %s", tc.ID)
		}
		req := m2Request(tc.Context)
		pl := authoredM3Meaning(s, "pl")
		en := authoredM3Meaning(s, "en")
		t.Run(tc.ID, func(t *testing.T) {
			t.Run("PL_to_IR", func(t *testing.T) { m2SingleMeaning(t, e.Analyze("pl", tc.PL, req), pl, c) })
			t.Run("EN_to_IR", func(t *testing.T) { m2SingleMeaning(t, e.Analyze("en", tc.EN, req), en, c) })
			t.Run("IR_to_PL", func(t *testing.T) { m2Contains(t, e.Realize("pl", pl), tc.PL, pl, c) })
			t.Run("IR_to_EN", func(t *testing.T) { m2Contains(t, e.Realize("en", en), tc.EN, en, c) })
			t.Run("PL_to_EN", func(t *testing.T) { m2Contains(t, e.Translate("pl", "en", tc.PL, req), tc.EN, pl, c) })
			t.Run("EN_to_PL", func(t *testing.T) { m2Contains(t, e.Translate("en", "pl", tc.EN, req), tc.PL, en, c) })
		})
	}
}

func TestM3NegationContrasts(t *testing.T) {
	e, c := newEngine(t)
	bad := []struct{ lang, text string }{
		{"en", "A cat do not sleep."}, {"en", "A cat does not sleeps."},
		{"en", "A dog not run."}, {"en", "A bird does run."},
		{"en", "A teacher does not read book."}, {"en", "A boy does not eats an apple."},
		{"en", "I does not have a car."}, {"en", "We does not have water."},
		{"pl", "Kot nie śpi psa."}, {"pl", "Nauczyciel nie czyta książkę."},
		{"pl", "Chłopiec nie je jabłko."}, {"pl", "Dziewczynka nie pije wodę."},
		{"pl", "Mama nie otwiera okno."}, {"pl", "Nie ma samochodu."},
		{"pl", "Nie nie mam samochodu."}, {"pl", "Nauczyciel nie czyta książka."},
	}
	for _, tc := range bad {
		t.Run(tc.lang+":"+tc.text, func(t *testing.T) {
			got := e.Analyze(tc.lang, tc.text, context.Request{})
			if len(got.Candidates) != 0 || got.Outcome == status.OK || got.Outcome == status.Ambiguous {
				t.Fatalf("illegal reading accepted: %+v", got)
			}
		})
	}
	positive := e.Analyze("pl", "Kot śpi.", context.Request{})
	negated := e.Analyze("pl", "Kot nie śpi.", context.Request{})
	if len(positive.Candidates) != 1 || len(negated.Candidates) != 1 {
		t.Fatalf("contrast unavailable: %+v %+v", positive, negated)
	}
	same, err := sem.Equivalent(positive.Candidates[0].Meaning, negated.Candidates[0].Meaning, c)
	if err != nil || same {
		t.Fatalf("negation collapsed into affirmative: %v", err)
	}
	for _, s := range []m3Spec{m3Requirements[2], m3Requirements[8]} {
		meaning := authoredM3Meaning(s, "pl")
		for i := range meaning.Nodes {
			if meaning.Nodes[i].Kind == sem.Event {
				meaning.Nodes[i].Arguments = meaning.Nodes[i].Arguments[:1]
				break
			}
		}
		if sem.Validate(meaning, c) == nil {
			t.Fatal("missing argument accepted")
		}
		res := e.Realize("en", meaning)
		if res.Outcome != status.Invalid || len(res.Candidates) > 0 {
			t.Fatalf("partial clause emitted: %+v", res)
		}
	}
	input := authoredM3Meaning(m3Requirements[2], "pl")
	scopeIndex := -1
	for i, n := range input.Nodes {
		if n.Kind == sem.Scope {
			scopeIndex = i
		}
	}
	input.Nodes[scopeIndex].Content = "source-independent.object"
	if sem.Validate(input, c) == nil {
		t.Fatal("negative scope over entity accepted")
	}
}
