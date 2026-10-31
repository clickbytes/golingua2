package a1_test

import (
	"fmt"
	"strings"
	"testing"
)

// A1Case is an authored bilingual test requirement, never a language rule.
// EN and PL are illustrative correct realizations, not an assertion that each
// language always has exactly one possible realization of the same meaning.
// The independent typed-IR fixtures and actual engine acceptance tests are
// delivered in later implementation stages, NOT by these inventory tests.
type A1Case struct {
	ID      string
	Group   string
	PL      string
	EN      string
	Context ContextFlags
}

type ContextFlags uint8

const NoExtraContext ContextFlags = 0

const (
	NearDeixis ContextFlags = 1 << iota
	SingularAddressee
	PluralAddressee
)

func a1(id int, group, pl, en string, flags ...ContextFlags) A1Case {
	var context ContextFlags
	for _, flag := range flags {
		context |= flag
	}
	return A1Case{ID: fmt.Sprintf("A1-%03d", id), Group: group, PL: pl, EN: en, Context: context}
}

// One hundred authored PL–EN sentence pairs, i.e. 100 semantic scenarios and
// two translation directions each. There are no sentence fixtures in JSON.
// The corpus includes intentional underspecification (e.g. Polish articles),
// so a production acceptance runner must validate meaning, assumptions,
// losses, and allowed alternatives rather than requiring string equality.
var cases = []A1Case{
	// 01–10: identification / classification.
	a1(1, "identification", "To jest kot.", "This is a cat.", NearDeixis),
	a1(2, "identification", "To jest pies.", "This is a dog.", NearDeixis),
	a1(3, "identification", "To jest książka.", "This is a book.", NearDeixis),
	a1(4, "identification", "To jest dom.", "This is a house.", NearDeixis),
	a1(5, "identification", "To jest stół.", "This is a table.", NearDeixis),
	a1(6, "identification", "To jest krzesło.", "This is a chair.", NearDeixis),
	a1(7, "identification", "To jest nauczyciel.", "This is a teacher.", NearDeixis),
	a1(8, "identification", "To jest uczeń.", "This is a student.", NearDeixis),
	a1(9, "identification", "To jest jabłko.", "This is an apple.", NearDeixis),
	a1(10, "identification", "To jest okno.", "This is a window.", NearDeixis),

	// 11–20: properties, adjective agreement, and demonstratives.
	a1(11, "properties", "Ten kot jest czarny.", "This cat is black.", NearDeixis),
	a1(12, "properties", "Ten pies jest mały.", "This dog is small.", NearDeixis),
	a1(13, "properties", "Ta książka jest nowa.", "This book is new.", NearDeixis),
	a1(14, "properties", "Ta torba jest duża.", "This bag is big.", NearDeixis),
	a1(15, "properties", "To okno jest otwarte.", "This window is open.", NearDeixis),
	a1(16, "properties", "To jabłko jest czerwone.", "This apple is red.", NearDeixis),
	a1(17, "properties", "Ten dom jest duży.", "This house is big.", NearDeixis),
	a1(18, "properties", "Ta szkoła jest nowa.", "This school is new.", NearDeixis),
	a1(19, "properties", "To krzesło jest małe.", "This chair is small.", NearDeixis),
	a1(20, "properties", "Ta kawa jest gorąca.", "This coffee is hot.", NearDeixis),

	// 21–30: have/possession, personal agreement, count vs mass.
	a1(21, "possession", "Mam książkę.", "I have a book."),
	a1(22, "possession", "Masz psa.", "You have a dog.", SingularAddressee),
	a1(23, "possession", "On ma kota.", "He has a cat."),
	a1(24, "possession", "Ona ma torbę.", "She has a bag."),
	a1(25, "possession", "Mamy dom.", "We have a house."),
	a1(26, "possession", "Macie rower.", "You have a bicycle.", PluralAddressee),
	a1(27, "possession", "Oni mają samochód.", "They have a car."),
	a1(28, "possession", "Mam telefon.", "I have a phone."),
	a1(29, "possession", "Ona ma jabłko.", "She has an apple."),
	a1(30, "possession", "Mamy wodę.", "We have water."),

	// 31–40: transitive predicates and argument-dependent case.
	a1(31, "transitive", "Nauczyciel czyta książkę.", "A teacher reads a book."),
	a1(32, "transitive", "Uczeń czyta list.", "A student reads a letter."),
	a1(33, "transitive", "Pies widzi kota.", "A dog sees a cat."),
	a1(34, "transitive", "Kot widzi ptaka.", "A cat sees a bird."),
	a1(35, "transitive", "Dziewczynka pije wodę.", "A girl drinks water."),
	a1(36, "transitive", "Chłopiec je jabłko.", "A boy eats an apple."),
	a1(37, "transitive", "Mama otwiera okno.", "A mother opens a window."),
	a1(38, "transitive", "Tata zamyka drzwi.", "A father closes a door."),
	a1(39, "transitive", "Lekarz pomaga dziecku.", "A doctor helps a child."),
	a1(40, "transitive", "Nauczyciel pisze list.", "A teacher writes a letter."),

	// 41–50: intransitive predicates, finite verb agreement.
	a1(41, "intransitive", "Kot śpi.", "A cat sleeps."),
	a1(42, "intransitive", "Pies biega.", "A dog runs."),
	a1(43, "intransitive", "Ptak lata.", "A bird flies."),
	a1(44, "intransitive", "Dziecko siedzi.", "A child sits."),
	a1(45, "intransitive", "Nauczyciel pracuje.", "A teacher works."),
	a1(46, "intransitive", "Mama gotuje.", "A mother cooks."),
	a1(47, "intransitive", "Chłopiec chodzi.", "A boy walks."),
	a1(48, "intransitive", "Dziewczynka śpiewa.", "A girl sings."),
	a1(49, "intransitive", "Pies szczeka.", "A dog barks."),
	a1(50, "intransitive", "Dziecko płacze.", "A child cries."),

	// 51–60: event negation and Polish government under negation.
	a1(51, "negation", "Kot nie śpi.", "A cat does not sleep."),
	a1(52, "negation", "Pies nie biega.", "A dog does not run."),
	a1(53, "negation", "Nauczyciel nie czyta książki.", "A teacher does not read a book."),
	a1(54, "negation", "Chłopiec nie je jabłka.", "A boy does not eat an apple."),
	a1(55, "negation", "Dziewczynka nie pije wody.", "A girl does not drink water."),
	a1(56, "negation", "Ptak nie lata.", "A bird does not fly."),
	a1(57, "negation", "Mama nie otwiera okna.", "A mother does not open a window."),
	a1(58, "negation", "Tata nie zamyka okna.", "A father does not close a window."),
	a1(59, "negation", "Nie mam samochodu.", "I do not have a car."),
	a1(60, "negation", "Nie mamy wody.", "We do not have water."),

	// 61–70: polar questions and basic WH questions. Stage after location/count.
	a1(61, "questions", "Czy to jest kot?", "Is this a cat?", NearDeixis),
	a1(62, "questions", "Czy ten pies biega?", "Does this dog run?", NearDeixis),
	a1(63, "questions", "Czy masz psa?", "Do you have a dog?", SingularAddressee),
	a1(64, "questions", "Czy nauczyciel czyta książkę?", "Does the teacher read a book?"),
	a1(65, "questions", "Czy ta książka jest nowa?", "Is this book new?", NearDeixis),
	a1(66, "questions", "Czy masz wodę?", "Do you have water?", SingularAddressee),
	a1(67, "questions", "Kto czyta książkę?", "Who reads a book?"),
	a1(68, "questions", "Co czyta nauczyciel?", "What does the teacher read?"),
	a1(69, "questions", "Gdzie jest kot?", "Where is the cat?"),
	a1(70, "questions", "Ile masz książek?", "How many books do you have?", SingularAddressee),

	// 71–80: locative relations, PP government, and language-specific adpositions.
	a1(71, "location", "Kot jest w domu.", "The cat is in the house."),
	a1(72, "location", "Książka jest na stole.", "The book is on the table."),
	a1(73, "location", "Pies jest pod stołem.", "The dog is under the table."),
	a1(74, "location", "Uczeń jest w szkole.", "The student is at school."),
	a1(75, "location", "Mama jest w kuchni.", "The mother is in the kitchen."),
	a1(76, "location", "Rower jest przed domem.", "The bicycle is in front of the house."),
	a1(77, "location", "Samochód jest za domem.", "The car is behind the house."),
	a1(78, "location", "Dziecko jest obok mamy.", "The child is next to the mother."),
	a1(79, "location", "Ptak jest na drzewie.", "The bird is in the tree."),
	a1(80, "location", "Stół jest w pokoju.", "The table is in the room."),

	// 81–90: exact cardinality, count/mass, grammatical number, existential clauses.
	a1(81, "quantity", "Mam dwie książki.", "I have two books."),
	a1(82, "quantity", "Mam trzy jabłka.", "I have three apples."),
	a1(83, "quantity", "Widzę dwa psy.", "I see two dogs."),
	a1(84, "quantity", "Na stole są cztery jabłka.", "There are four apples on the table."),
	a1(85, "quantity", "W domu są dwa koty.", "There are two cats in the house."),
	a1(86, "quantity", "W pokoju jest jedno okno.", "There is one window in the room."),
	a1(87, "quantity", "Mam pięć książek.", "I have five books."),
	a1(88, "quantity", "Ona ma dwa rowery.", "She has two bicycles."),
	a1(89, "quantity", "Mamy trzy krzesła.", "We have three chairs."),
	a1(90, "quantity", "Na stole są dwa kubki.", "There are two mugs on the table."),

	// 91–100: basic past/future, tense anchoring and controlled aspect alternatives.
	a1(91, "time", "Wczoraj kot spał.", "Yesterday a cat slept."),
	a1(92, "time", "Wczoraj pies biegał.", "Yesterday a dog ran."),
	a1(93, "time", "Wczoraj nauczyciel czytał książkę.", "Yesterday a teacher read a book."),
	a1(94, "time", "Wczoraj mama gotowała.", "Yesterday a mother cooked."),
	a1(95, "time", "Wczoraj padał deszcz.", "It rained yesterday."),
	a1(96, "time", "Jutro kot będzie spał.", "Tomorrow a cat will sleep."),
	a1(97, "time", "Jutro nauczyciel będzie czytał książkę.", "Tomorrow a teacher will read a book."),
	a1(98, "time", "Jutro dziecko będzie biegało.", "Tomorrow a child will run."),
	a1(99, "time", "Jutro mama będzie gotowała.", "Tomorrow a mother will cook."),
	a1(100, "time", "Jutro będę pracować.", "I will work tomorrow."),
}

