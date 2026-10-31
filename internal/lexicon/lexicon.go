// Package lexicon separates lexical identity and sense from neutral concepts.
package lexicon

import (
	"fmt"
	"sort"
	"strings"

	"github.com/clickbytes/golingua2/internal/core"
	"github.com/clickbytes/golingua2/internal/features"
)

type POS string

const (
	Noun       POS = "noun"
	Adjective  POS = "adjective"
	Verb       POS = "verb"
	Determiner POS = "determiner"
	Article    POS = "article"
	Pronoun    POS = "pronoun"
	Auxiliary  POS = "auxiliary"
	Particle   POS = "particle"
	Adposition POS = "adposition"
	Numeral    POS = "numeral"
)

type Entry struct {
	ID       string
	Lemma    string
	Language string
	POS      POS
	ClassID  string
	Stems    map[string]string
	Gender   string
	// Onset is lexical phonology, not an English spelling heuristic.
	Onset string
	// Countability determines article/number behavior without inventing cardinality.
	Countability string // "count", "mass", "plural_only" for nouns
	NumeralValue int    // exact cardinality spelled by a numeral, not a lexical sense
	// SiteClass is a lexical selection class (e.g. institution/tree/general),
	// never a CORE predicate or a sentence-specific condition.
	SiteClass     string
	PronounPerson string
	PronounNumber string
	PronounGender string
	// Overrides are attested exceptional cells; keys are features.Spec.Key().
	Overrides map[string]string
}

type Sense struct {
	ID       string
	LexemeID string
	Concept  core.ID
	// FrameIDs license a verb sense to exact language-local valency frames.
	FrameIDs []string
}

type Index struct {
	entries   map[string]Entry
	senses    map[string]Sense
	byConcept map[core.ID][]Entry
	byLexeme  map[string][]Sense
}

func Compile(entries []Entry, senses []Sense, catalog *core.Catalog, language string) (*Index, error) {
	if catalog == nil || language == "" || len(entries) == 0 || len(entries) > 10000 {
		return nil, fmt.Errorf("invalid lexical package")
	}
	i := &Index{entries: make(map[string]Entry, len(entries)), senses: make(map[string]Sense, len(senses)), byConcept: make(map[core.ID][]Entry), byLexeme: make(map[string][]Sense)}
	for _, e := range entries {
		if e.ID == "" || e.Lemma == "" || strings.TrimSpace(e.Lemma) != e.Lemma || e.Language != language || e.ClassID == "" || (e.POS != Noun && e.POS != Adjective && e.POS != Verb && e.POS != Determiner && e.POS != Article && e.POS != Pronoun && e.POS != Auxiliary && e.POS != Particle && e.POS != Adposition && e.POS != Numeral) {
			return nil, fmt.Errorf("invalid lexeme %q", e.ID)
		}
		if _, ok := i.entries[e.ID]; ok {
			return nil, fmt.Errorf("duplicate lexeme %q", e.ID)
		}
		if (e.POS == Numeral && (e.NumeralValue < 1 || e.NumeralValue > 20)) || (e.POS != Numeral && e.NumeralValue != 0) {
			return nil, fmt.Errorf("invalid numeral value for %q", e.ID)
		}
		if e.POS == Noun && e.Gender == "" && language == "pl" {
			return nil, fmt.Errorf("Polish noun %q requires gender", e.ID)
		}
		if e.POS == Noun && e.Countability != "" && e.Countability != "count" && e.Countability != "mass" && e.Countability != "plural_only" {
			return nil, fmt.Errorf("invalid noun countability %q", e.ID)
		}
		if e.SiteClass != "" && (e.POS != Noun || !validSiteClass(e.SiteClass)) {
			return nil, fmt.Errorf("invalid site class on %q", e.ID)
		}
		if e.POS == Pronoun && (e.PronounPerson != "first" && e.PronounPerson != "second" && e.PronounPerson != "third" || (e.PronounNumber != "" && e.PronounNumber != "singular" && e.PronounNumber != "plural")) {
			return nil, fmt.Errorf("invalid pronoun properties %q", e.ID)
		}
		if e.Onset != "" && e.Onset != "vowel" && e.Onset != "consonant" {
			return nil, fmt.Errorf("invalid lexical onset %q", e.ID)
		}
		e.Stems = copyMap(e.Stems)
		e.Overrides = copyMap(e.Overrides)
		i.entries[e.ID] = e
	}
	for _, s := range senses {
		if s.ID == "" || s.LexemeID == "" || s.Concept == "" {
			return nil, fmt.Errorf("invalid sense")
		}
		if _, ok := i.senses[s.ID]; ok {
			return nil, fmt.Errorf("duplicate sense %q", s.ID)
		}
		e, ok := i.entries[s.LexemeID]
		if !ok {
			return nil, fmt.Errorf("unbound sense %q", s.ID)
		}
		if e.POS == Auxiliary || e.POS == Particle || e.POS == Adposition || e.POS == Numeral {
			return nil, fmt.Errorf("structural operator lexeme cannot claim a semantic predicate: %q", s.ID)
		}
		d, ok := catalog.Lookup(s.Concept)
		if !ok {
			return nil, fmt.Errorf("unknown concept for %q", s.ID)
		}
		if (e.POS == Noun || e.POS == Pronoun) && d.Kind != core.Entity || e.POS == Adjective && d.Kind != core.Property || e.POS == Verb && d.Kind != core.Predicate {
			return nil, fmt.Errorf("lexical concept/POS mismatch at %q", s.ID)
		}
		if (e.POS == Verb && len(s.FrameIDs) == 0) || (e.POS != Verb && len(s.FrameIDs) > 0) {
			return nil, fmt.Errorf("invalid valency licenses for sense %q", s.ID)
		}
		s.FrameIDs = append([]string(nil), s.FrameIDs...)
		i.senses[s.ID] = s
		i.byLexeme[s.LexemeID] = append(i.byLexeme[s.LexemeID], s)
		i.byConcept[s.Concept] = append(i.byConcept[s.Concept], e)
	}
	for k, v := range i.byConcept {
		sort.Slice(v, func(a, b int) bool { return v[a].ID < v[b].ID })
		i.byConcept[k] = v
	}
	for k, v := range i.byLexeme {
		sort.Slice(v, func(a, b int) bool { return v[a].ID < v[b].ID })
		i.byLexeme[k] = v
	}
	return i, nil
}

