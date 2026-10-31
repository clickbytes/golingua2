package a1_test

import (
	"testing"

	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/engine"
	"github.com/clickbytes/golingua2/internal/status"
	"github.com/clickbytes/golingua2/languages/en"
	"github.com/clickbytes/golingua2/languages/pl"
)

// The fixed 100 Go cases are the work profile, not a lookup for translations.
// Compilation is excluded so this measures the warmed semantic pipeline.
func BenchmarkA1TranslationRoundTrip(b *testing.B) {
	catalog, err := core.BaseA1()
	if err != nil {
		b.Fatal(err)
	}
	service, err := engine.New(catalog, pl.Definition(), en.Definition())
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc := cases[i%len(cases)]
		ctx := m2Request(tc.Context)
		if tc.Context&NearDeixis != 0 {
			ctx.Deixis = context.Proximal
		}
		src, dst, text := "pl", "en", tc.PL
		if i%2 == 1 {
			src, dst, text = "en", "pl", tc.EN
		}
		result := service.Translate(src, dst, text, ctx)
		if result.Outcome != status.OK && result.Outcome != status.Ambiguous {
			b.Fatalf("%s: %+v", tc.ID, result)
		}
	}
}
