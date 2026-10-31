# GoLingua 2 — odbiór i testy A1 w kodzie Go

**Status:** wymagania + rzeczywisty odbiór M0–M8 dla 100 przypadków w Go. Kompletne 100 zdań znajduje się wyłącznie w `tests/a1/a1_cases_test.go`; nie ma katalogu zdań w JSON.

## 1. Trzy niezależne warstwy testów

### I. Testy kontraktu i kompilatora

Złe `ConceptID`, nieznany `SenseID`, rama o niewłaściwej roli, niedeklarowany operator, błędny kierunek wariantu, niedostępny temat i sprzeczna klasa → przewidywalny błąd podczas kompilacji. Przy poprawnym pakiecie indeksy są niemutowalne i deterministyczne.

### II. Testy morfologii i kompozycji

Każdy regularny leksem ma zweryfikowane formy, a generator i analizator są względem nich zgodne: `GenerateForm` → `AnalyzeForm` zwraca co najmniej jego własny odczyt, bez utraty legalnej homografii. A1 wymaga w PL mianownika, dopełniacza, biernika, celownika, narzędnika i miejscownika w odpowiednim zakresie oraz form czasowników i przymiotników; nie oznacza to pełnego globalnego paradygmatu każdego polskiego słowa. EN wymaga form, pomocniczych i operacji ortograficznych potrzebnych dla tej próbki.

Kontrasty: błędny przypadek, 3sg `-s`, do-support, podmiot 1/2/3 osoby, `a/an` z faktyczną wymową, liczba obiektu i czas.

### III. Testy znaczenia i tłumaczenia

Każdy przypadek wymaga *niezależnego, napisanego ręcznie w Go* grafu/warunków znaczenia oraz jawnych wariantów i kontekstów, jeśli występują. Dla każdej z 100 par wykonujemy:

| Kontrola | Dane wejściowe | Kryterium |
| --- | --- | --- |
| `PL→IR` | źródłowe PL | znaczenie i wszystkie role zgodne z niezależnym wzorcem |
| `EN→IR` | źródłowe EN | analogicznie, z zachowaniem rozróżnień źródła |
| `IR→PL` | ręczne IR | brak źródła/parsera; tekst PL + zgoda/przypadki |
| `IR→EN` | ręczne IR | tekst EN + właściwe artykuły/operator |
| `PL→EN` | realny pipeline | każda wynikowa alternatywa zachowuje znaczenie i ujawnia założenia |
| `EN→PL` | realny pipeline | analogicznie, ze stratami źródłowych rozróżnień |

Razem **600 kontroli ścieżek** oraz dodatkowe odmowy. Każdy test zbiera `caseID`, `sourceSHA`, `packDigest`, `input`, `expectedMeaning`, `output`, `diagnostics` i dokładną ocenę. `PASS` tylko gdy spełniony cały kontrakt. `NOT RUN`, `UNSUPPORTED`, `INCOMPLETE`, `INVALID` i `RESOURCE_LIMIT` są widoczne, nigdy maskowane.

## 2. Wzorzec sześciu kontroli i rzeczywisty kod testowy

```go
// Docelowy wzorzec; symbole engine/sem zostaną określone po M0–M1.
func TestA1Acceptance(t *testing.T) {
    runtime := newProductionEngine(t) // zawsze prawdziwy, skompilowany pakiet
    for _, tc := range cases {
        t.Run(tc.ID, func(t *testing.T) {
            ctx := fixtureContext(tc)        // speaker, addressee, deixis, stały zegar
            want := independentMeaning(tc.ID) // ręcznie zbudowany w Go, nie przez parser
            checkAnalysis(t, runtime, "pl", tc.PL, ctx, want)
            checkAnalysis(t, runtime, "en", tc.EN, ctx, want)
            checkRealization(t, runtime, "pl", want, ctx)
            checkRealization(t, runtime, "en", want, ctx)
            checkIntegration(t, runtime, "pl", "en", tc.PL, ctx, want)
            checkIntegration(t, runtime, "en", "pl", tc.EN, ctx, want)
        })
    }
}
```

Rzeczywiste testy `TestM2SixIndependentFlows`, `TestM3SixIndependentFlows`, `TestM4SixIndependentFlows`, `TestM5SixIndependentFlows`, `TestM6SixIndependentFlows`, `TestM7SixIndependentFlows` oraz niezależne testy M1 sprawdzają komplet 100 przypadków. Powyższy przykład API pozostaje jedynie ilustracją. Odbiór techniczny 600/600 nie jest certyfikacją językową CEFR ani akceptacją właściciela.

## 3. Dopuszczone warianty i nieporównywanie samych stringów

**Testujemy semantykę, cechy i warunki, a dopiero potem tekst.** Użyteczna kontrola przykładowych powierzchni pozostaje, ale nie jest jedyną definicją poprawności.

