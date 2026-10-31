# GoLingua 2 — kompletna architektura docelowa dla pierwszych 100 scenariuszy A1

> **Implementacja M0–M8:** ukończona dla 100 par w `tests/a1/a1_cases_test.go`. Wszystkie 600 niezależnych przepływów zaliczone, nie oznacza to pełnego języka A1. Pozostałe rozdziały przedstawiają projekt i kontrakty.

**Wersja projektu:** RFC-001 / DRAFT 2026-10-31. **Język implementacji:** Go. **Języki startowe:** polski (PL) i angielski (EN). **Docelowe repozytorium:** `clickbytes/golingua2`. **Status:** dokument projektowy, nie deklaracja wykonanej implementacji.

## 0. Cel, zakres i decyzje wejściowe

Budujemy **wspólny, kompilowany silnik regułowy**, który tłumaczy:

```text
PL —(analiza z pakietu PL)→ typowane IR —(realizacja z pakietu EN)→ EN
EN —(analiza z pakietu EN)→ typowane IR —(realizacja z pakietu PL)→ PL
```

**Pierwszy cel: 100 ręcznie napisanych scenariuszy A1**, zapisanych w Go w `tests/a1/a1_cases_test.go`; każdy ma przykładowy tekst PL i EN. Oznacza to **200 kierunków integracji**, a nie 100 angielskich + 100 niezwiązanych polskich przykładów. Każda para reprezentuje ten sam zakres intencji/znaczenia, choć oba zdania mogą kodować nieco inne informacje; test musi jawnie dopuścić lub odnotować utratę/założenie. Nie przypisujemy tej próbce formalnej certyfikacji poziomu CEFR.

**Nie zapisujemy zdań testowych w JSON, YAML ani plikach językowych.** Na starcie wszystkie deklaracje reguł i leksykonów również mogą być konstruktorami Go. Ich kompilator musi jednak czytać *dane i typowane reguły*, a nie wykonywać dowolny kod „per zdanie”. Jeżeli później dodamy format zewnętrzny dla słowników, ten format nie będzie drugim źródłem reguł.

### Główne niezmienniki

1. Jeden autorytatywny IR semantyczny; żaden drugi DSE/Meaning i brak ukrytej translacji przez inną reprezentację.
2. `concept` ≠ `lexeme` ≠ `sense` ≠ `morph-class` ≠ `valency-frame` ≠ `construction` ≠ `procedure` ≠ `semantic occurrence`.
3. Wszystkie odmiany PL/EN powstają z **klas, tematów, operacji i jawnych wyjątków**, a nie kompletnych tabel przypisanych do konkretnego leksemu (poza wymaganymi supletivami).
4. Rola semantyczna nie przesądza o polskim przypadku ani pozycji w zdaniu. O tym decyduje kontrakt językowy.
5. Nowy leksem należący do już obsługiwanej klasy wymaga wyłącznie deklaracji danych; **nie powstaje nowy `switch` z nazwą leksemu/sensu**.
6. Parser i generator korzystają z tej samej zatwierdzonej deklaracji konstrukcji; jeżeli licencja jest kierunkowa, rozróżnienie jest jawne.
7. Każdy wynik ma świadectwo pokrycia całego znaczenia: uczestników, predykatów, operatorów, czasu, referencji i zakresów.
8. Nierozstrzygnięte znaczenie i wybory języka celu nie są zamieniane w domyślną pewność; alternatywy mają jawne założenia/straty.
9. Indeksy i reguły są kompilowane przed żądaniem; wykonanie ma limity pracy, stabilną kolejność i nie wymaga GPU.
10. Testy mają niezależne semantyczne wzorce przygotowane w Go i nie są autogenerowane z outputu parsera ani generatora.

## 1. Przekrój funkcjonalny A1 — 10 grup po 10

