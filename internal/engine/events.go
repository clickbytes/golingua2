package engine

import (
	"sort"
	"strings"

	"github.com/clickbytes/golingua2/internal/compile"
	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/grammar"
	"github.com/clickbytes/golingua2/internal/lexicon"
	"github.com/clickbytes/golingua2/internal/sem"
	"github.com/clickbytes/golingua2/internal/status"
)

// Phrase components are matched independently of a particular sentence or
// predicate name. Different occurrences never share grammatical evidence.
type nounReading struct {
	ref    sem.Node
	lex    lexicon.Entry
	sense  lexicon.Sense
	tokens int
}

type verbReading struct {
	lex    lexicon.Entry
	sense  lexicon.Sense
	number string
	person string
}

func licensedFrame(s lexicon.Sense, frameID string) bool {
	for _, f := range s.FrameIDs {
		if f == frameID {
			return true
		}
	}
	return false
}

func (e *Engine) parseNounPhrase(p *compile.Pack, tokens []string, slotCase string, req context.Request, subject bool) []nounReading {
	if len(tokens) == 0 || len(tokens) > 2 {
		return nil
	}
	out := []nounReading{}
	// Exactly one word for lexical pronouns; pronoun morphology is nominative.
	if subject && len(tokens) == 1 {
		for _, form := range p.Morph.Analyze(tokens[0]) {
			lex, ok := p.Lexicon.Lookup(form.LexemeID)
			if !ok || lex.POS != lexicon.Pronoun {
				continue
			}
			for _, sense := range p.Lexicon.SensesFor(lex.ID) {
				if sense.Concept != "entity.person" {
					continue
				}
				ref := sem.Node{Kind: sem.Referent, Concept: sense.Concept, Person: lex.PronounPerson, Number: lex.PronounNumber, PronounGender: lex.PronounGender}
				if ref.Person == "second" && ref.Number == "" {
					ref.Number = req.AddresseeNumber
					if ref.Number == "" {
						for _, number := range []string{"singular", "plural"} {
							choice := ref
							choice.Number = number
							out = append(out, nounReading{ref: choice, lex: lex, sense: sense, tokens: 1})
						}
						continue
					}
				}
				out = append(out, nounReading{ref: ref, lex: lex, sense: sense, tokens: 1})
			}
		}
	}
	word := tokens[len(tokens)-1]
	article := ""
	if len(tokens) == 2 {
		if p.Language != "en" {
			return out
		}
		article = tokens[0]
		if article != "a" && article != "an" && article != "the" {
			return out
		}
	}
	for _, form := range p.Morph.Analyze(word) {
		lex, ok := p.Lexicon.Lookup(form.LexemeID)
		if !ok || lex.POS != lexicon.Noun {
			continue
		}
		if form.Features.Case != slotCase {
			continue
		}
		countability := lex.Countability
		if countability == "" {
			countability = "count"
		}
		number := form.Features.Number
		if countability == "mass" || countability == "plural_only" {
			number = ""
		}
		def := ""
		if p.Language == "en" {
			switch article {
			case "a", "an":
				if countability != "count" || form.Features.Number != "singular" || (article == "a" && lex.Onset != "consonant") || (article == "an" && lex.Onset != "vowel") {
					continue
				}
				def = "indefinite"
			case "the":
				def = "definite"
			case "":
				if countability == "count" && form.Features.Number == "singular" {
					continue
				}
				if countability == "count" {
					def = "indefinite"
				}
			}
		} else if article != "" {
			continue
		}
		for _, sense := range p.Lexicon.SensesFor(lex.ID) {
			ref := sem.Node{Kind: sem.Referent, Concept: sense.Concept, Number: number, Definiteness: def}
			out = append(out, nounReading{ref: ref, lex: lex, sense: sense, tokens: len(tokens)})
		}
	}
	return out
}