func (i *Index) Lookup(id string) (Entry, bool) {
	if i == nil {
		return Entry{}, false
	}
	e, ok := i.entries[id]
	e.Stems = copyMap(e.Stems)
	e.Overrides = copyMap(e.Overrides)
	return e, ok
}
func (i *Index) ByConcept(id core.ID) []Entry {
	if i == nil {
		return nil
	}
	v := i.byConcept[id]
	out := make([]Entry, len(v))
	for k, x := range v {
		out[k] = x
		out[k].Stems = copyMap(x.Stems)
		out[k].Overrides = copyMap(x.Overrides)
	}
	return out
}
func (i *Index) SenseFor(lexemeID string, concept core.ID) (Sense, bool) {
	if i == nil {
		return Sense{}, false
	}
	for _, s := range i.senses {
		if s.LexemeID == lexemeID && s.Concept == concept {
			return s, true
		}
	}
	return Sense{}, false
}

// SensesFor never guesses semantic identity from a lexeme's spelling or ID.
func (i *Index) SensesFor(lexemeID string) []Sense {
	if i == nil {
		return nil
	}
	v := append([]Sense(nil), i.byLexeme[lexemeID]...)
	for k := range v {
		v[k].FrameIDs = append([]string(nil), v[k].FrameIDs...)
	}
	return v
}

func (i *Index) Entries() []Entry {
	if i == nil {
		return nil
	}
	ids := make([]string, 0, len(i.entries))
	for id := range i.entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	v := make([]Entry, 0, len(ids))
	for _, id := range ids {
		e, _ := i.Lookup(id)
		v = append(v, e)
	}
	return v
}
func copyMap[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return nil
	}
	o := make(map[K]V, len(m))
	for k, v := range m {
		o[k] = v
	}
	return o
}

// OverrideKey is the stable key for one exceptional morphological form.
func OverrideKey(f features.Spec) string { return f.Key() }

// Language packs own the site-class vocabulary; the neutral lexicon validates
// shape, not an English/Polish list of concepts or prepositions.
func validSiteClass(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}
