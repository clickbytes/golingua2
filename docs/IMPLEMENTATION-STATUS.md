# M0 + M1 — potwierdzony zakres GoLingua 2

## Co działa

- 100 par w jednej tabeli `tests/a1/a1_cases_test.go`, 10 grup po 10, bez JSON ze zdaniami;
- typowany katalog CORE i jeden graf `sem.Meaning`, lokalne ID wystąpień, walidacja sygnatur ról, porównanie niezależne od kolejności i nazw węzłów;
- jeden kompilator dwóch pakietów, z indeksem leksykalnym, osobnymi Sense/Frame/Construction/Direction i niemutowalnymi planami morfologii;
- proceduralne klasy: podstawowe polskie rodzaje rzeczownika, przymiotnik odmieniający się przez rodzaj, zaimek wskazujący, polska kopula; angielski rzeczownik regularny, przymiotnik nieodmienny, `a/an` z jawnej cechy fonetycznej, `this` i `is`;
- 20 niezależnie opisanych w Go znaczeń M1 (10 identyfikacji, 10 cech) sprawdzonych przez `PL→IR`, `EN→IR`, niezależne `IR→PL`, `IR→EN`, `PL→EN`, `EN→PL`;
- testy kompozycji trzech dodatkowych par, które nie znajdują się w próbce 100, oraz test nowego leksemu `lampa/lamp` z niestandardowym ID (bez modyfikacji parsera i generatora);
- odmowy niewłaściwego `a/an`, zgody demonstrative/adjective/copula, błędnej semantyki, niekompletnych ról i nieautoryzowanego wariantu konstrukcji;
- dwa języki kompilowane jednorazowo do indeksów wyszukiwania form i sensów, z jawnym limitem kroków analizy i liczby kandydatów.

## Co jeszcze NIE działa

- Produkcyjne zdania testowe 021–100 oraz brakujące 480 ścieżek semantycznych;
- poprawna, kompletna deklinacja i koniugacja całego polskiego i angielskiego (M1 zawiera tylko potrzebne, sprawdzone komórki; nie należy zgadywać pozostałych);
- pełna kompozycja NP/VP/PP/Clause, formy bez zaimka wskazującego, wieloról, czasy, aspekt, negacja i pytania;
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

**Wymaganie językowe: 20/100 scenariuszy × 6 = 120/600 ścieżek.** Testy nowego słowa, morfologii, odwracania i kompilerów są dodatkowe; nie dolicza się ich do 600. Akceptacja właściciela: NOT EVALUATED.
