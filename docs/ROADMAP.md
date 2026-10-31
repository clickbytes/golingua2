# GoLingua 2 — kolejność implementacji na 100 przykładach A1

**Zasada:** jeden writer i małe, sprawdzalne integracje na bazie jednego kontraktu. Nie programujemy 100 zdań po kolei; najpierw budujemy algorytmy, a grupy pojawiają się jako ich konsekwencje.

| Faza | Główny artefakt | Grupy A1 dodane do odbioru | Suma scenariuszy | Bramka techniczna |
| --- | --- | --- | ---: | --- |
| M0 ✓ | Repo Go, typy, Go test inventory 100, decyzje | 0 | 0 | `go test ./...` (testy struktury) |
| M1 ✓ | CORE/IR, PL+EN klasy noun/copula/adjective, NP i identyfikacja | 001–020 | 20 | 6 przepływów / każdy z 20 |
| M2 ✓ | `have`, czasownik, walencja, regularne NP/VP | 021–050 | 50 | +30 scenariuszy |
| M3 ✓ | `Not`, EN auxiliary, polska rekcja pod negacją | 051–060 | 60 | +10 i kontrasty zakresu |
| M4 ✓ | Relacje przestrzenne i PP | 071–080 | 70 | +10, bez na→on globalnie |
| M5 | Ilość 1–5, dokładne referenty, egzystencja | 081–090 | 80 | +10, poprawna zgoda/count |
| M6 | Pytania polar/WH/where/how-many, jeden `Ask` | 061–070 | 90 | +10, brak wymyślonej odpowiedzi |
| M7 | Przeszłość, przyszłość, odniesienie czasowe | 091–100 | 100 | 100 scenariuszy ×6, brak maskowania GAP |
| M8 | Stabilizacja, porównanie, wydajność i dokumentacja | Bez nowych zdań | 100 | powtarzalne CI i benchmarks CPU |

## M0 — bootstrap i granice API

1. Tworzymy repo `golingua2` (obecnie puste), `go.mod`, docs i katalog 100 par z tego pakietu.
2. `internal/core` opisuje sorty i koncepty, ale nie odniesienia do rzeczowników PL/EN.
3. `internal/sem` definiuje **jeden** IR i walidację; testy ręcznie zbudowanych znaczeń przechodzą, zanim pojawi się parser.
4. `internal/morph` kompiluje typowane klasy, w tym brakujące części i wyjątki; nie zapisujemy pełnej tablicy form pojedynczych leksemów.
5. Domykamy interfejs `LanguagePack`, `CompilePack`, statusy i budżety.

## M1 — pierwszy prawdziwy przekrój

Implementacja obejmuje wskazania `To jest X` / `This is a/an X` i atrybucję `Ten/Ta/To X jest Y` / `This X is Y`. Wymagane: osoby/referenty, rodzajnik EN, wskazujący PL, formy `być/be`, rodzaj/liczba/przypadek i rozdzielenie definicji od wystąpienia. Wszystkie 20 znaczeń tworzymy niezależnie w Go.

## M2 — predykaty i role

`Have`, trzy rodzaje podmiotów, `SVO` z dwoma referentami i leksemami, walencja różnicująca `pomaga dziecku` od biernika, oraz zdania nieprzechodnie bez fikcyjnej roli dopełnienia. Nowe leksemy z już znanej klasy dopisywane tylko do `lexemes.go`/`senses.go`. Test dodatkowego słowa sprawdza rozszerzalność.

## M3 — negacja

Wyprowadzenie operatora `Not(event)` w IR; EN do/does not + base; PL `nie` + czasownik i licencjonowany dopełniacz np. przy `czytać`/`jeść`, ale nie globalny dopełniacz każdego dopełnienia. Odbiór: zmiana zakresu i brak prawdziwego pacjenta daje odmowę, tekst nie może być pozornie kompletny.

## M4 — relacje lokalizacji

Jeden model relacji przestrzennych z właścicielem lokacji, argumentami i językową rekcją. `w domu`, `na stole`, `pod stołem`, `obok mamy`, `at school`, `in the tree`: przyimki wybiera lokalna konstrukcja w języku, nie słownik PL→EN na stringach.

## M5 — liczebność

Liczebność semantyczna oddzielona od gramatycznej. Różne reguły `dwie książki`, `dwa psy`, `pięć książek`. `There is/are` nie tworzy samodzielnie nowego konceptu „there”; istniejąca semantyka egzystencjalna i locative licencjonuje konstrukcję.

## M6 — pytania

Wprowadzenie typu `Ask` i zmiennej odpowiedzi z sortem. Dla `who/what/where/how many` używamy dokładnych ról/relacji poprzednich etapów. EN `do-support` i inversion wynikają z planu zdania, nie osobnych tablic pełnych pytań.

## M7 — czas i aspekt

Podstawowa przeszłość/nieciągły czas teraźniejszy, przyszłość `will` i PL `będzie + past/infinitive` w dozwolonym fragmencie. Relacja `yesterday/tomorrow` do stałego zegara testowego. Nie przypisujemy perfective bez źródłowej przesłanki; jeśli EN past nie rozstrzyga aspektu PL, wariant ma jawny wybór.

## M8 — wypuszczenie v2 A1

Rzeczywisty CLI, wszystkie 600 ścieżek, kontrasty, zero specjalnych reguł po zdaniu, niewrażliwość na niepowiązane leksemy, trace błędów, powtarzalny hash pakietu i pomiary CPU/pamięci. Raportuje się liczbę obsługiwanych klas, leksemów, kompozycji i wyjątków, nie tylko wynik 100/100.

## Gdy propozycja nie działa

Nie zmieniamy zdań testowych, żeby ukryć błąd, chyba że udowodniono ich niezależny błąd językowy i zmiana jest jawna. Nie kopiujemy komponentów v1 hurtowo. Wykorzystujemy ich wiedzę i testy kontrastowe, ale brakujące reguły projektujemy jako klasy, nie kolejne specjalizacje.
