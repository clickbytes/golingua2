package pl

import (
	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/lexicon"
)

func addQuestions(d compile.Definition) compile.Definition {
	d.Specials.QuestionWords = map[string]string{}
	for _, p := range []struct{ key, surface string }{{"polar", "czy"}, {"who", "kto"}, {"what", "co"}, {"where", "gdzie"}, {"how_many", "ile"}} {
		id := "pl.question." + p.key
		d.Specials.QuestionWords[p.key] = id
		d.Lexemes = append(d.Lexemes, lexicon.Entry{ID: id, Language: "pl", POS: lexicon.Particle, ClassID: "pl.particle.invariant", Lemma: p.surface})
	}
	return d
}
