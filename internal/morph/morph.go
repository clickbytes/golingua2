// Package morph compiles reusable procedural inflection classes and one inverse index.
// It deliberately does not inspect sentence strings or semantic concept names.
package morph

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua2/internal/features"
	"github.com/clickbytes/golingua2/internal/lexicon"
)

type OpKind string

const (
	SelectLemma   OpKind = "select_lemma"
	SelectStem    OpKind = "select_stem"
	StripSuffix   OpKind = "strip_suffix"
	Append        OpKind = "append"
	ReplaceSuffix OpKind = "replace_suffix"
)

type Operation struct {
	Kind OpKind
	Arg  string
	With string
}

type Rule struct {
	Features features.Spec
	Steps    []Operation
}
type Class struct {
	ID       string
	Language string
	POS      lexicon.POS
	Rules    []Rule
}

type Analysis struct {
	LexemeID string
	POS      lexicon.POS
	Features features.Spec
	Surface  string
}
type Generator struct {
	classes map[string]Class
	lexemes map[string]lexicon.Entry
	inverse map[string][]Analysis
}

func Compile(classes []Class, entries []lexicon.Entry) (*Generator, error) {
	if len(classes) == 0 || len(classes) > 1024 || len(entries) > 10000 {
		return nil, fmt.Errorf("invalid morphology compilation size")
	}
	g := &Generator{classes: make(map[string]Class, len(classes)), lexemes: make(map[string]lexicon.Entry, len(entries)), inverse: make(map[string][]Analysis)}
	for _, c := range classes {
		if c.ID == "" || c.Language == "" || len(c.Rules) == 0 || len(c.Rules) > 100 {
			return nil, fmt.Errorf("invalid morphological class %q", c.ID)
		}
		if _, ok := g.classes[c.ID]; ok {
			return nil, fmt.Errorf("duplicate morphological class %q", c.ID)
		}
		seen := map[string]bool{}
		for _, r := range c.Rules {
			if len(r.Steps) == 0 || len(r.Steps) > 32 {
				return nil, fmt.Errorf("invalid rule in %q", c.ID)
			}
			if seen[r.Features.Key()] {
				return nil, fmt.Errorf("duplicate cell in class %q", c.ID)
			}
			seen[r.Features.Key()] = true
			if r.Steps[0].Kind != SelectLemma && r.Steps[0].Kind != SelectStem {
				return nil, fmt.Errorf("rule must select lemma/stem in %q", c.ID)
			}
			for j, s := range r.Steps {
				switch s.Kind {
				case SelectLemma:
					if j != 0 || s.Arg != "" || s.With != "" {
						return nil, fmt.Errorf("invalid select lemma")
					}
				case SelectStem:
					if j != 0 || s.Arg == "" || s.With != "" {
						return nil, fmt.Errorf("invalid select stem")
					}
				case StripSuffix, Append:
					if j == 0 || s.Arg == "" || s.With != "" {
						return nil, fmt.Errorf("invalid suffix operation")
					}
				case ReplaceSuffix:
					if j == 0 || s.Arg == "" {
						return nil, fmt.Errorf("invalid replacement")
					}
				default:
					return nil, fmt.Errorf("unsupported operation %q", s.Kind)
				}
			}
		}
		frozen := Class{ID: c.ID, Language: c.Language, POS: c.POS, Rules: make([]Rule, len(c.Rules))}
		for ri, r := range c.Rules {
			frozen.Rules[ri] = Rule{Features: r.Features, Steps: append([]Operation(nil), r.Steps...)}
		}
		g.classes[c.ID] = frozen
	}
	readings := 0
	for _, e := range entries {
		c, ok := g.classes[e.ClassID]
		if !ok || c.Language != e.Language || c.POS != e.POS {
			return nil, fmt.Errorf("missing/incompatible class for %q", e.ID)
		}
		if _, ok := g.lexemes[e.ID]; ok {
			return nil, fmt.Errorf("duplicate morphological lexeme %q", e.ID)
		}
		g.lexemes[e.ID] = e
		licensed := map[string]bool{}
		for _, r := range c.Rules {
			licensed[r.Features.Key()] = true
			surface, err := run(e, r)
			if override, ok := e.Overrides[r.Features.Key()]; ok {
				surface = override
			}
			if err != nil && surface == "" {
				return nil, fmt.Errorf("%s/%s: %w", e.ID, r.Features.Key(), err)
			}
			if !safeSurface(surface) {
				return nil, fmt.Errorf("invalid generated surface for %q", e.ID)
			}
			readings++
			if readings > 100_000 {
				return nil, fmt.Errorf("morphology finite reading budget exceeded")
			}
			g.inverse[strings.ToLower(surface)] = append(g.inverse[strings.ToLower(surface)], Analysis{LexemeID: e.ID, POS: e.POS, Features: r.Features, Surface: surface})
		}
		for k := range e.Overrides {
			if !licensed[k] {
				return nil, fmt.Errorf("unlicensed exception cell %q for %q", k, e.ID)
			}
		}
	}
	for k := range g.inverse {
		sort.Slice(g.inverse[k], func(i, j int) bool {
			a, b := g.inverse[k][i], g.inverse[k][j]
			if a.LexemeID != b.LexemeID {
				return a.LexemeID < b.LexemeID
			}
			return a.Features.Key() < b.Features.Key()
		})
	}
	return g, nil
}

