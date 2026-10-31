// Package sem owns the single authoritative, language-neutral Meaning IR.
package sem

import (
	"fmt"
	"sort"
	"strings"

	"github.com/clickbytes/golingua2/internal/context"
	"github.com/clickbytes/golingua2/internal/core"
)

type NodeID string
type NodeKind string

const (
	Referent NodeKind = "referent"
	State    NodeKind = "state"
	Event    NodeKind = "event"
)

type Argument struct {
	Role   string
	Target NodeID
}

type Node struct {
	ID      NodeID
	Kind    NodeKind
	Concept core.ID
	// Value is the typed concept value of a state, not a surface form.
	Value     core.ID
	Arguments []Argument
	Reference context.Deixis
	// Definiteness, Number and PronounGender describe the specific referring
	// occurrence, never a permanent characteristic of its CORE concept.
	Definiteness  string // "", "indefinite", "definite"
	Number        string // "", "singular", "plural"
	Person        string // "", "first", "second", "third" for lexical pronouns
	PronounGender string // "", "masculine", "feminine", "masculine_personal", "nonmasculine_personal"
	Tense         string // "present" for the first finite event constructions
}

type Meaning struct {
	Root  NodeID
	Nodes []Node
}

func (m Meaning) Clone() Meaning {
	out := Meaning{Root: m.Root, Nodes: make([]Node, len(m.Nodes))}
	for i, n := range m.Nodes {
		out.Nodes[i] = n
		out.Nodes[i].Arguments = append([]Argument(nil), n.Arguments...)
	}
	return out
}

func Validate(m Meaning, catalog *core.Catalog) error {
	if catalog == nil || len(m.Nodes) == 0 || len(m.Nodes) > 64 {
		return fmt.Errorf("missing catalog or invalid meaning size")
	}
	byID := make(map[NodeID]Node, len(m.Nodes))
	for _, n := range m.Nodes {
		if n.ID == "" || len(n.ID) > 80 {
			return fmt.Errorf("invalid occurrence ID")
		}
		if _, ok := byID[n.ID]; ok {
			return fmt.Errorf("duplicate occurrence %q", n.ID)
		}
		def, ok := catalog.Lookup(n.Concept)
		if !ok {
			return fmt.Errorf("unknown concept %q", n.Concept)
		}
		switch n.Kind {
		case Referent:
			if def.Kind != core.Entity || n.Value != "" || len(n.Arguments) > 0 || n.Tense != "" ||
				(n.Reference != "" && n.Reference != context.Unresolved && n.Reference != context.Proximal && n.Reference != context.Distal) ||
				(n.Definiteness != "" && n.Definiteness != "indefinite" && n.Definiteness != "definite") ||
				(n.Number != "" && n.Number != "singular" && n.Number != "plural") ||
				(n.Person != "" && n.Person != "first" && n.Person != "second" && n.Person != "third") ||
				(n.PronounGender != "" && n.PronounGender != "masculine" && n.PronounGender != "feminine" && n.PronounGender != "masculine_personal" && n.PronounGender != "nonmasculine_personal") {
				return fmt.Errorf("invalid referent %q", n.ID)
			}
			if n.Person != "" && n.Concept != "entity.person" {
				return fmt.Errorf("person anchor cannot be assigned to lexical noun %q", n.ID)
			}
			if n.Person == "" && n.PronounGender != "" {
				return fmt.Errorf("pronoun gender lacks person anchor %q", n.ID)
			}
		case State, Event:
			if def.Kind != core.Predicate || n.Reference != "" || n.Definiteness != "" || n.Number != "" || n.Person != "" || n.PronounGender != "" {
				return fmt.Errorf("invalid predicate node %q", n.ID)
			}
			if n.Kind == State {
				if def.ValueKind == "" || n.Tense != "" {
					return fmt.Errorf("invalid state %q", n.ID)
				}
				v, ok := catalog.Lookup(n.Value)
				if !ok || v.Kind != def.ValueKind {
					return fmt.Errorf("invalid predicate value at %q", n.ID)
				}
			} else if def.ValueKind != "" || n.Value != "" || n.Tense != "present" {
				return fmt.Errorf("invalid event %q", n.ID)
			}
			if len(n.Arguments) != len(def.Roles) {
				return fmt.Errorf("predicate %q has incomplete roles", n.ID)
			}
			seen := map[string]bool{}
			for _, a := range n.Arguments {
				roleOk := false
				for _, r := range def.Roles {
					if r.Name == a.Role {
						roleOk = true
						break
					}
				}
				if !roleOk || seen[a.Role] {
					return fmt.Errorf("unlicensed/duplicate role %q", a.Role)
				}
				seen[a.Role] = true
			}
		default:
			return fmt.Errorf("unsupported node kind %q", n.Kind)
		}
		byID[n.ID] = n
	}
	root, ok := byID[m.Root]
	if !ok || (root.Kind != State && root.Kind != Event) {
		return fmt.Errorf("invalid meaning root")
	}
	reachable := map[NodeID]bool{m.Root: true}
	for _, n := range m.Nodes {
		if n.Kind != State && n.Kind != Event {
			continue
		}
		def, _ := catalog.Lookup(n.Concept)
		for _, a := range n.Arguments {
			target, ok := byID[a.Target]
			if !ok {
				return fmt.Errorf("dangling role %q", a.Target)
			}
			for _, r := range def.Roles {
				if r.Name == a.Role {
					if r.Target == core.Entity && target.Kind != Referent {
						return fmt.Errorf("invalid target kind for %q", a.Role)
					}
					if r.Target == core.Predicate && target.Kind != State && target.Kind != Event {
						return fmt.Errorf("invalid target kind for %q", a.Role)
					}
				}
			}
			reachable[a.Target] = true
		}
	}
	if len(reachable) != len(m.Nodes) {
		return fmt.Errorf("orphaned occurrence")
	}
	return nil
}

// CanonicalKey preserves role binding and repeated occurrence identity, but
// ignores graph-local IDs, field order and node order.
func CanonicalKey(m Meaning, catalog *core.Catalog) (string, error) {
	if err := Validate(m, catalog); err != nil {
		return "", err
	}
	byID := map[NodeID]Node{}
	for _, n := range m.Nodes {
		byID[n.ID] = n
	}
	names := map[NodeID]int{}
	var visit func(NodeID) string
	visit = func(id NodeID) string {
		if v, ok := names[id]; ok {
			return fmt.Sprintf("@%d", v)
		}
		names[id] = len(names) + 1
		n := byID[id]
		b := strings.Builder{}
		fmt.Fprintf(&b, "(%d,%s,%s,%s,%s,%s,%s,%s,%s,%s", names[id], n.Kind, n.Concept, n.Value, n.Reference, n.Definiteness, n.Number, n.Person, n.PronounGender, n.Tense)
		args := append([]Argument(nil), n.Arguments...)
		sort.Slice(args, func(i, j int) bool { return args[i].Role < args[j].Role })
		for _, a := range args {
			fmt.Fprintf(&b, ",%s:%s", a.Role, visit(a.Target))
		}
		b.WriteByte(')')
		return b.String()
	}
	return visit(m.Root), nil
}
func Equivalent(a, b Meaning, c *core.Catalog) (bool, error) {
	x, e := CanonicalKey(a, c)
	if e != nil {
		return false, e
	}
	y, e := CanonicalKey(b, c)
	return e == nil && x == y, e
}
