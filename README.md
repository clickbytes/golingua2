# GoLingua 2

Proceduralny prototyp tłumacza **PL ↔ jeden typowany Semantic IR ↔ EN** w Go.

**Stan: M0–M8 zweryfikowane dla zamkniętego zestawu A1.** Wszystkie **100/100** ręcznie dobranych par przechodzi sześć niezależnych ścieżek odbioru: **600/600** kontroli semantyki i tekstu. Wynik dotyczy dokładnie wybranego podzbioru A1, nie certyfikacji CEFR, pełnego translatora języka naturalnego ani akceptacji lingwistycznej właściciela.

Nie zapisujemy zdań w danych runtime ani JSON. Wszystkie zdania testowe są w `tests/a1/a1_cases_test.go`; osobne znaczenia zostały napisane ręcznie w Go, z odrębnymi warunkami dla źródeł EN i PL.

## Zrealizowany zakres A1

- M7: czas przeszły i przyszły jako pola zdarzenia w IR z jawnym odniesieniem `previous_day/next_day`; przeszłość `slept/ran/read` i polskie formy rodzaju, `will + base` i `będzie + forma -ł/-ła/-ło` lub `będę + infinitive`. Bezosobowe EN `it` nie jest dodatkowym referentem w IR.
- M6: jeden typowany `Ask` z pytaniami polar, who, what, where i how-many. Pytania wskazują dokładną lukę roli lub ilości i nie zgadują odpowiedzi.
- M5: semantyczna liczebność 1–5 i deklaratywne licencje liczebników, plural/genitive, count/mass oraz ograniczone zdania egzystencjalne.
- M4: typowane relacje przestrzenne z dokładnymi rolami `located/site`, klasą miejsca, przypadkiem i językowym przyimkiem. `Kot jest w domu.` / `The cat is in the house.`, `Uczeń jest w szkole.` / `The student is at school.`, `Ptak jest na drzewie.` / `The bird is in the tree.`. Multiword `in front of` / `next to` to licencjonowane adpozycje, nie translacja stringów.

- M3: negacja zakresu zdarzenia `Not(Event)`, polskie `nie` + dopełniacz pod zadeklarowaną ramą, angielskie `do/does not + base`, z pozytywnym i negatywnym przebiegiem przez ten sam IR.
- M2: posiadanie „Mam książkę.” / „I have a book.”, ramy przechodnie SVO i nieprzechodnie SV; celownik „pomaga dziecku” i liczba nieokreślona w `drzwi`.
- Identyfikacja „To jest kot.” ↔ „This is a cat.” i dziewięć innych klas rzeczowników.
- Opis „Ten kot jest czarny.” ↔ „This cat is black.” i dziewięć innych przypisań cech.
- **Jeden** językowo neutralny IR: osobne definicje konceptów, wystąpienia referentów, predykaty, role i wskazanie bliskie.
- Kompilator: deklaratywne leksemy i sensy, klasy odmiany PL/EN, dokładne ramy walencji oraz wzorce z osobnymi uprawnieniami analizy/realizacji.
- Produkcyjna analiza używa indeksu odwrotnego form, generowanie używa proceduralnych operacji na tematach i końcówkach, bez dopasowania pełnych zdań do tablicy 100 par.
- Polskie `to/ten/ta` nie przesądza samo o angielskim `this`: w testach M1 występuje jawny kontekst bliskiego wskazania (`--near`). Angielskie `this` koduje bliskie wskazanie bez dodatkowej opcji.

## Wykonanie

```sh
go test ./... -count=1
go test -race ./... -count=1
go vet ./...
go build ./cmd/golingua2

go run ./cmd/golingua2 translate --from pl --to en --text 'To jest kot.' --near
go run ./cmd/golingua2 translate --from en --to pl --text 'This book is new.'
go run ./cmd/golingua2 analyze --from pl --text 'Ta książka jest nowa.' --near --json
go run ./cmd/golingua2 translate --from pl --to en --text 'Lekarz pomaga dziecku.'
go run ./cmd/golingua2 translate --from en --to pl --text 'You have a dog.' --addressee singular
```

Przykłady `M1–M7` podlegają pełnym testom analizy źródła, niezależnego IR→PL/EN i integracji PL→EN/EN→PL. `tests/a1/a1_cases_test.go` pozostaje **jedynym katalogiem 100 zdań**, zapisanym w Go — nie mamy pliku JSON ani ukrytych szablonów tłumaczeń.

## Architektura i granice

- [Całościowa architektura i 100 scenariuszy](docs/ARCHITECTURE.md)
- [Jeden IR i kontrakty języków](docs/IR-AND-CONTRACTS.md)
- [Plan M0–M8](docs/ROADMAP.md)
- [Akceptacja i testy](docs/ACCEPTANCE.md)
- [Zrealizowany zakres i ograniczenia](docs/IMPLEMENTATION-STATUS.md)
- [Zasady pracy](AGENTS.md)

**Wszystkie 100 przypadków z pierwszego katalogu jest zaimplementowanych i automatycznie odebranych.** Zakres pozostałych możliwych zdań nie jest ograniczony do tych tekstów, ale nie wolno wnioskować, że obsługiwany jest cały poziom A1. Ograniczenia poniżej obowiązują poza wymienionymi konstrukcjami: nie ma zagnieżdżonych klauzul, ogólnej kwantyfikacji, dowolnego czasu/aspektu, pełnej polskiej odmiany ani dowolnych pytań. Akceptacja językowa właściciela pozostaje `NOT EVALUATED`.

## Zakres i pochodzenie wariantów M2