| Numery | Rodzina | Co musi naprawdę działać | Od której dostawy |
| --- | --- | --- | --- |
| 001–010 | Identyfikacja | „To jest…”, `this`, `a/an`, rzeczowniki, kontekst wskazania | M1 |
| 011–020 | Cechy | `be/być`, zaimki wskazujące, zgoda przymiotników | M1 |
| 021–030 | Posiadanie | `have/mieć`, ja/ty/on/ona/my/oni, osoba/liczba, count/mass | M2 |
| 031–040 | Przechodnie | actor/theme, leksykalna walencja i przypadki | M2 |
| 041–050 | Nieprzechodnie | brak fikcyjnego dopełnienia, czas teraźniejszy, liczba/osoba | M2 |
| 051–060 | Negacja | zakres `Not`, EN do-support, PL rządzenie pod negacją | M3 |
| 061–070 | Pytania | polar, who, what, where, how-many, zmienne typowane | M6 |
| 071–080 | Lokalizacja | relacje w/on/under/at/behind/near, polskie przypadki PP | M4 |
| 081–090 | Ilość | 1–5, liczebność semantyczna, zgoda, egzistencja, 5+ dopełniacz | M5 |
| 091–100 | Czas | wczoraj/jutro, podstawowa przeszłość i peryfrastyczna przyszłość | M7 |

Kolejność katalogu ułatwia czytanie. **Kolejność implementacji to M1→M2→M3→M4→M5→M6→M7**, więc pytania korzystają z działających wcześniej relacji i ilości. Każda grupa ma własne znaczenia, ale kompozycja korzysta ze wspólnych procedur. Pełna lista 100 par jest wyłącznie w pliku Go.

### Ważne ograniczenie oceny tekstów

`PL` i `EN` w tabeli są **przykładowymi poprawnymi powierzchniami**, nie jedynymi dozwolonymi odpowiedziami. Polski bez rodzajnika często nie przesądza `a/the`, a angielski `you` nie przesądza liczby. `drzwi` są gramatycznie mnogie bez określenia liczby drzwi. Polskie `to/ten/ta` i angielskie `this/that` mogą wymagać kontekstu deiktycznego. Nie wolno wymuszać jednej wersji EN wyłącznie dlatego, że wpisano ją jako przykładowe pole `EN`.

W przyszłym teście odbiorczym każdy przypadek otrzyma **niezależnie napisany konstruktor oczekiwanego IR**, politykę otwartych informacji i dozwolonych alternatyw oraz testy kontrastowe. Te dane również będą w Go, nie w JSON.

## 2. Struktura kodu i reguły zależności

```text
golingua2/
├── go.mod
├── cmd/golingua2/
│   └── main.go                  # composition root i CLI, bez gramatyki
├── internal/
│   ├── core/                    # typowane ID, sorty, koncepty i sygnatury
│   ├── sem/                     # JEDEN IR, walidacja, porównanie, scope
│   ├── context/                 # speaker/addressee, deixis, reference time
│   ├── lexicon/                 # leksemy, sensy, wiązania z CORE
│   ├── features/                # cechy językowe i deterministyczna unifikacja
│   ├── morph/                   # kompilator klas i algorytmów; generate/inverse
│   ├── grammar/                 # konstrukcje NP/VP/PP/Clause/Ask; compiler
│   ├── compile/                 # walidacja i zamrożenie pakietu języka
│   ├── engine/                  # Analyze, Realize, Translate; rozgałęzienia
│   └── status/                  # wyniki, diagnostyki, budżety i ślad decyzji
├── languages/
│   ├── pl/                      # deklaracje i polskie reguły fonologii/ortografii
│   │   ├── register.go
│   │   ├── features.go
│   │   ├── lexemes.go
│   │   ├── morphology.go
│   │   ├── frames.go
│   │   └── constructions.go
│   └── en/                      # analogiczny zestaw EN
├── tests/
│   ├── a1/
│   │   ├── a1_cases_test.go      # dokładnie 100 par, W GO
│   │   ├── a1_meanings_test.go   # przyszłe niezależne IR, W GO
│   │   └── a1_contract_test.go   # przyszłe prawdziwe sześć przepływów
│   ├── morphology/             # niezależne paradygmaty, round-trip form
│   ├── compiler/               # walidacja, licencje, brak duplikatów
│   └── metamorphic/            # zamiana ról, znana/nieznana cecha, nowe leksemy
└── docs/
    ├── ARCHITECTURE.md
    ├── IR-AND-CONTRACTS.md
    ├── ACCEPTANCE.md
    └── ROADMAP.md
```

