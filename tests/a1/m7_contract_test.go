package a1_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/sem"
	"github.com/clickbytes/golingua2/internal/status"
)

func TestM7SixIndependentFlows(t *testing.T) {
	e, c := newEngine(t)
	for i, tc := range cases[90:100] {
		req := m2Request(tc.Context)
		pl := authoredM7Meaning(m7Requirements[i], "pl")
		en := authoredM7Meaning(m7Requirements[i], "en")
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
func TestM7TemporalScopeAndAgreementRefusals(t *testing.T) {
	e, c := newEngine(t)
	past := authoredM7Meaning(m7Requirements[0], "pl")
	future := past.Clone()
	for i := range future.Nodes {
		if future.Nodes[i].Kind == sem.Event {
			future.Nodes[i].Tense = "future"
			future.Nodes[i].TimeReference = "next_day"
		}
	}
	equal, err := sem.Equivalent(past, future, c)
	if err != nil || equal {
		t.Fatal("past and future collapsed")
	}
	orphan := past.Clone()
	for i := range orphan.Nodes {
		if orphan.Nodes[i].Kind == sem.Event {
			orphan.Nodes[i].TimeReference = "next_day"
		}
	}
	if sem.Validate(orphan, c) == nil {
		t.Fatal("inconsistent time reference was accepted")
	}
	if r := e.Realize("pl", orphan); r.Outcome != status.Invalid || len(r.Candidates) != 0 {
		t.Fatalf("invalid temporal IR realized: %+v", r)
	}
	for _, bad := range []struct{ lang, text string }{
		{"en", "Yesterday a dog run."}, {"en", "Tomorrow a cat sleeps."}, {"en", "Yesterday a teacher reads a book."},
		{"pl", "Wczoraj kot spała."}, {"pl", "Wczoraj mama gotował."}, {"pl", "Jutro kot będę spał."},
		{"pl", "Jutro będę pracował."}, {"en", "It rains yesterday."}, {"en", "It rained tomorrow."},
		{"en", "Yesterday a child will run."}, {"pl", "Jutro dziecko będzie biegała."},
	} {
		r := e.Analyze(bad.lang, bad.text, context.Request{})
		if r.Outcome == status.OK || r.Outcome == status.Ambiguous || len(r.Candidates) > 0 {
			t.Errorf("illegal temporal form accepted: %s -> %+v", bad.text, r)
		}
	}
}
func TestM7ComposesUnseenTemporalClauses(t *testing.T) {
	e, _ := newEngine(t)
	for _, test := range []struct{ pl, en string }{
		{"Wczoraj pies spał.", "Yesterday a dog slept."},
		{"Jutro pies będzie biegał.", "Tomorrow a dog will run."},
		{"Jutro mama będzie czytała książkę.", "Tomorrow a mother will read a book."},
		{"Wczoraj mama czytała książkę.", "Yesterday a mother read a book."},
	} {
		r := e.Translate("pl", "en", test.pl, context.Request{})
		found := false
		for _, candidate := range r.Candidates {
			if candidate.Text == test.en {
				found = true
			}
		}
		if !found {
			t.Errorf("missing procedural time realization %q → %q: %+v", test.pl, test.en, r)
		}
	}
}