- Leksemy EN/PL są danymi. Dwa rodzaje ramek predykatów (`state` i `event`) korzystają z jednego `sem.Meaning`, a sens czasownika musi licencjonować dokładną ramę składniową.
- `a/an`, `the` i brak polskiego rodzajnika nie są równoważne: źródłowa określoność występuje w IR, a docelowe wybory i straty mają adnotacje przypięte do konkretnego referenta.
- `you` nie ma angielskiego wykładnika liczby. `--addressee singular|plural` ją określa; bez tej wartości analiza jawnie zachowuje dwa warianty. `they` nie określa polskiego rodzaju osobowego, dlatego obie realizacje pozostają alternatywami.
- Polski leksykalny plural `drzwi` nie rozstrzyga liczby przedmiotów. EN zachowuje warianty `door` i `doors` z jawnym założeniem liczebności.
- **Nie ma dopisywania pełnych zdań do silnika.** Próbka 100 znajduje się tylko w kodzie testowym, a nowo zadany czasownik `obserwować / observe` przechodzi w nowej parze bez zmian runtime.
- **Granica historycznej dostawy M2:** sama M2 nie obejmowała negacji, czasu, PP ani pytań. Kolejne M3–M7 wdrażają ich ograniczone konstrukcje; koordynacja pozostaje poza zakresem A1.

## Dostawa M3 — negacja

- `sem.Scope{Operator: Not, Content: eventID}` jest jedyną reprezentacją zakresu negacji; nie dodano nowej semantyki równoległej do `sem.Meaning`.
- Negacja jest licencjonowaną konstrukcją ramy, a nie osobnym szablonem każdego zdania. Źródłowy `not/nie`, angielski auxiliary `do/does`, forma bazowa czasownika i polska rekcja pod negacją są walidowane przez pakiet języka.
- `Nie mam samochodu.` i `Nie mamy wody.` używają tej samej deklaracji pomijania podmiotu co pozytywne `Mam/Mamy`; nie wnioskujemy osoby z brakującego rzeczownika.
- Kolejnych 10 scenariuszy ma sześć niezależnych ścieżek; testy kontrastowe sprawdzają złe `do/does`, nieodmieniony czasownik z `does`, polski błędny biernik oraz utratę operatora z grafu. Przykładowe `obserwować/observe` w negacji działa po dodaniu danych, nie dodatkowego `switch`.
- Negacja `Not(Event)` jest nadal ograniczona do tych prostych zdarzeń i licencjonowanych konstrukcji; M7 dodaje wybrane przeszłe/przyszłe zdania pozytywne, nie pełną kompozycję negacji z dowolnym czasem i scope kwantyfikacji.

## Dostawa M4 — relacje przestrzenne

- Siedem neutralnych relacji i dokładne role `located/site` w tym samym `sem.Meaning` (stan relacyjny, nie zdarzenie). Kompilator odróżnia `StateRelation` od czasownikowego Event.
- Deklaratywne, językowe `LocationFrame` wybierają `w/na/pod/przed/za/obok` i ich odpowiedniki EN wraz z przypadkiem i klasą miejsca. EN `at school` ma licencję na bare institutional site; PL `na drzewie` i EN `in the tree` dzielą neutralne `at_site` z klasą drzewa.
- Nowe rzeczowniki `kuchnia/kitchen`, `drzewo/tree`, `pokój/room` oraz komórki morfologii przypadków, nie kopie pełnych zdań. Dodatkowe kombinacje `Kot jest pod stołem.`, `Książka jest za domem.` przechodzą bez modyfikacji parsera.
- 10 nowych przykładów A1 (071–080) ma sześć niezależnie testowanych ścieżek każdy. Kontrole sprawdzają zamianę uczestników, różne relacje, zły przypadek, brak przyimka, złą zgodę, dodatkowe człony i fałszywe licencje pakietu. `tests/a1/a1_cases_test.go` pozostaje nietknięty.
- **Nie obsługujemy dowolnego PP** ani pytania `where`, wielu miejsc jednocześnie, wszystkich przyimków, ruchu i relacji 3D. Testy ograniczonej konstrukcji nie są certyfikacją pełnego A1; przyjęcie językowe właściciela oddzielnie.

## Dostawa M5–M8 — ilość, pytania, czas i stabilizacja

- M5: referent `Cardinality` i `Existential` są w tym samym IR; liczba semantyczna nie jest gramatyczną liczbą wyrazu. Numeral morphology i reżim plural/genitive po polsku są deklaratywne; akceptacja A1-081–090.
- M6: `Ask` z rodzajem pytania i dokładną luką argumentu lub liczbą; licencjonowane interrogatywy PL/EN i odpowiednia zgoda oraz do-support; akceptacja A1-061–070.
- M7: pole czasu na predykacie `Event` z wymaganym `previous_day/next_day`, językowo różne konstrukcje czasu, zjawisko pogodowe z gramatycznym `it` i właściwym referentem deszczu, a nie sztucznym obiektem; akceptacja A1-091–100.
- M8: oddzielna kontrola wydajności ciepłego procesu `BenchmarkA1TranslationRoundTrip`, pełna macierz 100×6 i kontrasty niepoprawnej gramatyki. Pomiary procesora nie są pomiarem zużycia energii.

**Znane ograniczenia:** tylko wymienione czasowe formy i konstrukcje (`wczoraj/jutro`, nie dowolna data), ograniczone liczby 1–5 i klasa rzeczowników, pytania o wskazane role, brak pełnej deklinacji, aspect/perspektywy, dialogu, pełnej składni mowy potocznej, koordynacji i rozumowania z wiedzy o świecie. Proceduralne klasy wyrażają tylko jawnie licencjonowane cechy; nie zgadują brakujących tematów i form.
