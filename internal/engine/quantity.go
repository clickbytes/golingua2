package engine

import (
	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/sem"
)

// Cardinality is a referent property, not a noun lexeme or a sentence template.
// Numeral morphology is selected in the target language's own package.
func numberedNounCase(language, required string, cardinality int) string {
	if language == "pl" && cardinality >= 5 {
		return "genitive"
	}
	if language == "pl" && cardinality >= 2 && required == "accusative" {
		return "nominative"
	}
	return required
}
func numeralMatches(n lexicon.Entry, noun lexicon.Entry, cardinality int) bool {
	return n.POS == lexicon.Numeral && n.NumeralValue == cardinality && (n.Gender == "" || n.Gender == noun.Gender)
}
func (e *Engine) parseNumberedNP(p *compile.Pack, tokens []string, requiredCase string) []nounReading {
	if len(tokens) != 2 {
		return nil
	}
	out := []nounReading{}
	for _, numeralForm := range p.Morph.Analyze(tokens[0]) {
		numeral, ok := p.Lexicon.Lookup(numeralForm.LexemeID)
		if !ok || numeral.POS != lexicon.Numeral {
			continue
		}
		number := "plural"
		if numeral.NumeralValue == 1 {
			number = "singular"
		}
		wantCase := numberedNounCase(p.Language, requiredCase, numeral.NumeralValue)
		for _, form := range p.Morph.Analyze(tokens[1]) {
			noun, ok := p.Lexicon.Lookup(form.LexemeID)
			if !ok || noun.POS != lexicon.Noun || noun.Countability == "mass" || noun.Countability == "plural_only" ||
				form.Features.Number != number || form.Features.Case != wantCase || !numeralMatches(numeral, noun, numeral.NumeralValue) {
				continue
			}
			for _, sense := range p.Lexicon.SensesFor(noun.ID) {
				ref := sem.Node{Kind: sem.Referent, Concept: sense.Concept, Number: number, Cardinality: numeral.NumeralValue}
				out = append(out, nounReading{ref: ref, lex: noun, sense: sense, tokens: 2})
			}
		}
	}
	return out
}
func (e *Engine) renderNumberedNP(p *compile.Pack, ref sem.Node, grammaticalCase string) []phraseOption {
	if ref.Cardinality == 0 || ref.Person != "" || ref.Definiteness != "" {
		return nil
	}
	var out []phraseOption
	for _, noun := range selectLexemes(p, ref.Concept, lexicon.Noun) {
		if noun.Countability == "mass" || noun.Countability == "plural_only" {
			continue
		}
		nounSurface, err := p.Morph.Generate(noun.ID, features.Spec{Number: ref.Number, Case: numberedNounCase(p.Language, grammaticalCase, ref.Cardinality)})
		if err != nil {
			continue
		}
		for _, n := range p.Lexicon.Entries() {
			if !numeralMatches(n, noun, ref.Cardinality) {
				continue
			}
			numeral, err := p.Morph.Generate(n.ID, features.Spec{})
			if err == nil {
				out = append(out, phraseOption{text: numeral + " " + nounSurface, grammaticalNumber: ref.Number})
			}
		}
	}
	return out
}