func safeSurface(s string) bool {
	return s != "" && len(s) <= 160 && utf8.ValidString(s) && strings.TrimSpace(s) == s && !strings.ContainsAny(s, "\t\n\r")
}

func run(e lexicon.Entry, rule Rule) (string, error) {
	surface := ""
	for _, step := range rule.Steps {
		switch step.Kind {
		case SelectLemma:
			surface = e.Lemma
		case SelectStem:
			root, ok := e.Stems[step.Arg]
			if !ok || !safeSurface(root) {
				return "", fmt.Errorf("required principal stem %q unavailable", step.Arg)
			}
			surface = root
		case StripSuffix:
			if !strings.HasSuffix(surface, step.Arg) {
				return "", fmt.Errorf("cannot strip suffix %q from %q", step.Arg, surface)
			}
			surface = strings.TrimSuffix(surface, step.Arg)
		case ReplaceSuffix:
			if !strings.HasSuffix(surface, step.Arg) {
				return "", fmt.Errorf("cannot replace suffix %q from %q", step.Arg, surface)
			}
			surface = strings.TrimSuffix(surface, step.Arg) + step.With
		case Append:
			surface += step.Arg
		}
	}
	return surface, nil
}

// Generate selects exactly one matching morphological cell; it never guesses
// a gender, tense or case from a lexeme's spelling.
func (g *Generator) Generate(lexemeID string, query features.Spec) (string, error) {
	if g == nil {
		return "", fmt.Errorf("morphology unavailable")
	}
	e, ok := g.lexemes[lexemeID]
	if !ok {
		return "", fmt.Errorf("unknown lexeme %q", lexemeID)
	}
	c := g.classes[e.ClassID]
	matches := []Rule{}
	for _, r := range c.Rules {
		if r.Features.Matches(query) {
			matches = append(matches, r)
		}
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("morphology %q has %d matching cells, require exactly one", lexemeID, len(matches))
	}
	r := matches[0]
	if v, ok := e.Overrides[r.Features.Key()]; ok {
		return v, nil
	}
	return run(e, r)
}

func (g *Generator) Analyze(surface string) []Analysis {
	if g == nil {
		return nil
	}
	a := g.inverse[strings.ToLower(surface)]
	return append([]Analysis(nil), a...)
}
func (g *Generator) CountForms() int {
	if g == nil {
		return 0
	}
	n := 0
	for _, v := range g.inverse {
		n += len(v)
	}
	return n
}
