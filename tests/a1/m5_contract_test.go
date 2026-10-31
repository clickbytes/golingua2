package a1_test

import (
	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/sem"
	"github.com/clickbytes/golingua2/internal/status"
	"testing"
)

func TestM5SixIndependentFlows(t *testing.T) {
	e, c := newEngine(t)
	for i, tc := range cases[80:90] {
		s := m5Requirements[i]
		pl := authoredM5Meaning(s, "pl")
		en := authoredM5Meaning(s, "en")
		t.Run(tc.ID, func(t *testing.T) {
			req := context.Request{}
			t.Run("PL_to_IR", func(t *testing.T) { m2SingleMeaning(t, e.Analyze("pl", tc.PL, req), pl, c) })
			t.Run("EN_to_IR", func(t *testing.T) { m2SingleMeaning(t, e.Analyze("en", tc.EN, req), en, c) })
			t.Run("IR_to_PL", func(t *testing.T) { m2Contains(t, e.Realize("pl", pl), tc.PL, pl, c) })
			t.Run("IR_to_EN", func(t *testing.T) { m2Contains(t, e.Realize("en", en), tc.EN, en, c) })
			t.Run("PL_to_EN", func(t *testing.T) { m2Contains(t, e.Translate("pl", "en", tc.PL, req), tc.EN, pl, c) })
			t.Run("EN_to_PL", func(t *testing.T) { m2Contains(t, e.Translate("en", "pl", tc.EN, req), tc.PL, en, c) })
		})
	}
}
func TestM5NumberIsNotAccidentalPluralOrExistence(t *testing.T) {
	e, c := newEngine(t)
	a := authoredM5Meaning(m5Requirements[0], "pl")
	b := a.Clone()
	for i := range b.Nodes {
		if b.Nodes[i].Kind == sem.Referent && b.Nodes[i].Cardinality == 2 {
			b.Nodes[i].Cardinality = 3
		}
	}
	eq, err := sem.Equivalent(a, b, c)
	if err != nil || eq {
		t.Fatal("cardinality distinctions lost")
	}
	for _, tc := range []struct{ language, text string }{
		{"en", "I have two book."}, {"en", "I have two a books."}, {"en", "I have five water."},
		{"pl", "Mam dwa książki."}, {"pl", "Mam pięć książki."}, {"pl", "Mam trzy jabłko."},
		{"en", "There are two cat in the house."}, {"en", "There is two cats in the house."},
		{"pl", "W domu jest dwa koty."}, {"pl", "Na stole są trzy jabłko."},
	} {
		r := e.Analyze(tc.language, tc.text, context.Request{})
		if r.Outcome == status.OK || r.Outcome == status.Ambiguous || len(r.Candidates) > 0 {
			t.Errorf("illegal count syntax %q accepted: %+v", tc.text, r)
		}
	}
}
