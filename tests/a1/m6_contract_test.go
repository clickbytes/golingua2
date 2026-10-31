package a1_test

import (
	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/sem"
	"github.com/clickbytes/golingua2/internal/status"
	"testing"
)

func TestM6SixIndependentFlows(t *testing.T) {
	e, c := newEngine(t)
	for i, tc := range cases[60:70] {
		req := m2Request(tc.Context)
		if tc.Context&NearDeixis != 0 {
			req.Deixis = context.Proximal
		}
		pl, en := authoredM6Meaning(m6Requirements[i], "pl"), authoredM6Meaning(m6Requirements[i], "en")
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
func TestM6QuestionScopeAndInvalidSurfaces(t *testing.T) {
	e, c := newEngine(t)
	normal := authoredM6Meaning(m6Requirements[6], "pl")
	different := authoredM6Meaning(m6Requirements[7], "pl")
	eq, err := sem.Equivalent(normal, different, c)
	if err != nil || eq {
		t.Fatal("who and what question identity conflated")
	}
	notQuestion := normal.Clone()
	for i := range notQuestion.Nodes {
		if notQuestion.Nodes[i].Kind == sem.Scope {
			notQuestion.Nodes[i].QuestionRole = "wrong"
		}
	}
	if sem.Validate(notQuestion, c) == nil {
		t.Fatal("question with unbound role accepted")
	}
	if r := e.Realize("en", notQuestion); r.Outcome != status.Invalid || len(r.Candidates) != 0 {
		t.Fatal("invalid query realized")
	}
	for _, bad := range []struct{ language, text string }{
		{"en", "Does this dog runs?"}, {"en", "Do this dog run?"}, {"en", "Does teacher read a book?"},
		{"pl", "Czy ten pies biegają?"}, {"en", "Who read a book?"}, {"pl", "Ile masz książka?"},
		{"en", "How many books does you have?"}, {"pl", "Co czyta kota?"},
	} {
		r := e.Analyze(bad.language, bad.text, context.Request{Deixis: context.Proximal, AddresseeNumber: "singular"})
		if r.Outcome == status.OK || r.Outcome == status.Ambiguous || len(r.Candidates) != 0 {
			t.Errorf("unlicensed question accepted %s: %+v", bad.text, r)
		}
	}
}