func (e *Engine) parseEventVerb(p *compile.Pack, token string, frameID string, base bool) []verbReading {
	out := []verbReading{}
	for _, form := range p.Morph.Analyze(token) {
		lex, ok := p.Lexicon.Lookup(form.LexemeID)
		if !ok || lex.POS != lexicon.Verb || base && (form.Features.VerbForm != "base" || form.Features.Tense != "") ||
			!base && (form.Features.Tense != "present" || form.Features.VerbForm != "") {
			continue
		}
		for _, sense := range p.Lexicon.SensesFor(lex.ID) {
			if !licensedFrame(sense, frameID) {
				continue
			}
			out = append(out, verbReading{lex: lex, sense: sense, person: form.Features.Person, number: form.Features.Number})
		}
	}
	return out
}

// A negative auxiliary is an independent lexical and morphological license.
// A marker or auxiliary is never accepted merely by its surface spelling.
func markerLicensed(p *compile.Pack, token, lexemeID string) bool {
	for _, reading := range p.Morph.Analyze(token) {
		if reading.LexemeID == lexemeID {
			return true
		}
	}
	return false
}

func negativeAuxiliaries(p *compile.Pack, token string) []verbReading {
	var out []verbReading
	for _, reading := range p.Morph.Analyze(token) {
		if reading.LexemeID != p.Specials.NegationAuxiliary || reading.Features.Tense != "present" {
			continue
		}
		out = append(out, verbReading{person: reading.Features.Person, number: reading.Features.Number})
	}
	return out
}

func eventObjectCase(frame grammar.EventFrame, negated bool) string {
	if negated && frame.NegatedObjectCase != "" {
		return frame.NegatedObjectCase
	}
	return frame.ObjectCase
}

func agreesWithVerb(ref sem.Node, v verbReading) bool {
	if ref.Person != "" {
		if v.person != ref.Person {
			return false
		}
	} else if v.person != "third" {
		return false
	}
	// English "you" is semantically unresolved until the request states number;
	// any matching verb cell is permissible only if it does not invent number.
	return ref.Number == "" || v.number == ref.Number
}

