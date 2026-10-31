# GoLingua 2 — decyzje architektoniczne do utrzymania

Każdy ADR w tej tabeli jest **projektowaną decyzją**, nie zatwierdzeniem właściciela ani działającą funkcją.

| ADR | Decyzja | Czego unikamy | Koszt / konsekwencja |
| --- | --- | --- | --- |
| 001 | Go jako język implementacji | odtwarzania dawnego JS | kompilator i pluginy muszą być czytelne dla Go |
| 002 | 100 par A1 jako tabela Go | JSON z ręcznie zapisanymi zdaniami | aktualizacja korpusu wymaga zmiany kodu testów |
| 003 | Jeden typowany IR, z osobnym `Evidence` | Meaning v1 + DSE jako dwa autorytety | jawne typy/operator scope od pierwszej iteracji |
| 004 | Pakiet języka jest zbiorem deklaracji kompilowanych | `switch` per słowo/sens/zdanie | potrzebny porządny kompilator/walidator |
| 005 | Proceduralne klasy odmiany PL/EN | 14- lub 21-formowe ręczne tabele na leksem | część nieregularności nadal jest danymi |
| 006 | Parser i generator mają jeden kontrakt, z jawnymi kierunkami | bezpodstawna automatyczna odwracalność | obowiązkowe testy obu kierunków |
| 007 | A1 oznacza kontrolowany podzbiór, nie certyfikat CEFR | obietnicy całego języka | widoczne `unsupported` dla nieobsługiwanych struktur |
| 008 | Kontrakt odbioru zawiera IR i alternatywy, nie tylko tekst | `expected == actual` jako jedyna miara | większy koszt napisania semantycznych fixture |
| 009 | Kompilacja pakietów na starcie i niezmienne indeksy | skanowania całych słowników per żądanie | koszt początkowy, kontrola pamięci |
| 010 | Najpierw działający A1, później trzeci język/FST/optymalizacja | nadmierna abstrakcja przed odbiorem | wydajność oceniamy na pomiarach, nie deklaracjach |

## Rzeczy do jeszcze zamknięcia z właścicielem przed zamrożeniem API

1. Czy wszystkie deklaracje leksemów i paradygmatów mają pozostać w Go na stałe? Propozycja: na A1 tak; późniejszy importer danych może skompilować dokładnie te same typowane kontrakty, **bez zmiany algorytmów**.
2. Czy 100 par jest sztywno zamrożonym pierwszym testem akceptacyjnym? Propozycja: tak, po jednokrotnym review językowym, z publicznym śladem każdej korekty.
3. Które informacje z angielskiego `a/the` zaliczamy do semantycznej referencji, a które do językowego planu? Propozycja: jedna jawna polityka referencji oraz dowody/założenia per wystąpienie, nie jedno uniwersalne „article=...”.
4. Czy dla przyszłości PL w A1 używamy bezokolicznika (`będę pracować`), formy przeszłej (`będzie spał`) i jak mapujemy aspekt EN? Propozycja: obie zatwierdzone konstrukcje, z jawnymi warunkami i alternatywami.
5. Jaki jest minimalny zestaw klas odmiany do M1? Propozycja: rzeczowniki masc/fem/neut w A1, przymiotniki regularne, be/być oraz niezbędne funkcje wskazujące; M2 dodaje koniugacje i czasowniki nieregularne.

## Obszary świadomie wyłączone

Pamięć długoterminowa, rozumowanie o faktach świata, automatyczne uczenie z danych online, pełny parser języka, tłumaczenie nieograniczonych tekstów, generatywny model neuronowy, prywatne API/pluginy, JS parity i benchmark energetyczny, dopóki nie ma działającego, poprawnego silnika do porównania.
