# M0 + M1 + M2 — potwierdzony zakres GoLingua 2

## Co działa

- 100 par w jednej tabeli `tests/a1/a1_cases_test.go`, 10 grup po 10, bez JSON ze zdaniami;
- typowany katalog CORE i jeden graf `sem.Meaning`, lokalne ID wystąpień, walidacja sygnatur ról, porównanie niezależne od kolejności i nazw węzłów;
- jeden kompilator dwóch pakietów, z indeksem leksykalnym, osobnymi Sense/Frame/Construction/Direction i niemutowalnymi planami morfologii;
- proceduralne klasy: podstawowe polskie rodzaje rzeczownika, przymiotnik odmieniający się przez rodzaj, zaimek wskazujący, polska kopula; angielski rzeczownik regularny, przymiotnik nieodmienny, `a/an` z jawnej cechy fonetycznej, `this` i `is`;
- 50 niezależnie opisanych w Go znaczeń M1 i M2 (10 identyfikacji, 10 cech) sprawdzonych przez `PL→IR`, `EN→IR`, niezależne `IR→PL`, `IR→EN`, `PL→EN`, `EN→PL`;
- testy kompozycji trzech dodatkowych par, które nie znajdują się w próbce 100, oraz test nowego leksemu `lampa/lamp` z niestandardowym ID (bez modyfikacji parsera i generatora);
- odmowy niewłaściwego `a/an`, zgody demonstrative/adjective/copula, błędnej semantyki, niekompletnych ról i nieautoryzowanego wariantu konstrukcji;
- dwa języki kompilowane jednorazowo do indeksów wyszukiwania form i sensów, z jawnym limitem kroków analizy i liczby kandydatów.

## Co jeszcze NIE działa

- Produkcyjne zdania testowe 051–100 oraz brakujące 300 ścieżek semantycznych;
- poprawna, kompletna deklinacja i koniugacja całego polskiego i angielskiego (M1 zawiera tylko potrzebne, sprawdzone komórki; nie należy zgadywać pozostałych);
- pełna kompozycja NP/VP/PP/Clause poza obsługiwanymi ramami A1 SVO/SV, czasy inne niż present, aspekt, negacja, lokalizacja, liczebniki i pytania;
- alternatywy leksykalne o różnych wymaganiach, pełna semantyka określoności i kwantyfikacji;
- snapshot, szerokie benchmarki, weryfikacja energochłonności i trzeci język.

## Ustalenia architektoniczne

1. Zmienne kontekstowe nie są odczytywane z tekstu testu. Bliskie wskazanie w PL jest jawną przesłanką testową. Bez niej tłumaczenie na EN nie twierdzi, że bliskość została rozpoznana.
2. Oryginał zdania jest analizowany przy użyciu indeksu form, walencji i wzorca, a nie `switch` od identyfikatora przypadku lub pełnej powierzchni.
3. Typowane formy pochodzą z operacji na leksemach i klasach (np. polski `książk- + ę`, angielski `apple + s`). Kompilacja wymaga wszystkich zadeklarowanych tematów, odmawia brakujących komórek i przechowuje indeks do analizy odwrotnej.
4. Każdy test M1 używa **oddzielnie napisanego** grafu o innych lokalnych ID niż parser. Przyjęta implementacja nie ma dostępu do testowego oracle.
5. `t.Log` zbierający 100 × 2 kierunki jest **obserwacyjny** i nie zalicza niezaimplementowanych scenariuszy.

## Odbiór i interpretacja wyniku

Wykonane kontrole: `go test ./... -count=1`, `go test -race ./... -count=1`, `go vet ./...`, `go build ./cmd/golingua2`, `--help`, `--version`, rzeczywiste CLI z identyfikacją, cechą, nieznanym wskazaniem i błędną zgodą.

**Wymaganie językowe: 50/100 scenariuszy × 6 = 300/600 ścieżek.** Testy nowego słowa, morfologii, odwracania i kompilerów są dodatkowe; nie dolicza się ich do 600. Akceptacja właściciela: NOT EVALUATED.

## Dostawa M2 (A1-021–050)

- `Event` jest drugą postacią predykacji we **wspólnym** IR, nie nowym IR. Ma dokładne role agent/patient, holder/theme, agent/beneficiary, czas present i referent-local definiteness/person/number.
- Sense czasownika licencjonuje konkretną kompilowaną ramę. Wzorce zdania obsługują SVO, SV i polski podmiot pominięty przy `mieć`; generator i parser używają tego samego pakietu.
- EN: wspólne present-s, present-y oraz `have → has` jako punktowy wyjątek; PL: wspólne końcówki 3sg -a/-e/-i, proceduralne 6 form `mieć`, przypadek dopełnienia i celownik `dziecku`.
- 30 niezależnych par wymaga 180 ścieżek produkcyjnego silnika; wszystkie są sprawdzane w `TestM2SixIndependentFlows`. Przykłady PL i EN zachowują **różne** źródłowe informacje o określoności; ich IR nie jest fałszywie utożsamiany. Przekład docelowy raportuje przypięte do referenta założenia i straty.
- Negatywne testy obejmują niepoprawne `has/have`, `a/an`, brak rodzajnika angielskiej policzalnej liczby pojedynczej, błędną rekcję, niezgodną osobę/liczbę, zmienione role i powtórzone koncepty. Rozszerzenie pakietu o `predicate.observe` i regularny czasownik obu języków nie wymaga nowego parsera ani konstrukcji.
- Wynik obserwacyjny: 50/100 PL→EN oraz 50/100 EN→PL; testy 001–050 mają pełny odbiór 300/600, 051–100 nie są zatwierdzone. Nie zmieniono żadnego zdania z katalogu 100.

## Znane ograniczenia do następnej dostawy

- Obsługiwane nominalne przypadki i czasownikowe komórki to *wyłącznie* wymagany wycinek, nie pełna deklinacja lub koniugacja PL/EN. Wspólne klasy odmiany nie uprawniają do zgadywania form bez leksykalnego tematu.
- `OK` w API może zawierać `Losses`, gdy PL nie koduje rozróżnienia rodzajnika EN; pełna ocena musi czytać adnotacje. Formalne statusy kandydatów/etapów i rozszerzalne planowanie liczby będą dopracowane osobno.
- Nie ma jeszcze negacji, modalności ani skończonego czasu przeszłego/przyszłego w zdaniach M2.