func (e *Engine) analyzeEventPattern(p *compile.Pack, pat grammar.Pattern, tokens []string, req context.Request) []Candidate {
	frame, ok := p.EventFrame(pat.FrameID)
	if !ok {
		return nil
	}
	var result []Candidate
	hasSubject := false
	hasObject := frame.ObjectRole != ""
	for _, el := range pat.Sequence {
		if el.Slot == grammar.SubjectPhrase {
			hasSubject = true
		}
	}
	hasAux := false
	for _, el := range pat.Sequence {
		if el.Slot == grammar.NegationAuxiliary {
			hasAux = true
		}
	}
	// Every pattern has one finite VP (or a base VP licensed by do-support);
	// an NP consumes at most an article+noun.
	for subjectLen := 0; subjectLen <= 2; subjectLen++ {
		if hasSubject && subjectLen == 0 || !hasSubject && subjectLen != 0 {
			continue
		}
		verbIndex := subjectLen
		markerIndex := -1
		auxIndex := -1
		if hasAux {
			auxIndex = verbIndex
			verbIndex++
		}
		if pat.Negated {
			markerIndex = verbIndex
			verbIndex++
		}
		if verbIndex >= len(tokens) {
			continue
		}
		if pat.Negated && !markerLicensed(p, tokens[markerIndex], p.Specials.NegationMarker) {
			continue
		}
		objectTokens := tokens[verbIndex+1:]
		if hasObject && (len(objectTokens) < 1 || len(objectTokens) > 2) || !hasObject && len(objectTokens) != 0 {
			continue
		}
		// Unlike a sentence lookup, all choices are independently derived by token,
		// POS, morphology, frame license, and the same reusable phrase grammar.
		var subjects []nounReading
		if hasSubject {
			subjects = e.parseNounPhrase(p, tokens[:subjectLen], "nominative", req, true)
			if len(subjects) == 0 {
				continue
			}
		}
		objectCase := eventObjectCase(frame, pat.Negated)
		if p.Language == "en" {
			objectCase = "nominative"
		}
		var objects []nounReading
		if hasObject {
			objects = e.parseNounPhrase(p, objectTokens, objectCase, req, false)
			if len(objects) == 0 {
				continue
			}
		} else {
			objects = []nounReading{{}}
		}
		for _, verb := range e.parseEventVerb(p, tokens[verbIndex], pat.FrameID, hasAux) {
			if !hasSubject {
				if p.Language != "pl" || (verb.person != "first" && verb.person != "second") || verb.number == "" {
					continue
				}
				subjects = []nounReading{{ref: sem.Node{Kind: sem.Referent, Concept: "entity.person", Person: verb.person, Number: verb.number}}}
			}
			for _, sub := range subjects {
				if hasAux {
					licensed := false
					for _, aux := range negativeAuxiliaries(p, tokens[auxIndex]) {
						if agreesWithVerb(sub.ref, aux) {
							licensed = true
							break
						}
					}
					if !licensed {
						continue
					}
				} else if !agreesWithVerb(sub.ref, verb) {
					continue
				}
				for _, obj := range objects {
					pred := sem.Node{ID: "e1", Kind: sem.Event, Concept: verb.sense.Concept, Tense: "present", Arguments: []sem.Argument{{Role: frame.SubjectRole, Target: "r1"}}}
					subj := sub.ref
					subj.ID = "r1"
					nodes := []sem.Node{pred, subj}
					if hasObject {
						ob := obj.ref
						ob.ID = "r2"
						pred.Arguments = append(pred.Arguments, sem.Argument{Role: frame.ObjectRole, Target: ob.ID})
						nodes[0] = pred
						nodes = append(nodes, ob)
					}
					m := sem.Meaning{Root: pred.ID, Nodes: nodes}
					if pat.Negated {
						neg := sem.Node{ID: "n1", Kind: sem.Scope, Operator: sem.Not, Content: pred.ID}
						m.Root = neg.ID
						m.Nodes = append(m.Nodes, neg)
					}
					if err := sem.Validate(m, e.catalog); err != nil {
						continue
					}
					evidence := []Evidence{{Token: verbIndex, Lexeme: verb.lex.ID, Sense: verb.sense.ID, Node: pred.ID}}
					if pat.Negated {
						evidence = append(evidence, Evidence{Token: markerIndex, Lexeme: p.Specials.NegationMarker, Node: "n1"})
						if hasAux {
							evidence = append(evidence, Evidence{Token: auxIndex, Lexeme: p.Specials.NegationAuxiliary, Node: pred.ID})
						}
					}
					if hasSubject {
						evidence = append(evidence, Evidence{Token: 0, Lexeme: sub.lex.ID, Sense: sub.sense.ID, Node: "r1"})
					}
					if hasObject {
						evidence = append(evidence, Evidence{Token: verbIndex + 1, Lexeme: obj.lex.ID, Sense: obj.sense.ID, Node: "r2"})
					}
					candidate := Candidate{Meaning: m, Evidence: evidence}
					if hasSubject && sub.ref.Person == "second" && sub.lex.PronounNumber == "" && req.AddresseeNumber == "" {
						candidate.Assumptions = append(candidate.Assumptions, Annotation{Node: "r1", Field: "addressee_number", Value: sub.ref.Number})
					}
					result = append(result, candidate)
				}
			}
		}
	}
	return result
}

type phraseOption struct {
	text              string
	grammaticalNumber string
	assumptions       []Annotation
	losses            []Annotation
}

func addAssumption(node sem.NodeID, field, value string) []Annotation {
	return []Annotation{{Node: node, Field: field, Value: value}}
}

