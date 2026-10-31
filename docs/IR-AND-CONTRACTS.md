# GoLingua 2 — kontrakty typów, semantyki i gramatyki (RFC-002)

**Status:** projekt do implementacji, nie API już istniejącej biblioteki. **Źródło prawdy:** deklaracje języka i niezależny IR; nie wynik testowego tłumaczenia.

## A. Typy i własność danych

| Typ | Właściciel | Zakres identyfikatora | Co zawiera |
| --- | --- | --- | --- |
| `ConceptID` | `core` | globalny, wersjonowany | sort i role semantyczne |
| `LexemeID` | `languages/{lang}` | w języku | lemma, POS, klasa i tematy |
| `SenseID` | `lexicon` | w języku | wiązanie do konceptu |
| `MorphClassID` | `morph`/pakiet języka | w języku | zasady tworzenia komórek |
| `FrameID` | `grammar`/pakiet języka | w języku | dozwolone sloty i rządzenie |
| `ConstructionID` | `grammar`/pakiet języka | w języku | składane frazy i role |
| `NodeID` | `sem` | jeden graf znaczenia | konkretne zdarzenie/osoba/przedmiot |
| `ReferentID` | `context` | jedno żądanie | tożsamość znana/nieznana w kontekście |
| `EvidenceID` | `engine` | odczyt | ślad źródła decyzji, a nie fakt |

Tych typów nie wolno utożsamiać przez porównanie tego samego stringa. W szczególności `NodeID` może mieć wartość `e1` tylko w obrębie jednego grafu.

## B. Wstępny interfejs semantyczny w Go

Szkic nazw i pól, wymagający przeglądu przed zamrożeniem API:

```go
type Sort uint8 // Entity, Event, State, Relation, Proposition, Interval, Quantity, Location

type Node struct {
    ID      NodeID
    Sort    Sort
    Concept ConceptID
    Args    []RoleTarget    // relacja nazwana i zweryfikowana w sygnaturze konceptu
    Time    *TemporalFrame  // tylko tam, gdzie typ i kontrakt to dopuszczają
    Count   *Cardinality    // jeśli node jest referentem lub liczoną grupą
}

type RoleTarget struct { Role RoleID; Target NodeID }

type Expr interface { isSemanticExpr() }  // w implementacji zamknięta suma typów
// Atom(nodeID), Not(expr), And(left,right), Exists(binding,body), Ask(query,body)

type Meaning struct {
    Version  uint16
    Nodes    []Node
    Root     Expr
    Act      SpeechAct     // Assert/Ask
}

type Context struct {
    Speaker   *ReferentID
    Addressee *ReferentID
    // Optional number only when explicitly known about this addressee.
    AddresseeNumber Knowledge[GrammaticalNumber]
    Deixis          Knowledge[DeixisValue]
    ReferenceTime   *time.Time
}

type Evidence struct {
    OwnerPredicate NodeID
    MeaningNode    NodeID
    Role           *RoleID
    SourceSpan     Span
    FormReading    MorphReadingID
}
```

`Expr` musi w implementacji być zamkniętym zbiorem wariantów, sprawdzanym przez kompilator. Jeśli operator nie jest wyrażalny, ma powstać jawny `unsupported/incomplete`, nie `nil` lub cicha utrata członu.

### Reguły walidacji

- Każde `NodeID` występuje dokładnie raz, każda rola kieruje do istniejącego węzła dopuszczonego sortu.
- Dwie różne role wymagające odrębnych uczestników nie mogą zostać scalone, chyba że graf ma jawną tożsamość.
- Każdy akt wypowiedzi ma jedną kompletną, zweryfikowaną formułę; `Ask` nie wymyśla wartości pytanej zmiennej.
- `Not`/`Exists`/`Ask` posiadają własne zakresy i identyfikatory powiązań; zagnieżdżenie nie może zostać przestawione przez linearyzację.
- Semantyczna kardynalność i gramatyczna liczba są osobnymi cechami; nie wolno wnioskować jednej z drugiej przy pluralia tantum.
- Oznaczenie podmiotu angielskim `it` przy pogodzie nie wprowadza nieistniejącego referenta do IR.
- Wszelkie adresy `Evidence` sprawdzamy po `OwnerPredicate + Role + MeaningNode`, a nie po samym ConceptID.

## C. Przykład z dwoma kotami

```text
Not(see(event1, agent=cat#1, patient=cat#2))
```

Znaczenie ma trzy odrębne wystąpienia. Polski analizator `Kot nie widzi kota` rozróżnia mianownik podmiotu i dopełniacz po negacji. Angielski plan używa `cat` jako tej samej klasy wyrazu, ale dwóch osobnych NP. Nie wolno kopiować źródłowej określoności między `cat#1` i `cat#2`.

## D. Kontrakt leksemu i sensu