**Reguły importów:** `core` nie importuje języków. `sem` importuje tylko typowane ID/sorty i stabilne definicje; nie importuje parsera. `morph` nie zna zdań ani znaczeń konceptów. `grammar` zna typowane sygnatury i publiczne API morfologii, lecz nie importuje `languages/pl` ani `languages/en`. `engine` operuje na skompilowanych pakietach jako danych. `languages/{pl,en}` deklarują pakiety, nie importują `engine`. `cmd` scala składniki. Nie ma pakietu `translation/pl-en` ani `en-pl` — kombinacje są produktem dwóch dowolnych pakietów.

## 3. Dokładny model dziedziny

- **Concept**: językowo neutralny symbol (`cat`, `see`, `location_in`), sort i sygnatura ról. Koncept nie jest konkretnym kotem ani słowem „kot”.
- **Lexeme**: jednostka językowa, np. `pl:kot/n` albo `en:read/v`; może mieć kilka sensów, własną klasę i tematy odmiany.
- **Sense**: znaczenie tego leksemu z jawnym powiązaniem do konceptu/ramy; homonimia i polisemiczność nie znikają przez zbieżność pisowni.
- **MorphClass**: wspólny algorytm odmiany; z niego i z tematów wyprowadza się formy, a nie osobne reguły dla jednego słowa.
- **ValencyFrame**: zestaw wymagalnych ról semantycznych i ich dopuszczalnych językowych realizacji.
- **Construction**: typowana, składana deklaracja liniowego/strukturalnego wzorca, zgody i operatorów (NP/VP/PP/clause/question).
- **Procedure**: skończone, bezpieczne operacje na formach oraz cechach, bez dynamicznych callbacków z zewnątrz.
- **MeaningOccurrence**: konkretne wystąpienie zdarzenia, stanu albo uczestnika w jednym IR, osobne od definicji konceptu.
- **Evidence**: skąd pochodzi odczyt; wskazuje dokładny node/owner/role, nie zastępuje samego znaczenia.

Jeden symbol w CORE może występować wiele razy w tym samym zdaniu. `Kot widzi kota` musi posiadać **dwa odrębne referenty**, nawet gdy oba mają `concept=cat`.

## 4. IR: jedna władza semantyczna

IR stanowi **typowany graf semantyczny na żądanie**. Minimalne typy: `Entity`, `Event`, `State`, `Relation`, `Proposition`, `Quantity`, `Location`, `Interval` i `QueryVariable`. Nie wprowadzamy arbitralnych etykiet jako dowodu poprawnego typu. IR przechowuje:

- węzły z lokalnymi identyfikatorami, konceptami i sortami,
- powiązania ról w konkretnych zdarzeniach, stanach i relacjach,
- referencje, kardynalność (nie mylić z liczbą gramatyczną), dane kontekstu,
- czas zdarzenia i punkt odniesienia (nie jako wymyślona data przy braku źródła),
- aspekty/iteratywność tylko jeśli mają znaczenie w zakresie obsługiwanej semantyki,
- operatory ze **strukturą zakresu** (`Not`, ewentualne `Exists`, `And`, `Ask`),
- osobne typowane pytania polarne i pytania o rolę, miejsce oraz ilość.

Nie wpisujemy do IR polskiego przypadka ani angielskiej reguły do-support. Nie zakładamy, że angielski rodzajnik jest jedynie dekoracją: źródłowe warunki referencyjne, jeśli są istotne, należą do IR lub jawnego kontekstu; sam wybór powierzchniowy i jego źródło pozostają w `Evidence`/`TargetChoice`.