func (e *Engine) renderEventNP(p *compile.Pack, ref sem.Node, grammaticalCase string) []phraseOption {
	if ref.Person != "" {
		out := []phraseOption{}
		for _, lex := range selectLexemes(p, ref.Concept, lexicon.Pronoun) {
			if lex.PronounPerson != ref.Person {
				continue
			}
			if lex.PronounNumber != "" && ref.Number != "" && lex.PronounNumber != ref.Number {
				continue
			}
			if lex.PronounGender != "" && ref.PronounGender != "" && lex.PronounGender != ref.PronounGender {
				continue
			}
			if p.Language == "pl" && ref.Person == "third" && lex.PronounGender == "" {
				continue
			}
			text, err := p.Morph.Generate(lex.ID, features.Spec{})
			if err != nil {
				continue
			}
			option := phraseOption{text: text, grammaticalNumber: lex.PronounNumber}
			if option.grammaticalNumber == "" {
				option.grammaticalNumber = ref.Number
			}
			if ref.Number == "" && lex.PronounNumber != "" {
				option.assumptions = append(option.assumptions, Annotation{Node: ref.ID, Field: "number", Value: lex.PronounNumber})
			}
			if ref.PronounGender == "" && lex.PronounGender != "" {
				option.assumptions = append(option.assumptions, Annotation{Node: ref.ID, Field: "pronoun_gender", Value: lex.PronounGender})
			}
			out = append(out, option)
		}
		return out
	}
	out := []phraseOption{}
	for _, lex := range selectLexemes(p, ref.Concept, lexicon.Noun) {
		countability := lex.Countability
		if countability == "" {
			countability = "count"
		}
		numbers := []string{ref.Number}
		if countability == "mass" {
			if ref.Number == "plural" {
				continue
			}
			numbers = []string{"singular"}
		}
		if countability == "plural_only" {
			numbers = []string{"plural"}
		}
		if ref.Number == "" && countability == "count" {
			numbers = []string{"singular", "plural"}
		}
		for _, number := range numbers {
			suffix, err := p.Morph.Generate(lex.ID, features.Spec{Number: number, Case: grammaticalCase})
			if err != nil {
				continue
			}
			if p.Language == "pl" {
				option := phraseOption{text: suffix, grammaticalNumber: number}
				if ref.Definiteness != "" {
					option.losses = append(option.losses, Annotation{Node: ref.ID, Field: "determiner", Value: ref.Definiteness})
				}
				out = append(out, option)
				continue
			}
			// The English NP owns articles. Bare singular count nouns are forbidden.
			definiteModes := []string{ref.Definiteness}
			if ref.Definiteness == "" {
				definiteModes = []string{"indefinite", "definite"}
			}
			for _, mode := range definiteModes {
				option := phraseOption{text: suffix, grammaticalNumber: number}
				if mode == "definite" {
					article, err := p.Morph.Generate(p.Specials.DefiniteArticle, features.Spec{})
					if err != nil {
						continue
					}
					option.text = article + " " + suffix
				} else if mode == "indefinite" && countability == "count" && number == "singular" {
					if lex.Onset == "" {
						continue
					}
					article, err := p.Morph.Generate(p.Specials.Article, features.Spec{Onset: lex.Onset})
					if err != nil {
						continue
					}
					option.text = article + " " + suffix
				} else if mode != "indefinite" {
					continue
				}
				if ref.Definiteness == "" {
					option.assumptions = append(option.assumptions, Annotation{Node: ref.ID, Field: "determiner", Value: mode})
				}
				if ref.Number == "" && countability == "count" {
					option.assumptions = append(option.assumptions, Annotation{Node: ref.ID, Field: "number", Value: number})
				}
				out = append(out, option)
			}
		}
	}
	return out
}