```go
// Docelowy szkic. Wszystkie wartości są danymi.
Lexeme{
    ID: "pl:kot/n", Lemma: "kot", POS: Noun,
    MorphClass: "pl:noun.masc.animate.hard",
    Stems: Stems{"basic":"kot"},
    Properties: {Gender: MasculineAnimate},
}
Sense{ID:"pl:kot/n#animal", Lexeme:"pl:kot/n", Concept:"core:cat"}
```

`Cat` jest konceptem; leksem `kot` i jego morfologia są danymi PL. Dodanie np. `ryś` albo innego regularnego rzeczownika tej samej klasy nie tworzy osobnego parsera rzeczownika. Jeśli występuje alternacja lub wyjątek, zapisujemy ją w `Stems`/`Overrides` i testujemy lokalnie.

## E. Kontrakt klasy morfologicznej

```go
MorphClass{
    ID: "en:verb.regular",
    POS: Verb,
    Cells: {
        Base:     SelectStem("base"),
        Present3: SelectStem("base").ApplySpelling("3sg"),
        Past:     SelectStem("base").ApplySpelling("past"),
        Participle: SelectStem("base").ApplySpelling("ing"),
    },
}
```

Powyższe API jest ideą, nie obietnicą identycznych symboli. `ApplySpelling` jest **typowaną operacją z językową regułą**, nie callbackiem wskazującym konkretny czasownik. Ogólne zachowanie obejmuje np. `work → works/worked/working`, `study → studies/studied/studying`; nieregularne `go → went` ma jawne zastąpienie komórki.

Dla PL analogicznie `pl:noun.feminine.a` oraz `pl:verb.present.am-asz.a...`, z zapisanym wyborem tematów i oboczną realizacją `książka → książce`. Klasy i reguły są wielokrotnego użytku. Jedyny nowy leksem nie może generować specjalnej procedury tylko dla swojej nazwy.

**Kompletny system fleksji nie musi mieć skończonej tabeli wszystkich słów.** Musi umieć wyraźnie powiedzieć, że dla danego leksemu brakuje tematu, wyjątkowej komórki albo jednoznacznej klasy.

## F. Kontrakt walencji i konstrukcji

```text
Frame `see(agent:Entity, patient:Entity)`:
  PL  agent -> NP[nominative], patient -> NP[accusative]
  PL+Not patient -> NP[genitive] jeśli licencjonowana rekcja tego predykatu
  EN  agent -> NP[subject], patient -> NP[object]

Construction `en:declarative.transitive.present`:
  [NP:agent] [VP:predicate+agreement] [NP:patient] [.]
Construction `en:negative.transitive.present`:
  [NP:agent] [DO/DOES:agreed] [not] [VERB:base] [NP:patient] [.]
Construction `pl:negative.transitive.present`:
  [NP:agent(optional only when licensed)] [nie] [VERB:finite] [NP:patient(case licensed)] [.]
```

Te trzy konstrukcje **nie** wymieniają po kolei `dog/cat/teacher/book`. Konkretne słowa, aspekty i przypadki pochodzą z zatwierdzonych wiązań. Kontrakt może jawnie odrzucić podmiot niedopasowany do formy czasownika. Licencja `analyze` lub `realize` jest odrębna i weryfikowana statycznie.

### Plan NP/VP a licencja słowa

Najpierw wybieramy dopuszczony leksem, dopiero potem obliczamy jego odmianę, rodzajnik, wymowę i założenia. Dzięki temu dwa synonimy o różnych właściwościach morfologicznych nie przejmują swoich cech nawzajem. To usuwa jeden z problemów v1 bez poświęcania bezpieczeństwa.

## G. Umowa „nie dopisujemy kodu dla nowego słowa”

W ramach znanej klasy i znanej ramy nowy leksem jest podłączony przez deklarację:

```go
b.Lexeme(Lexeme{Lemma: "lampa", MorphClass: "pl:noun.feminine.a", ...})
b.Sense(Sense{Lexeme: "pl:lampa/n", Concept: "core:lamp", ...})
```

**Żadna zmiana** w `internal/engine`, `internal/grammar`, `internal/morph` nie jest wtedy dopuszczalna. Jeśli nowa klasa wymaga nowej operacji, jest to osobna, mierzalna zmiana ogólnego algorytmu; nie wyjątek „dla słowa lampa”.

## H. Niezależność testów od wejścia

Dla każdego przypadku A1 ręcznie piszemy w Go **semantyczne wymagania** z użyciem publicznych konstruktorów `sem`, zamiast:

```go
expected := translator.Analyze(Polish, tc.PL) // NIE JEST ORACLE
```

Poprawne jest:

```go
expected := sem.TestMeaning(... /* ręcznie określony graph z rolami */)
actualPL := engine.Analyze(Polish, tc.PL)
actualEN := engine.Analyze(English, tc.EN)
assertSemanticCompatibility(actualPL, expected)
assertSemanticCompatibility(actualEN, expected)
```

Test kompatybilności musi odróżniać brak informacji PL o rodzajniku od niedopuszczalnej zmiany znaczenia. Sam status `ok` albo obecność tekstu nie jest zaliczeniem.
