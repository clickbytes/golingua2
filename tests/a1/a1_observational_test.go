package a1_test

import (
	"github.com/clickbytes/golingua2/internal/context"
	"testing"

	"github.com/clickbytes/golingua2/internal/status"
)

// The catalog remains exactly 100 pairs, even though only 100 currently have
// full six-flow acceptance. Observing the other 30 is NOT a passing assertion.
func TestA1ObservationalTranslationCoverage(t *testing.T) {
	e, _ := newEngine(t)
	accepted := map[string]int{"pl-en": 0, "en-pl": 0}
	for _, tc := range cases {
		ctx := m2Request(tc.Context)
		if tc.Context&NearDeixis != 0 {
			ctx.Deixis = context.Proximal
		}
		for _, direction := range []struct{ source, target, text, key string }{{"pl", "en", tc.PL, "pl-en"}, {"en", "pl", tc.EN, "en-pl"}} {
			r := e.Translate(direction.source, direction.target, direction.text, ctx)
			if r.Outcome == status.OK || r.Outcome == status.Ambiguous {
				accepted[direction.key]++
			}
		}
	}
	t.Logf("OBSERVED ONLY: PL→EN %d/100, EN→PL %d/100; 100 M1–M7 cases have six independent semantic contracts", accepted["pl-en"], accepted["en-pl"])
}