// These are inventory checks, not translation tests: they do not import or
// pretend to execute a translator. A separate acceptance runner must invoke
// the real engine and verify all six directions against independent IR oracles.
func TestA1CorpusIntegrity(t *testing.T) {
	const expected = 100
	if got := len(cases); got != expected {
		t.Fatalf("want %d A1 cases, got %d", expected, got)
	}
	counts := make(map[string]int)
	plSeen := make(map[string]string)
	enSeen := make(map[string]string)
	for i, tc := range cases {
		wantID := fmt.Sprintf("A1-%03d", i+1)
		if tc.ID != wantID {
			t.Fatalf("record %d has ID %q; want %q", i+1, tc.ID, wantID)
		}
		if strings.TrimSpace(tc.PL) == "" || strings.TrimSpace(tc.EN) == "" {
			t.Errorf("%s missing Polish or English source", tc.ID)
		}
		if !(strings.HasSuffix(tc.PL, ".") || strings.HasSuffix(tc.PL, "?")) ||
			!(strings.HasSuffix(tc.EN, ".") || strings.HasSuffix(tc.EN, "?")) {
			t.Errorf("%s missing sentence punctuation", tc.ID)
		}
		if former, exists := plSeen[tc.PL]; exists {
			t.Errorf("%s duplicates Polish text of %s", tc.ID, former)
		}
		if former, exists := enSeen[tc.EN]; exists {
			t.Errorf("%s duplicates English text of %s", tc.ID, former)
		}
		plSeen[tc.PL] = tc.ID
		enSeen[tc.EN] = tc.ID
		if tc.Context&SingularAddressee != 0 && tc.Context&PluralAddressee != 0 {
			t.Errorf("%s has contradictory addressee number", tc.ID)
		}
		counts[tc.Group]++
	}
	for _, g := range []string{
		"identification", "properties", "possession", "transitive", "intransitive",
		"negation", "questions", "location", "quantity", "time",
	} {
		if counts[g] != 10 {
			t.Errorf("group %q: want 10, got %d", g, counts[g])
		}
		delete(counts, g)
	}
	if len(counts) != 0 {
		t.Errorf("unexpected groups: %v", counts)
	}
}
