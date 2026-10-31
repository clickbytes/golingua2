# GoLingua 2

Proceduralny prototyp tłumacza **PL ↔ jeden typowany Semantic IR ↔ EN** w Go.

**Stan: M0–M4 zweryfikowane w ograniczonym zakresie A1.** 70 z 100 ręcznie dobranych par A1 przechodzi sześć niezależnych ścieżek odbioru, czyli **420/600 wymaganych ścieżek**. Dla pozostałych 30 scenariuszy nie deklarujemy obsługi ani PASS. Nie jest to certyfikacja CEFR ani pełny translator języka naturalnego.

## Pierwszy działający przekrój

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

Przykłady `M1–M4` podlegają pełnym testom analizy źródła, niezależnego IR→PL/EN i integracji PL→EN/EN→PL. `tests/a1/a1_cases_test.go` pozostaje **jedynym katalogiem 100 zdań**, zapisanym w Go — nie mamy pliku JSON ani ukrytych szablonów tłumaczeń.

## Architektura i granice

- [Całościowa architektura i 100 scenariuszy](docs/ARCHITECTURE.md)
- [Jeden IR i kontrakty języków](docs/IR-AND-CONTRACTS.md)
- [Plan M0–M8](docs/ROADMAP.md)
- [Akceptacja i testy](docs/ACCEPTANCE.md)
- [Zrealizowany zakres i ograniczenia](docs/IMPLEMENTATION-STATUS.md)
- [Zasady pracy](AGENTS.md)

**Pozostałe 30 par (ilość, pytania i czas) nie są zaimplementowane.** Wspólne klasy i kompilator są podstawą do ich dalszej realizacji, ale nie zastępują potrzebnych reguł gramatycznych i semantycznych. Kontrole techniczne nie oznaczają akceptacji językowej właściciela.

## Zakres i pochodzenie wariantów M2

- Leksemy EN/PL są danymi. Dwa rodzaje ramek predykatów (`state` i `event`) korzystają z jednego `sem.Meaning`, a sens czasownika musi licencjonować dokładną ramę składniową.
- `a/an`, `the` i brak polskiego rodzajnika nie są równoważne: źródłowa określoność występuje w IR, a docelowe wybory i straty mają adnotacje przypięte do konkretnego referenta.
- `you` nie ma angielskiego wykładnika liczby. `--addressee singular|plural` ją określa; bez tej wartości analiza jawnie zachowuje dwa warianty. `they` nie określa polskiego rodzaju osobowego, dlatego obie realizacje pozostają alternatywami.
- Polski leksykalny plural `drzwi` nie rozstrzyga liczby przedmiotów. EN zachowuje warianty `door` i `doors` z jawnym założeniem liczebności.
- **Nie ma dopisywania pełnych zdań do silnika.** Próbka 100 znajduje się tylko w kodzie testowym, a nowo zadany czasownik `obserwować / observe` przechodzi w nowej parze bez zmian runtime.
- **M2 nie obsługuje jeszcze** negacji, przeszłości, przyszłości, koordynacji, PP ani WH.

## Dostawa M3 — negacja

- `sem.Scope{Operator: Not, Content: eventID}` jest jedyną reprezentacją zakresu negacji; nie dodano nowej semantyki równoległej do `sem.Meaning`.
- Negacja jest licencjonowaną konstrukcją ramy, a nie osobnym szablonem każdego zdania. Źródłowy `not/nie`, angielski auxiliary `do/does`, forma bazowa czasownika i polska rekcja pod negacją są walidowane przez pakiet języka.
- `Nie mam samochodu.` i `Nie mamy wody.` używają tej samej deklaracji pomijania podmiotu co pozytywne `Mam/Mamy`; nie wnioskujemy osoby z brakującego rzeczownika.
- Kolejnych 10 scenariuszy ma sześć niezależnych ścieżek; testy kontrastowe sprawdzają złe `do/does`, nieodmieniony czasownik z `does`, polski błędny biernik oraz utratę operatora z grafu. Przykładowe `obserwować/observe` w negacji działa po dodaniu danych, nie dodatkowego `switch`.
- Zasięg pozostaje ograniczony do prostych pojedynczych zdarzeń w teraźniejszości. `Not(State)`, zagnieżdżona negacja, do-support w pytaniach, czas przeszły i scope kwantyfikacji nie są jeszcze zaimplementowane.

## Dostawa M4 — relacje przestrzenne

- Siedem neutralnych relacji i dokładne role `located/site` w tym samym `sem.Meaning` (stan relacyjny, nie zdarzenie). Kompilator odróżnia `StateRelation` od czasownikowego Event.
- Deklaratywne, językowe `LocationFrame` wybierają `w/na/pod/przed/za/obok` i ich odpowiedniki EN wraz z przypadkiem i klasą miejsca. EN `at school` ma licencję na bare institutional site; PL `na drzewie` i EN `in the tree` dzielą neutralne `at_site` z klasą drzewa.
- Nowe rzeczowniki `kuchnia/kitchen`, `drzewo/tree`, `pokój/room` oraz komórki morfologii przypadków, nie kopie pełnych zdań. Dodatkowe kombinacje `Kot jest pod stołem.`, `Książka jest za domem.` przechodzą bez modyfikacji parsera.
- 10 nowych przykładów A1 (071–080) ma sześć niezależnie testowanych ścieżek każdy. Kontrole sprawdzają zamianę uczestników, różne relacje, zły przypadek, brak przyimka, złą zgodę, dodatkowe człony i fałszywe licencje pakietu. `tests/a1/a1_cases_test.go` pozostaje nietknięty.
- **Nie obsługujemy dowolnego PP** ani pytania `where`, wielu miejsc jednocześnie, wszystkich przyimków, ruchu i relacji 3D. Testy ograniczonej konstrukcji nie są certyfikacją pełnego A1; przyjęcie językowe właściciela oddzielnie.