### Przykład: „Pies nie widzi kota.”

```text
Meaning(
  root: Not(e1),
  e1: Event(concept=see, tense=present,
            agent=r1, patient=r2),
  r1: Entity(concept=dog, multiplicity=one,
             reference=unresolved),
  r2: Entity(concept=cat, multiplicity=one,
             reference=unresolved)
)
```

Przykład jest **notacją dokumentacyjną**, nie formatem JSON i nie kodem produktu. Rezultat EN może mieć alternatywy `A/The dog does not see a/the cat`, a każdy wariant powinien zachować oba referenty i jedno `Not`. Bez ukrytego przejścia do „kot nie widzi psa”.

### Niejednoznaczności

Przechowujemy rozłącznie: (a) różne odczytania źródłowe i ich Meaning, (b) warianty wymuszone przez nieznane cechy celu wraz z założeniami, (c) warianty szyku/stylu bez zmiany znaczenia. Cechy `unknown`, `absent`, `known`, `conflicting` mają odrębne znaczenie. Walidacja grafu sprawdza wszystkie ID/typy/role/scope; nie porównujemy tylko etykiet tekstowych.

Pytanie `Ile masz książek?` to `Ask(quantity, body=have(holder=addressee, object=books, number=?q))`. Nie tworzymy udawanej odpowiedzi i nie traktujemy zmiennej ilości jako nieobecnego argumentu.

## 5. Kontrakt pakietu językowego i kompilator

Na początku zdeklarowany w Go, typowany builder bez swobodnych funkcji wykonujących translację zdania:

```go
// Interfejs docelowy — szkic API, nie istniejąca implementacja.
func RegisterPL(b *lang.Builder) error {
    b.Feature(...)
    b.MorphClass(...)
    b.Lexeme(...)
    b.Sense(...)
    b.Frame(...)
    b.Construction(...)
    return b.Validate()
}
```

Dozwolone konstrukcje są licencjonowane przez dokładny zestaw `(language, sense, frame, construction, variant, direction, semantic-conditions)`. `analyze-only` **nie** przechodzi automatycznie do realizacji. Kompilator nie przepisuje reguł pod nazwą sensu i nie dopuszcza konstrukcji tylko dlatego, że zgadza się rodzina.

### Fazy kompilacji (`CompileLanguagePack`)

1. Sprawdzenie typów, schematów, nazw, referencji i wersji.
2. Weryfikacja sygnatur konceptów, sensów i ramek, oddzielenie gramatyki od semantyki.
3. Kompilacja klas morfologicznych i wyjątków, odwrotny indeks analizy form.
4. Kompilacja symbolicznych konstrukcji NP/VP/PP/Clause/Ask do planów parsowania i realizacji.
5. Wiązanie planów wyłącznie z uprawnionymi sensami/ramami/wariantami.
6. Weryfikacja dostępności potrzebnych form dla **zadeklarowanej** obsługi A1, nie dla wszystkich słów świata.
7. Utworzenie niezmiennych indeksów i stabilnego fingerprintu konfiguracji/słownika/reguł.

W runtime nie walidujemy wielokrotnie całych ramek i nie skanujemy wszystkich sensów dla każdej formy. `CompileLanguagePack` odrzuca niespójne definicje przed wejściem użytkownika; błędnego języka nie próbujemy „naprawiać” podczas tłumaczenia.

## 6. Proceduralna morfologia **obu** języków

### Wspólne API

```go
AnalyzeForm(langID, surface) -> []MorphReading
GenerateForm(langID, lexemeID, FeatureBundle) -> []GeneratedForm
```

Jedna definicja klasy generuje formy i odwracalne odczytania analizy; dopuszczamy synkretyzm i homografię. Analiza może korzystać z indeksu **wygenerowanego algorytmicznie** podczas kompilacji, nie z ręcznie wpisanej tabeli odmian każdego słowa. Inna implementacja (transduktor skończony/FST) jest możliwa dopiero gdy realne pomiary uzasadnią jej koszt.

