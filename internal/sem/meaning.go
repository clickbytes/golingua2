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
	Scope    NodeKind = "scope"
)

type Operator string

const (
	Not Operator = "not"
	Ask Operator = "ask"
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
	// Operator and Content belong only to a scope, never to a lexical predicate.
	Operator     Operator
	Content      NodeID
	QuestionKind string // polar, who, what, where, how_many on Ask only
	QuestionRole string // exact licensed semantic argument role for a gap
	Reference    context.Deixis
	// Definiteness, Number and PronounGender describe the specific referring
	// occurrence, never a permanent characteristic of its CORE concept.
	Definiteness  string // "", "indefinite", "definite"
	Number        string // "", "singular", "plural"
	Cardinality   int    // 0 means unknown; positive means exact, bounded cardinality
	Person        string // "", "first", "second", "third" for lexical pronouns
	PronounGender string // "", "masculine", "feminine", "masculine_personal", "nonmasculine_personal"
	Tense         string // "present" for the first finite event constructions
	TimeReference string // previous_day or next_day, exact request-relative event time
	Existential   bool   // only on a spatial state: assertion of existence at a site
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
		if n.Kind == Scope {
			if n.Concept != "" || n.Value != "" || len(n.Arguments) != 0 || n.Reference != "" ||
				n.Definiteness != "" || n.Number != "" || n.Person != "" || n.PronounGender != "" || n.Tense != "" || n.TimeReference != "" || n.Cardinality != 0 || n.Existential ||
				n.Content == "" || (n.Operator != Not && n.Operator != Ask) ||
				(n.Operator == Not && (n.QuestionKind != "" || n.QuestionRole != "")) ||
				(n.Operator == Ask && (n.QuestionKind != "polar" && n.QuestionKind != "who" && n.QuestionKind != "what" && n.QuestionKind != "where" && n.QuestionKind != "how_many" ||
					(n.QuestionKind == "polar" && n.QuestionRole != "") || (n.QuestionKind != "polar" && n.QuestionRole == ""))) {
				return fmt.Errorf("invalid scope %q", n.ID)
			}
			byID[n.ID] = n
			continue
		}
		def, ok := catalog.Lookup(n.Concept)
		if !ok {
			return fmt.Errorf("unknown concept %q", n.Concept)
		}
		if n.Operator != "" || n.Content != "" || n.QuestionKind != "" || n.QuestionRole != "" || n.Cardinality < 0 || n.Cardinality > 20 {
			return fmt.Errorf("scope data on a lexical node %q", n.ID)
		}
		switch n.Kind {
		case Referent:
			if def.Kind != core.Entity || n.Value != "" || len(n.Arguments) > 0 || n.Tense != "" || n.TimeReference != "" || n.Existential ||
				(n.Reference != "" && n.Reference != context.Unresolved && n.Reference != context.Proximal && n.Reference != context.Distal) ||
				(n.Definiteness != "" && n.Definiteness != "indefinite" && n.Definiteness != "definite") ||
				(n.Number != "" && n.Number != "singular" && n.Number != "plural") ||
				(n.Person != "" && n.Person != "first" && n.Person != "second" && n.Person != "third") ||
				(n.Cardinality > 0 && (n.Person != "" || n.Cardinality == 1 && n.Number != "singular" || n.Cardinality > 1 && n.Number != "plural")) ||
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
			if def.Kind != core.Predicate || n.Reference != "" || n.Definiteness != "" || n.Number != "" || n.Person != "" || n.PronounGender != "" || n.Cardinality != 0 {
				return fmt.Errorf("invalid predicate node %q", n.ID)
			}
			if n.Kind == State {
				if n.Tense != "" || n.TimeReference != "" {
					return fmt.Errorf("invalid state tense %q", n.ID)
				}
				if def.ValueKind == "" {
					// A relation is a state with a typed role signature and
					// no property value; its identity is the relation concept.
					if !def.StateRelation || n.Value != "" || len(def.Roles) != 2 {
						return fmt.Errorf("invalid relational state %q", n.ID)
					}
				} else {
					if n.Existential {
						return fmt.Errorf("existence requires spatial relation at %q", n.ID)
					}
					v, ok := catalog.Lookup(n.Value)
					if !ok || v.Kind != def.ValueKind {
						return fmt.Errorf("invalid predicate value at %q", n.ID)
					}
				}
			} else if n.Existential || def.StateRelation || def.ValueKind != "" || n.Value != "" || (n.Tense != "present" && n.Tense != "past" && n.Tense != "future") || (n.Tense == "present" && n.TimeReference != "") || (n.Tense == "past" && n.TimeReference != "previous_day") || (n.Tense == "future" && n.TimeReference != "next_day") {
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
	if !ok || (root.Kind != State && root.Kind != Event && root.Kind != Scope) {
		return fmt.Errorf("invalid meaning root")
	}
	reachable := map[NodeID]bool{m.Root: true}
	if root.Kind == Scope {
		content, exists := byID[root.Content]
		if !exists || (content.Kind != Event && (root.Operator != Ask || content.Kind != State)) || content.ID == root.ID {
			return fmt.Errorf("scope content must be a licensed event or queried state")
		}
		if root.Operator == Ask && root.QuestionRole != "" {
			var target NodeID
			for _, a := range content.Arguments {
				if a.Role == root.QuestionRole {
					target = a.Target
				}
			}
			ref, ok := byID[target]
			if !ok || ref.Kind != Referent {
				return fmt.Errorf("question gap must bind an exact referent role")
			}
			if root.QuestionKind == "how_many" {
				if ref.Number != "plural" || ref.Cardinality != 0 || ref.Person != "" {
					return fmt.Errorf("quantity question has bound count or wrong type")
				}
			} else if ref.Concept != "entity.unspecified" || ref.Cardinality != 0 || ref.Person != "" {
				return fmt.Errorf("question answer gap must be unresolved entity")
			}
		}
		reachable[root.Content] = true
	}
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
	for _, n := range m.Nodes {
		if n.Kind != State || !n.Existential {
			continue
		}
		def, _ := catalog.Lookup(n.Concept)
		if !def.StateRelation {
			return fmt.Errorf("existential has no relation")
		}
		var located NodeID
		for _, a := range n.Arguments {
			if a.Role == def.Roles[0].Name {
				located = a.Target
			}
		}
		ref, ok := byID[located]
		if !ok || ref.Kind != Referent || ref.Cardinality < 1 {
			return fmt.Errorf("existential requires exactly quantified located referent")
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
		fmt.Fprintf(&b, "(%d,%s,%s,%s,%s,%s,%s,%d,%s,%s,%s,%s,%s,%t,%s,%s", names[id], n.Kind, n.Concept, n.Value, n.Reference, n.Definiteness, n.Number, n.Cardinality, n.Person, n.PronounGender, n.Tense, n.TimeReference, n.Operator, n.Existential, n.QuestionKind, n.QuestionRole)
		args := append([]Argument(nil), n.Arguments...)
		sort.Slice(args, func(i, j int) bool { return args[i].Role < args[j].Role })
		for _, a := range args {
			fmt.Fprintf(&b, ",%s:%s", a.Role, visit(a.Target))
		}
		if n.Kind == Scope {
			fmt.Fprintf(&b, ",content:%s", visit(n.Content))
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
