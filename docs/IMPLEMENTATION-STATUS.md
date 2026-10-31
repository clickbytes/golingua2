# GoLingua 2 — techniczny odbiór M0–M8

**Stan:** pierwsze 100 ręcznie wybranych par A1 przeszło sześć niezależnych ścieżek na przypadek — **100/100 i 600/600**. Nie jest to certyfikacja CEFR, pełny translator języka naturalnego ani automatyczna akceptacja językowa właściciela (`NOT EVALUATED`).

## Wdrożony zakres

| Etap | Zakres testów Go | Zatwierdzone scenariusze |
| --- | --- | ---: |
| M0 | katalog 100 par w `tests/a1/a1_cases_test.go`, typy, kompilator, CLI | 0 |
| M1 | identyfikacja i cechy z niezależnym IR | 20 |
| M2 | posiadanie, predykaty 1- i 2-argumentowe, walencja | 50 |
| M3 | negacja `Not(Event)`, do-support i polska rekcja | 60 |
| M4 | 7 relacji miejsca i PP licencjonowane klasą miejsca | 70 |
| M5 | dokładna liczebność 1–5, istnienie, count/mass | 80 |
| M6 | pytania polar/who/what/where/how-many, jeden `Ask` | 90 |
| M7 | przeszłość i przyszłość względem wczoraj/jutro | 100 |
| M8 | stabilizacja, testy `-race`, CI, CLI, benchmark | 100 |

Źródła PL i EN mają osobno ręcznie zapisane grafy znaczenia w kodzie Go. Każdy przypadek sprawdza `PL→IR`, `EN→IR`, `IR→PL`, `IR→EN`, `PL→EN`, `EN→PL`. Różnice rodzajnika i referencji z EN nie są automatycznie dopisywane do polskiego IR. Przypadki nie są ukrytymi frazami do dopasowania podczas translacji.

## Kontrakt semantyczny czasu (M7)

Węzeł `Event` zawiera czas (`past`/`future`) i ograniczoną referencję `previous_day`/`next_day`. Polski past ma formy zgody gramatycznej; angielski ma niezmienny past lub leksykalnie potwierdzone `slept/ran/read`. Przyszłość EN używa `will + base`; PL używa `będzie + forma -ł/-ła/-ło` lub `będę + infinitive`. Zjawisko pogodowe ma jawne `theme=rain` w IR, choć EN używa czysto powierzchniowego `it`.

Wszystkie te formy są generowane przez wspólne klasy morfologii i lokalnie zadeklarowane konstrukcje, bez tabel 100 zdań w produkcyjnym kodzie. Ramy walencji licencjonują sensy; przypadek, do-support i czas dobiera pakiet językowy.

## Granice i dalszy rozwój

- A1 oznacza **wybraną setkę zdań**, nie pełen poziom ani swobodną rozmowę.
- Nie ma dowolnych dat, przysłówków czasu, temporalnego `Not(Event)` we wszystkich kombinacjach, aspektowych par perfektw/imperfective, modalności, zagnieżdżenia i koordynacji.
- Liczebność w tej dostawie ma formy 1–5 i ograniczone konstrukcje egzystencjalne; pytania obejmują wybrane typowane luki, nie pełną logikę zapytań.
- Nie ma wszystkich paradygmatów PL/EN, kontekstowej dezambiguacji świata, pełnej referencji/anaphora ani wnioskowania o faktach.
- Wydajność mierzona w ciepłym silniku to CPU/alloc, nie energia i nie porównanie z modelami neuronowymi.

**Warunek dalszego rozszerzania:** nowy rzeczownik, czasownik lub poprawne połączenie znanych elementów powinny korzystać z istniejących algorytmów. Gdy brakuje paradygmatu, roli lub konstrukcji, należy dodać kontrakt językowy albo zwrócić typowaną odmowę, a nie zakodować gotowe zdanie.