### Polski

- Rzeczowniki: rodzaj, osobowość/żywotność tam, gdzie potrzebne, klasy deklinacji, mianownik/dopełniacz/celownik/biernik/narzędnik/miejscownik/wołacz i liczba; tematy oboczne i klasy supletivne.
- Przymiotniki i wskazujące: zgodność z rzeczownikiem i pozycją NP; nie przechowywać wszystkich końcówek osobno dla `czarny`, `czerwony` itd.
- Czasowniki: klasa koniugacji, tematy dla teraźniejszości/przeszłości, osoba/liczba/czas/rodzaj we właściwym czasie, aspekt leksykalny, zwrotność z poprawną interpretacją.
- Negacja może zmienić rządzenie argumentem (`czytam książkę` vs `nie czytam książki`), ale **to reguła konstrukcji z ramą**, nie reguła „odmień każdy rzeczownik na dopełniacz po słowie nie”.
- Liczebniki wymagają własnych reguł zgody i przypadków; `pięć książek` nie oznacza pięciu pojedynczych tekstów „książka”.

### Angielski

- Rzeczowniki: regularne `-s/-es`, `y→ies`, klasy nieregularne `children`, `men` itd.; count/mass należy do leksemu/sensu i do planu NP.
- Czasowniki: `work/works/worked/working`, zmiana `y→ies`, dublowanie spółgłoski tam, gdzie klasa to dopuszcza, `be/have/do/go` i inne nieregularności jako jawne dane.
- `a/an/the/bare` to konstrukcja NP wybierana w połączeniu z liczbą, count/mass i kontekstem referencji. `a/an` wybieramy na podstawie wymowy/klasy początku wyrazu, nie wyłącznie pierwszej litery (`an hour`, `a university`).
- `do/does/did + not + base`, pytania i `will + base` to procedury frazy czasownikowej oraz zdania, nie komórki odmiany rzeczownika ani osobne szablony pełnych zdań.

### Wyjątki

Preferujemy: (1) wspólną klasę, (2) jawny temat oboczny, (3) lokalną alternację, (4) pojedynczą wyjątkową formę/supletywę. Każdy wyjątek ma ID, zakres komórki i źródło. Nie zmuszamy wyjątków do sztucznego regularnego wzorca. Nie wybieramy klasy nowego słowa heurystycznie, gdy wejście nie wystarcza do jednoznacznego wyboru.

### Test rozszerzalności

`AddLexeme(lemma=skakać, morphClass=pl.verb.ać, stems=...)` i `AddLexeme(lemma=lamp, morphClass=en.noun.regular)` po kompilacji muszą działać w znanej rodzinie bez zmiany kodu `engine`, `grammar` lub konstrukcji specyficznej dla zdania. Testy także sprawdzają, że brak wystarczających tematów lub niemożliwa kombinacja cech daje jawną diagnostykę zamiast zmyślonej formy.

## 7. Składalna gramatyka zamiast pełnozdaniowych wzorców

Kategorie: `NP`, `AP`, `PP`, `VP`, `Clause`, `Question`. Dla każdej konstrukcji zapisujemy sloty, ich kategorie, liczbę wystąpień, role w ramie, agreement/feature constraints, kolejność, markery i jawny kierunek licencji. **Nie zapisujemy osobnych konstrukcji dla `a cat sees a dog`, `the cat sees a dog`, `a cat sees the dog` itd.**

### Rzeczownikowe NP

W EN NP planuje rodzajnik, liczbę, przymiotnik, rzeczownik, a następnie właściwą ortografię. W PL NP planuje wskazujący/przymiotnik/rzeczownik i uzgadnia przypadek, liczbę oraz rodzaj. NP widzi tylko lokalne wystąpienie referenta; cechy `cat` nie wyciekają do drugiego kota.