func (e *Engine) realizeEvent(p *compile.Pack, m sem.Meaning, eventID sem.NodeID, negated bool) Result {
	nodes := map[sem.NodeID]sem.Node{}
	for _, n := range m.Nodes {
		nodes[n.ID] = n
	}
	event := nodes[eventID]
	produced := map[string]Candidate{}
	for _, pat := range p.Constructions {
		if !pat.Event || !pat.Realize || pat.Negated != negated {
			continue
		}
		frame, ok := p.EventFrame(pat.FrameID)
		if !ok {
			continue
		}
		args := map[string]sem.Node{}
		for _, a := range event.Arguments {
			args[a.Role] = nodes[a.Target]
		}
		subject, ok := args[frame.SubjectRole]
		if !ok {
			continue
		}
		object, hasObject := args[frame.ObjectRole]
		if (frame.ObjectRole != "") != hasObject {
			continue
		}
		if len(event.Arguments) != 1 && len(event.Arguments) != 2 {
			continue
		}
		hasSubject := false
		for _, el := range pat.Sequence {
			if el.Slot == grammar.SubjectPhrase {
				hasSubject = true
			}
		}
		if !hasSubject && (p.Language != "pl" || subject.Person != "first" && subject.Person != "second") {
			continue
		}
		if hasSubject && p.Language == "pl" && subject.Person != "" && subject.Person != "third" {
			continue
		}
		subjects := []phraseOption{{text: ""}}
		if hasSubject {
			subjects = e.renderEventNP(p, subject, "nominative")
		}
		objects := []phraseOption{{text: ""}}
		if hasObject {
			caseName := eventObjectCase(frame, negated)
			if p.Language == "en" {
				caseName = "nominative"
			}
			objects = e.renderEventNP(p, object, caseName)
		}
		person := subject.Person
		if person == "" {
			person = "third"
		}
		for _, verb := range selectLexemes(p, event.Concept, lexicon.Verb) {
			permitted := false
			for _, sense := range p.Lexicon.SensesFor(verb.ID) {
				if licensedFrame(sense, frame.ID) {
					permitted = true
					break
				}
			}
			if !permitted {
				continue
			}
			for _, sub := range subjects {
				number := subject.Number
				if sub.grammaticalNumber != "" {
					number = sub.grammaticalNumber
				}
				if number == "" {
					continue
				}
				hasAux := false
				for _, el := range pat.Sequence {
					if el.Slot == grammar.NegationAuxiliary {
						hasAux = true
					}
				}
				verbFeatures := features.Spec{Number: number, Person: person, Tense: event.Tense}
				if hasAux {
					verbFeatures = features.Spec{VerbForm: "base"}
				}
				surface, err := p.Morph.Generate(verb.ID, verbFeatures)
				if err != nil {
					continue
				}
				auxText := ""
				if hasAux {
					auxText, err = p.Morph.Generate(p.Specials.NegationAuxiliary, features.Spec{Number: number, Person: person, Tense: event.Tense})
					if err != nil {
						continue
					}
				}
				markerText := ""
				if pat.Negated {
					markerText, err = p.Morph.Generate(p.Specials.NegationMarker, features.Spec{})
					if err != nil {
						continue
					}
				}
				for _, obj := range objects {
					parts := make([]string, 0, len(pat.Sequence))
					for _, el := range pat.Sequence {
						switch el.Slot {
						case grammar.SubjectPhrase:
							parts = append(parts, sub.text)
						case grammar.PredicateVerb:
							parts = append(parts, surface)
						case grammar.NegationAuxiliary:
							parts = append(parts, auxText)
						case grammar.NegationMarker:
							parts = append(parts, markerText)
						case grammar.ObjectPhrase:
							parts = append(parts, obj.text)
						}
					}
					surfaceText := capitalize(strings.Join(parts, " ")) + "."
					key := surfaceText
					if _, seen := produced[key]; seen {
						continue
					}
					cand := Candidate{Meaning: m.Clone(), Text: surfaceText}
					cand.Assumptions = append(append([]Annotation(nil), sub.assumptions...), obj.assumptions...)
					cand.Losses = append(append([]Annotation(nil), sub.losses...), obj.losses...)
					produced[key] = cand
					if len(produced) > maxCandidates {
						return fail(status.ResourceLimit, "CANDIDATE_LIMIT", "event realization candidate budget exceeded")
					}
				}
			}
		}
	}
	if len(produced) == 0 {
		return fail(status.Unsupported, "TARGET_CONSTRUCTION_UNAVAILABLE", "no licensed event realization covers this meaning")
	}
	keys := make([]string, 0, len(produced))
	for key := range produced {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := Result{Outcome: status.OK}
	for _, key := range keys {
		result.Candidates = append(result.Candidates, produced[key])
	}
	if len(result.Candidates) > 1 {
		result.Outcome = status.Ambiguous
	}
	return result
}
