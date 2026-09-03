# ppalgannal 빨간날

**South Korea's public holidays in Go, computed from the decree — including
substitute holidays (대체공휴일), regular election days, and the reason for
each.**

```
$ ppalgannal 2026-10-05
2026-10-05 Mon 대체공휴일 — 개천절 (2026-10-03) fell on Saturday (제3조제1항제1호)
```

빨간날 is what Koreans call a day printed in red on the calendar. This
package produces those days for any year from 2014 to 2199 by applying the
three clauses of 관공서의 공휴일에 관한 규정 제3조 to the holiday set that
was in force on each date, with the lunar dates (설날, 부처님 오신 날, 추석)
coming from [`amlich`](https://github.com/doxuta/amlich) at Korea's UTC+9
civil clock. No API key, no network, no lookup table of substitutes.

## Why

There is no Go implementation of the Korean holiday calendar. `rickar/cal`
has forty countries and no `kr`; the usual answer is KASI's 특일 API on
data.go.kr, which needs a key and a network round-trip to tell you whether
next Monday is a working day.

The rule itself is short, and yet in the six weeks before this was written
two libraries in other languages filed issues about getting it wrong
([`commenthol/date-holidays#640`](https://github.com/commenthol/date-holidays/issues/640):
no substitution rules at all, "if weekend then next weekday" already found
insufficient;
[`holiday-calendar-java#260`](https://github.com/holiday-calendar/holiday-calendar-java/issues/260):
proposes shipping the lunar-block substitutes as a hand-curated CSV, and
states that 한글날 is never substituted — it has been since 2021). While
checking the golden data for this repo, Korean Wikipedia said 개천절 is
never substituted (2026-10-05 is one), publicholidays.co.kr showed a 2027
substitute for 현충일 (제8호 is excluded from 제3조) and put 2028's 부처님 오신 날
on 05-05 (it is 05-02), and officeholidays.com omitted 2022-10-10. Four
sources, four different mistakes, one three-clause decree.

## Install

```bash
go install github.com/doxuta/ppalgannal/cmd/ppalgannal@latest
```

```bash
ppalgannal 2027 | grep 대체
```

```
2027-02-09 Tue 대체공휴일 — 설날 (2027-02-07) fell on Sunday (제3조제1항제2호)
2027-05-03 Mon 대체공휴일 — 노동절 (2027-05-01) fell on Saturday (제3조제1항제1호)
2027-07-19 Mon 대체공휴일 — 제헌절 (2027-07-17) fell on Saturday (제3조제1항제1호)
2027-08-16 Mon 대체공휴일 — 광복절 (2027-08-15) fell on Sunday (제3조제1항제1호)
2027-10-04 Mon 대체공휴일 — 개천절 (2027-10-03) fell on Sunday (제3조제1항제1호)
2027-10-11 Mon 대체공휴일 — 한글날 (2027-10-09) fell on Saturday (제3조제1항제1호)
2027-12-27 Mon 대체공휴일 — 기독탄신일 (2027-12-25) fell on Saturday (제3조제1항제1호)
```

Those seven match the 2027 월력요항 published by KASI. `ppalgannal` alone
prints the current year; `ppalgannal 2026-10-06` answers `workday`;
`-json` emits an array with `date`, `name`, `name_en`, `kind`, `basis`,
`reason`.

## Library

```go
import "github.com/doxuta/ppalgannal"

hs, _ := ppalgannal.Year(2026)          // every holiday, date order, Sundays excluded
for _, h := range hs {
    fmt.Println(h)                       // 2026-06-03 Wed 전국동시지방선거 — 임기만료일 2026-06-30
}
ppalgannal.On(ppalgannal.Date(2025, 5, 6))   // []Holiday{대체공휴일 — 어린이날 coincided with 부처님 오신 날}
ppalgannal.IsWorkday(ppalgannal.Date(2026, 10, 5)) // false
```

`Holiday` carries `Date` (a civil date as `time.Time` at UTC midnight),
`Name` as written in the decree, `NameEN`, `Kind`
(`Fixed`/`Lunar`/`Substitute`/`Election`/`Temporary`), `Basis` (the article),
and `Reason` for anything that is not a plain fixed or lunar holiday.

## How the rule is applied

제3조 (대통령령 제36290호, in force 2026-05-11) says: for the holidays of
제2조 items 2–10,

| clause | condition | who |
|---|---|---|
| ①1 | falls on Saturday or Sunday | 국경일 (3·1절·제헌절·광복절·개천절·한글날), 부처님 오신 날, 노동절, 어린이날, 기독탄신일 |
| ①2 | falls on Sunday | 설날 연휴, 추석 연휴 (Saturday is *not* lost) |
| ①3 | coincides with another 제2조 holiday on a weekday | all of the above |
| ② | two substitutes land on the same day | the second moves to the next non-holiday |
| ③ | a substitute lands on Saturday | it moves to the next non-holiday |

1월 1일 and 현충일 are not in any clause and are never substituted.

The implementation is one loop: every date that carries an eligible holiday
and is lost under ①1–①3 produces one substitute, the first later day that is
neither a holiday, a weekend, nor an already-assigned substitute. That
single sentence is ② and ③ as well. Measured over 2014–2199, ③ never
fires: a weekday overlap only ever happens inside the 추석 block or on
어린이날, and the Saturday after is never free, so the clause is pinned by
a synthetic test instead of a real year. Eligibility is checked against the date,
not the year, using the amendment history:

| in force | eligible |
|---|---|
| 2013-11-05 (제24828호) | 설날·추석 연휴 (Sunday/overlap), 어린이날 (also Saturday) |
| 2021-08-04 (제31930호) | + 3·1절, 광복절, 개천절, 한글날 |
| 2023-05-04 (제33448호) | + 부처님 오신 날, 기독탄신일 |
| 2026-05-01 (제36290호) | + 노동절 and 제헌절, both new holidays under 법률 제21338호 |

Regular election days (제2조제10호의2) come from 공직선거법 제34조: the first
Wednesday on or after term expiry minus 70/50/30 days for
대통령/국회의원/지방선거, moved a week later if that day or a neighbour is a
holiday. Eight real election days 2014–2026 reproduce, including the two
that were shifted (2018-06-13 off 현충일, 2022-03-09 off 3·1절). Temporary
holidays (제11호) and the two vacancy presidential elections (2017-05-09,
2025-06-03) are a nine-row table.

Pinned by tests, meant to fail on a data or rule change:

```
substitute holidays 2014–2027 reproduced       32 of 32
regular election days 2014–2026 reproduced      8 of 8
2026 full list vs 인사혁신처 gazette             22 of 22
```

## Limitations

- **Years before 2014 are refused.** 대체공휴일 began on 2014-01-01 and the
  pre-2014 holiday set changed several times (제헌절 dropped in 2008, 한글날
  restored in 2013); modelling that history was not the point.
- **Temporary holidays are a table, so the future has none.** A 임시공휴일
  designated after this repo's last update will be missing until the row is
  added. Check the tag date before trusting a year that has already started.
- **Presidential election days after the 2030 term are projected** from a
  five-year continuation and marked `(projected)` in `Reason`; an early
  election resets the cycle, as 2017 and 2025 did. Assembly and local
  cycles cannot be cut short and are exact.
- **The two-substitutes case is an interpretation.** ② has only ever been
  exercised by the 2025-05-05 어린이날/부처님 오신 날 collision, which the
  government treated as one lost day. This code gives one substitute per
  lost *date*, which reproduces 2025 and predicts two substitutes for 2039
  (추석 on Sunday 10-02, 추석 다음 날 + 개천절 on 10-03 → 10-04 and 10-05).
  Nobody has gazetted 2039 yet.
- **관공서 공휴일 only.** 근로자의 날 before 2026 was a paid day off for
  workers under a different statute, not a public holiday, and is not listed;
  from 2026 노동절 is in 제2조 and is. Stock-exchange closures and company
  bridge days are out of scope.
- `IsWorkday` treats Saturday as non-working (the five-day week has applied
  to public offices since 2005), although the decree lists only Sunday.

## Prior art and sources

- [`vacanza/holidays`](https://github.com/vacanza/holidays) (Python) has the
  most complete Korean rule set and per-year expectations back to 1948; its
  tests were one of three sources for the golden substitutes here.
  [`commenthol/date-holidays`](https://github.com/commenthol/date-holidays)
  (JS) and [`holiday-calendar-java`](https://github.com/holiday-calendar/holiday-calendar-java)
  are the two whose open issues prompted this.
- Decree text: [관공서의 공휴일에 관한 규정](https://www.law.go.kr/법령/관공서의공휴일에관한규정)
  and its [amendment history](https://law.go.kr/LSW/lsRvsRsnListP.do?lsId=002404&chrClsCd=010102);
  [공직선거법 제34조](https://www.law.go.kr/법령/공직선거법).
- Golden data: 인사혁신처 annual 고시 (2024–2026 via kholidayz.com), the 2027
  월력요항 from KASI, officeholidays.com for 2014–2023.
- Lunar conversion: [`doxuta/amlich`](https://github.com/doxuta/amlich), the
  only dependency, itself standard-library only.

Built with an AI coding agent (Claude) under human review, as with the other
repositories on this account; the decree was read in full and every golden
row was checked against at least two sources before being pinned.

## 한국어 요약

「관공서의 공휴일에 관한 규정」 제3조 세 개 항을 그대로 구현해 대체공휴일을
계산하는 Go 라이브러리·CLI입니다. 설날·부처님 오신 날·추석은 `amlich`로
음력 변환하고, 임기만료 선거일은 공직선거법 제34조로 계산하며, 임시공휴일은
표로 넣었습니다. 2014–2027년의 대체공휴일 32건과 2014–2026년의 선거일 8건이
모두 재현되며, 각 공휴일에 근거 조문과 사유가 붙습니다. 2039년처럼 이틀이
겹치는 경우는 해석(잃은 날짜 하나당 대체 하루)에 따른 예측이고, 임시공휴일은
새로 지정되면 표에 추가해야 합니다.

## License

MIT © Xuan Tai Doan
