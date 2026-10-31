package en

import (
	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/lexicon"
)

// Interrogative markers are lexicalized in the language pack. Grammar/engine
// are independent of their actual English spelling.
func addQuestions(d compile.Definition) compile.Definition {
	d.Specials.QuestionWords = map[string]string{}
	for _, p := range []struct{ key, surface string }{{"polar", "whether"}, {"who", "who"}, {"what", "what"}, {"where", "where"}, {"how_many", "how many"}} {
		id := "en.question." + p.key
		d.Specials.QuestionWords[p.key] = id
		d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: id, Language: "en", POS: lexicon.Particle, ClassID: "en.particle.invariant", Lemma: p.surface})
	}
	return d
}
