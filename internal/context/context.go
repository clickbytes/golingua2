// Package context holds explicit request-local information, never persistent world facts.
package context

type Deixis string

const (
	Unresolved Deixis = "unresolved"
	Proximal   Deixis = "proximal"
	Distal     Deixis = "distal"
)

type Request struct {
	Deixis Deixis
	// The English "you" has no morphological number distinction. This
	// request-local value must not become an attribute of a CORE concept.
	AddresseeNumber string // "", "singular", "plural"
}

func (r Request) Valid() bool {
	return (r.Deixis == Unresolved || r.Deixis == Proximal || r.Deixis == Distal || r.Deixis == "") && (r.AddresseeNumber == "" || r.AddresseeNumber == "singular" || r.AddresseeNumber == "plural")
}