- `Mam książkę.` dopuszcza `I have a book.` i `I have the book.` tylko wtedy, gdy odpowiedni wariant ma prawdziwie potrzebne założenie lub dowód referencji. EN jako *źródło* może wnosić dodatkowe rozróżnienie do IR.
- `Macie rower.` używa kontekstu wskazującego wielu adresatów; brak takiego kontekstu nie uzasadnia cichego `singular`.
- `Tata zamyka drzwi.` nie uprawnia do ustalenia, że chodzi o jeden `door`; wybór liczebności musi być jawny.
- `To jest kot.` i `This is a cat.` mają w teście jawne bliskie wskazanie. Bez niego polskie `to` nie koduje jednoznacznie angielskiego `this`.
- `Pies nie widzi kota.` i `Pies widzi kota.` mają różne korzenie; `Not` nie może zniknąć w EN do-support.
- `There are two cats` wymaga **dwóch jednostek** w semantyce i oddzielnych reguł gramatycznej liczby PL/EN.
- `Wczoraj padał deszcz.` nie tworzy w IR dodatkowego referenta „it” z angielskiej konstrukcji bezosobowej.
- `Jutro` i `wczoraj` liczone są względem **zamrożonego czasu** `2026-01-15T12:00:00Z` i kontekstu strefy; nigdy w zależności od dnia uruchomienia CI.

## 4. Testy generatywne: kontrakt zamiast zapamiętania 100 zdań

### Test „nowe słowo bez nowego kodu”

Kompilujemy tę samą klasę z nieużytym wcześniej, ręcznie zadeklarowanym rzeczownikiem i czasownikiem. Sprawdzamy jego formy i umiejętność wykorzystania w istniejącej konstrukcji, **bez modyfikacji kodu parsera/generatora**. To osobny test; jego słów nie doliczamy do 100.

### Test niezależności od niepowiązanych danych

Rozbudowa słownika o 100 niepowiązanych sensów nie może zmienić wcześniej przyjętego wyniku ani spowodować `resource_limit` w niezwiązanym zdaniu.

### Test zakresu operatorów i tożsamości

Zmieniamy `agent` z `patient`, dodajemy drugi identyczny koncept, negujemy niewłaściwy event, usuwamy required role i sprawdzamy odmowę lub różne znaczenia.

### Test odwrotności morfologii

Dla każdej wygenerowanej formy poprawny analizator zachowuje co najmniej odczyt użytego leksemu, komórki i cech. Nie wolno wymuszać tylko jednej analizy dla homografów. Dodatkowo pojedynczy wyjątek nie może nadpisać formy innego leksemu tej samej klasy.

## 5. Czego nie wolno zrobić, aby zdobyć 100/100

Nie wolno dopisywać specjalnego `switch` po `caseID` ani tekstach próbki, przepisywać expected do obecnego outputu, usuwać adnotacji, rozszerzać `AllowedTexts` na wszystkie teksty, deklarować `unsupported` jako pass, brać oczekiwanego IR z analizatora ani wprowadzać niejawnego fallbacku tłumaczeniowego poza IR.

## 6. Szczegółowe testy graniczne A1

- **Polski:** żeński, męski żywotny/nieżywotny i nijaki; kategorie `książka → książki` pod Not, `książka → książce` w zależnym kontekście, `pomaga dziecku` z celownikiem; plurale i numeral genitive `pięć książek`.
- **Angielski:** `is/are`, `has/have`, `do/does/did`, `there is/are`, `a/an` i `zero article`, regularna/nieregularna przeszłość `sleep/slept`, `read/read`, `fly/flew` w osobnym profilu.
- **Semantyka:** co najmniej dwa uczestniki tego samego ConceptID, odrębne osoby `he/she`, pro-drop PL i jawny mówca w EN, `this` ze wskazaniem, leksykalne pluralia tantum oraz liczebność obiektów.
- **Pytania:** `who` wiąże podmiot, `what` dopełnienie, `where` zmienną lokalizacji, `how many` ilość, pytanie polarne nie zamienia się w stwierdzenie.
- **Czas:** prawidłowy punkt odniesienia oraz wybory angielskiego prostego czasu przeszłego wobec polskiego aspektu.

## 7. Raportowanie i testy techniczne

Po każdej dostawie zapisujemy dokładne polecenia i status:

```sh
go fmt ./...
go vet ./...
go test ./tests/a1 -count=1 -v          # 100 przypadków × sześć przepływów
go test ./... -count=1                 # od M1 także testy modułów
go build ./cmd/golingua2
```

`cmd/golingua2` już istnieje. W kompletnym przebiegu 100/100 i 600/600 kontraktów przechodzą testy, natomiast poza zadeklarowanym zakresem każda odmowa pozostaje widoczna. Techniczne CI i akceptacja językowa właściciela są odrębnymi bramkami.
