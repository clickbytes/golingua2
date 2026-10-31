# Zasady pracy nad GoLingua 2

**Obowiązuje RFC-001 i RFC-002** w `docs/`, testy są w Go, nie w JSON. Startowy cel to dokładnie 100 scenariuszy A1 i sześć ścieżek na każdy.

- Nie dodawaj produkcyjnego kodu rozpoznającego zdanie po pełnym tekście ani po ID przypadku.
- Jeden IR jest jedynym autorytetem znaczenia. Nie twórz drugiej reprezentacji autorytatywnej, cichego fallbacku ani przepisywania go przez inne DSL podczas translacji.
- Nie dodawaj `supportedSenses` jako drugiej listy uprawnień. Każde dopuszczenie wyprowadź z zatwierdzonych `Sense + Frame + Construction + Direction`.
- Morfologia EN/PL musi działać przez wielokrotnie wykorzystywane klasy i parametry leksemu; odrębne komórki zapisuj tylko, jeśli są rzeczywistymi wyjątkami.
- Nie przenoś automatycznie produkcyjnego GoLingua v1 ani jego skumulowanych historycznych testów. Używaj v1 jako źródła błędów i kontrastów, nie jako API do zachowania.
- Testy nie pobierają oczekiwanego znaczenia od produkcyjnego analizatora i nie uznają dowolnego wygenerowanego tekstu za poprawne tłumaczenie.
- Nie zmieniaj wpisów A1 ani oczekiwań pod istniejący output bez jawnego, językowo uzasadnionego przeglądu. Polskie niedookreślenie rodzajnika i liczby musi pozostać widoczne.
- Pisz najmniejszą sensowną zmianę i dokumentuj faktycznie uruchomione testy, dokładne polecenia oraz kody wyjścia. `NOT RUN` nie jest PASS.
- Nie twórz dodatkowych list testowych JSON/YAML. `tests/a1/a1_cases_test.go` jest wyłącznym spisem 100 przykładów.
- Specyfikacje i kroki M0–M8 są propozycją; nie deklaruj wdrożenia, zanim realny `engine` i jego odbiór nie istnieją.