### Predykatowe VP

VP planuje czas, osobę/liczbę, negację, operator pomocniczy i formę leksemu. Polski podmiot może być jawny albo pominięty tylko przy dopuszczonym zakotwiczeniu osoby; angielski zwykle wymaga subject. Różnica jest deklaracją językową, nie faktem o świecie.

### Zdanie

Rama wyznacza `agent`/`patient`/`holder`/`location`, a językowe sloty mapują to na mianownik/biernik/dopełniacz/celownik/PP. Polski pozwala na niektóre warianty szyku, ale bez dowodu akceptacji nie emitujemy dowolnego OVS. Negacja jest semantycznym operatorem nad konkretną klauzulą; polska rekcja i angielski auxiliary są sposobami jego realizacji.

### Pytania

`AskPolar(P)` vs `AskRole(variable=Entity, role=agent, body=P)` vs `AskWhere(variable=Location, body=P)` vs `AskCount(variable=Quantity, body=P)`. PL `czy`, EN inversion/do-support i WH-fronting pochodzą z planu `Question` z dostępem do tej samej bazowej klauzuli. Nie tworzymy osobnych systemów pytań dla każdego słowa `who/what/where/how many`.

### Lokalizacja i liczebność

`na drzewie` i `in the tree` mają wspólne znaczenie relacji przestrzennej, ale **nie** regułę PL `na` → EN `on`. Liczba gramatyczna `drzwi` jest niezależna od tego, ile drzwi przedstawia referent. To samo dotyczy English pluralia tantum, jeśli rozszerzymy pakiet.

## 8. Dokładny algorytm żądania

### `Analyze(lang, text, ctx)`

1. Tokenizacja/normalizacja dla języka, bez usuwania informacji potrzebnej do odmiany i wieloznaczności.
2. Analiza morfologiczna: każde słowo → dopuszczone odczytania leksykalne z cechami, nie pojedynczy zgadywany leksem.
3. Parser konstrukcji: indeksowane warianty, składanie NP/VP/Clause; każda potencjalna interpretacja ma role oraz dowody form.
4. Kompozycja semantyczna przez dokładne wiązanie sensu/ramy i jawne operatory.
5. Walidacja typów i pełnego pokrycia wszystkich znaczących elementów źródła.
6. Deduplikacja identycznych interpretacji bez łączenia różnych referentów, zdarzeń ani dowodów.
7. Zwrócenie skończonego zestawu `Reading` z IR, kontekstem, diagnostyką i provenance.

### `Realize(lang, meaning, ctx)`

1. Walidacja całego IR i kontekstu; weryfikacja zgodności wszystkich wystąpień.
2. Selekcja dopuszczonych sense bindings i konstrukcji realizacji, a następnie **pokrycie całego grafu**.
3. Lokalny wybór leksemu i przeliczenie jego własnych właściwości NP/VP; żadnego kopiowania adnotacji między niezgodnymi synonimami.
4. Planowanie gramatyczne (role → przypadki, szyk, pomocnicze, negacja, pytanie, referencje).
5. Generacja odmiany z klas proceduralnych; łączenie fraz, ortografia i interpunkcja.
6. Oddzielne adnotacje warunkowych założeń/strat. Weryfikacja realizacji przez zgodność coverage.
7. Zwrócenie wariantów językowych, ograniczonych budżetem.

### `Translate(from, to, text, ctx)`

Uruchamia `Analyze(from)`, a następnie `Realize(to)` dla każdej **różnej interpretacji**, z budżetami i jawną relacją source reading → target choice → surface variant. **Nie ma funkcji `TranslatePLtoEN` z własną tabelą reguł**. Jeśli nie można przekazać operatora/roli, brak nie może zniknąć pod tekstem pochodzącym z innej interpretacji.

## 9. Publiczne API i CLI

Sugerowane API pakietu Go:

```go
// Typy opisowe proponowanego API, nie obecne symbole kodu.
Analyze(ctx context.Context, language LanguageID, input string, opts Options) (AnalysisResult, error)
Realize(ctx context.Context, language LanguageID, meaning Meaning, opts Options) (SurfaceResult, error)
Translate(ctx context.Context, from, to LanguageID, input string, opts Options) (TranslationResult, error)
Inflect(ctx context.Context, language LanguageID, lexeme LexemeID, features FeatureSet) (FormsResult, error)
AnalyzeForm(ctx context.Context, language LanguageID, surface string) (ReadingsResult, error)
```

Przykładowe polecenia (projekt, nie istniejący działający CLI):

```sh
golingua2 pack check --lang pl
golingua2 morph inflect --lang pl --lemma kot --case genitive --number singular
golingua2 morph analyze --lang en --form children
golingua2 analyze --lang pl --text 'Pies widzi kota.'
golingua2 realize --lang en --ir-file meaning.bin
golingua2 translate --from pl --to en --text 'Pies widzi kota.'
golingua2 capabilities --lang pl
```

Wynik programu zawiera `status`, `meanings`, `candidates`, `assumptions`, `losses`, `diagnostics`, identyfikatory planów i opcjonalnie `trace` wraz z fingerprintem pakietu. `--ir-file` dotyczy **wersjonowanego IR**, a nie zdania testowego; format binarny lub tekstowy jest osobną kwestią API, nie drugim źródłem semantyki.

## 10. Statusy i koszt wykonania

`OK`: kompletna, zatwierdzona interpretacja/realizacja. `Ambiguous`: istnieją różne nierozstrzygnięte znaczenia lub istotne warunki targetu; nie ukrywać ich. `Incomplete`: brakuje obowiązkowego dowodu/argumentu. `Unsupported`: żądany fragment jest poza deklarowanym zakresem pakietu. `Unknown`: nierozpoznana forma/sens/koncept. `Invalid`: niepoprawny IR lub kontrakt. `ResourceLimit`: przekroczony budżet/czas; już gotowe warianty mogą pozostać wyraźnie oznaczone jako częściowe.

Kandydat ma własny status. Agregat żądania nie powstaje przez proste `max(status)`: przykładowo jeden nieobsługiwany wybór leksykalny nie musi przekreślać innego poprawnego i kompletnego. Brak tłumaczenia dowolnego operatora nie jest `OK`.

Kompilator pakietu mierzy liczbę klas/cells/wnioskowań, wykrywa nielegalne cykle i ogranicza liczbę wariantów. Runtime ma konfigurowalne wąskie limity: tokeny, odczytania słowa, korzenie parsera, węzły IR, wybory lexeme, warianty targetu i czas CPU; ich wartości są **propozycją benchmarku**, nie cechą semantyczną. W pierwszym przekroju celujemy w 32 tokeny, 64 odczytania/typowe punkty rozgałęzienia, 256 częściowych parse, 64 kandydatów IR i 32 warianty powierzchniowe. Jeśli potrzebne są inne liczby, mierzymy źródło wzrostu przed ich podniesieniem.

Snapshot skompilowanego pakietu powinien być wersjonowany i fingerprintowany; zmiana klasy odmiany unieważnia poprzedni cache. Nie korzystamy z przestarzałego snapshotu po zmianie paradygmatów. Tłumacz nie potrzebuje GPU, ale korzyści energetyczne i wydajność wymagają oddzielnych pomiarów.

## 11. Testowanie i formalny warunek zaliczenia

**Dokładne testy 100 par są w Go**, a ich definicje nie są regułami translatora. `go test ./tests/a1` w dostarczonym starterze obecnie sprawdza **tylko kompletność tabeli**; NIE sprawdza tłumaczeń. Testy integracyjne powstają po zbudowaniu prawdziwego `engine` i niezależnych IR oracles.

Każdy przypadek wymaga docelowo sześciu ścieżek:

1. `PL → IR`: analiza własnego polskiego wejścia; role, referenci, cechy i operatory.
2. `EN → IR`: niezależna analiza wejścia angielskiego; semantyczna zgodność w dozwolonym zakresie.
3. `IR → PL`: realizacja ręcznie przygotowanego grafu, **bez parsowania PL lub EN**.
4. `IR → EN`: analogiczna niezależna realizacja po angielsku.
5. `PL → IR → EN`: rzeczywisty pipeline + adnotacje wyborów.
6. `EN → IR → PL`: rzeczywisty pipeline + adnotacje strat/założeń.

**100 × 6 = 600 wymaganych kontroli ścieżek**, oprócz testów kontrastowych i morfologicznych. Pełne 100 zaliczamy dopiero gdy wszystkie wymagania semantyczne i kompletność coverage są zachowane, a nie gdy jest jakikolwiek tekst.

### Kontrole negatywne i metamorficzne

- Zamiana agent/patient nie może zachować tego samego meaning.
- Zmiana `Not(P)` na `P` albo na `Not(Q)` musi dać rozróżnialny zakres.
- Powtórzony `cat` nie scala dwóch odrębnych referentów.
- `5 books` nie jest liczbą pojedynczą; `drzwi` nie wyznaczają kardynalności.
- `You` z nieznaną liczbą nie staje się samodzielnie `you[singular]`.
- Różnica `a/the` i `this/that` jest jawna albo zachowana przez dopuszczony kontekst.
- Błędne `The dogs runs` i `Psy biega` nie są poprawną leksykalną realizacją.
- „Jutro” i „wczoraj” odnoszą się do ustalonego zegara testowego, nigdy do rzeczywistego czasu uruchomienia.
- Dodanie niepowiązanych leksemów nie zmienia gotowego tłumaczenia, kolejności kandydatów ani statusów.
- Dodanie nowego leksemu do istniejącej klasy nie zmienia plików źródłowych `engine`/`grammar`; dodatkowy test sprawdza to na nowym słowie niewystępującym w korpusie.

Pełne kontrakty i statusy odbioru: [ACCEPTANCE.md](ACCEPTANCE.md). Każda poprawka ma kod źródłowy, bazowy/końcowy SHA, komendy testów, exit codes oraz fragmenty CLI. `NOT RUN` ≠ `PASS`.

## 12. Iteracje, demonstracja i definicja DONE

Odbiór rozbijamy na M0–M7 zgodnie z [ROADMAP.md](ROADMAP.md). Od początku istnieje lista 100, ale **nie zaliczamy braków jako PASS**. Techniczne CI może być zielone, kiedy nie ukończono A1 — raport zakresu i niezaliczonych przypadków jest osobny.

**MVP A1 DONE** oznacza: 100 istniejących wymagań, niezależny IR dla każdego, sześć ścieżek na przypadek, poprawność czasów/pytania/liczby/przypadka oraz jawne straty/założenia, kompletne testy kontrastowe, techniczne CI, żadnych per-zdaniowych switchy, brak ukrytego fallbacku i jasny raport wydajności na CPU. Wersjonowana dokumentacja wymienia czego poza tym MVP nie obsługujemy.

## 13. Nie implementować w pierwszym zakresie

Nie wprowadzamy LLM, uczenia w czasie żądania, zewnętrznego AI wymagającego akceleratorów, pełnego parsowania dowolnego tekstu, ogólnego rozumowania o świecie, dialogów, pełnej koordynacji, dopasowywania zdań po podobieństwie ani bezwarunkowego fallbacku. Nie przenosimy wprost starego runtime PL/EN, starych list sensów ani dwóch historycznych modeli semantycznych.

Do późniejszego rozważenia: trzecia miniaturowa demonstracyjna gramatyka, FST dla dużych słowników, szybki parser przez skompilowane transduktory, samodzielne wykonywanie małych operacji na IR. **Najpierw trzeba wykazać data-only extensibility i poprawny pierwszy zakres A1.**
