// Package features defines typed morphology requests; empty fields are unrequested.
package features

import "strings"

type Spec struct {
	Number string
	Case   string
	Gender string
	Person string
	Tense  string
	Degree string
	Onset  string
}

func (s Spec) Key() string {
	return strings.Join([]string{s.Number, s.Case, s.Gender, s.Person, s.Tense, s.Degree, s.Onset}, "|")
}
func (rule Spec) Matches(query Spec) bool {
	return (rule.Number == "" || rule.Number == query.Number) && (rule.Case == "" || rule.Case == query.Case) &&
		(rule.Gender == "" || rule.Gender == query.Gender) && (rule.Person == "" || rule.Person == query.Person) &&
		(rule.Tense == "" || rule.Tense == query.Tense) && (rule.Degree == "" || rule.Degree == query.Degree) &&
		(rule.Onset == "" || rule.Onset == query.Onset)
}
func (s Spec) CompatibleRequest(query Spec) bool {
	return (query.Number == "" || s.Number == "" || s.Number == query.Number) && (query.Case == "" || s.Case == "" || s.Case == query.Case) &&
		(query.Gender == "" || s.Gender == "" || s.Gender == query.Gender) && (query.Person == "" || s.Person == "" || s.Person == query.Person) &&
		(query.Tense == "" || s.Tense == "" || s.Tense == query.Tense) && (query.Degree == "" || s.Degree == "" || s.Degree == query.Degree) &&
		(query.Onset == "" || s.Onset == "" || s.Onset == query.Onset)
}
