// Package ppalgannal computes South Korea's public holidays (관공서의 공휴일)
// for a Gregorian year straight from the text of the governing decree,
// 관공서의 공휴일에 관한 규정 (대통령령 제36290호, 시행 2026-05-11):
//
//   - the fixed-date holidays of 제2조;
//   - the lunar ones (설날 연휴, 부처님 오신 날, 추석 연휴), converted with
//     github.com/doxuta/amlich at Korea's UTC+9 civil clock;
//   - substitute holidays (대체공휴일) from the three clauses of 제3조, using
//     the eligibility set that was in force on the date in question, so 2014
//     to today are all reproduced with one rule;
//   - regular election days from 공직선거법 제34조;
//   - the temporary holidays (임시공휴일) the government has designated since
//     2014, as a table.
//
// Dates are civil dates carried as time.Time at UTC midnight; build them with
// Date and compare with Equal. Sundays are a 공휴일 under 제2조제1호 but are
// not listed; IsWorkday accounts for them.
package ppalgannal

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/doxuta/amlich"
)

// Kind says how a day became a holiday.
type Kind int

const (
	// Fixed is a fixed Gregorian date: 제2조제2호·제3호·제6호·제7호·제8호·제10호.
	Fixed Kind = iota
	// Lunar is computed from the Korean lunisolar calendar: 제2조제4호·제5호·제9호.
	Lunar
	// Substitute is a 대체공휴일 under 제3조.
	Substitute
	// Election is the day of a regular (임기만료) election: 제2조제10호의2.
	Election
	// Temporary is a 임시공휴일 designated by the government: 제2조제11호.
	Temporary
)

func (k Kind) String() string {
	return [...]string{"fixed", "lunar", "substitute", "election", "temporary"}[k]
}

// Holiday is one public holiday resolved to a civil date.
type Holiday struct {
	Date   time.Time // civil date at UTC midnight
	Name   string    // Korean name, as written in the decree
	NameEN string
	Kind   Kind
	Basis  string // the article that makes the day a holiday
	Reason string // Substitute, Election, Temporary: why this day
}

// String renders "2026-10-05 Mon 대체공휴일 — 개천절 (2026-10-03) fell on Saturday".
func (h Holiday) String() string {
	s := fmt.Sprintf("%s %s %s", ISO(h.Date), h.Date.Weekday().String()[:3], h.Name)
	if h.Reason != "" {
		s += " — " + h.Reason
	}
	return s
}

// MinYear is the first supported year: the substitute-holiday system took
// effect on 2014-01-01 and the pre-2014 holiday set differs in other ways.
const MinYear = 2014

// MaxYear is the last supported year; amlich's lunar conversion is
// cross-validated up to 2199.
const MaxYear = 2199

// ErrUnsupportedYear is returned for years outside [MinYear, MaxYear].
var ErrUnsupportedYear = errors.New("ppalgannal: year outside 2014–2199")

// Date builds a civil date (UTC midnight).
func Date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// ISO renders a civil date as YYYY-MM-DD.
func ISO(t time.Time) string { return t.Format("2006-01-02") }

func day(t time.Time) time.Time { return Date(t.Year(), t.Month(), t.Day()) }

// substitutable lists the holidays 제3조 makes eligible for a substitute, the
// date the decree first covered them, and whether Saturday counts
// (제3조제1항제1호) or only Sunday does (제1항제2호). Both classes are also
// substituted when they coincide with another 제2조 holiday on a weekday
// (제1항제3호). 1월 1일 and 현충일 are absent: they are never substituted.
var substitutable = map[string]struct {
	saturday bool
	from     time.Time
}{
	// 대통령령 제24828호 (2013-11-05): 설·추석 연휴 on Sunday/holiday, 어린이날 also on Saturday.
	"설날 전날":   {false, Date(2013, 11, 5)},
	"설날":      {false, Date(2013, 11, 5)},
	"설날 다음 날": {false, Date(2013, 11, 5)},
	"추석 전날":   {false, Date(2013, 11, 5)},
	"추석":      {false, Date(2013, 11, 5)},
	"추석 다음 날": {false, Date(2013, 11, 5)},
	"어린이날":    {true, Date(2013, 11, 5)},
	// 대통령령 제31930호 (2021-08-04): the four 국경일 that were holidays at the time.
	"3·1절": {true, Date(2021, 8, 4)},
	"광복절":  {true, Date(2021, 8, 4)},
	"개천절":  {true, Date(2021, 8, 4)},
	"한글날":  {true, Date(2021, 8, 4)},
	// 대통령령 제33448호 (2023-05-04).
	"부처님 오신 날": {true, Date(2023, 5, 4)},
	"기독탄신일":    {true, Date(2023, 5, 4)},
	// 대통령령 제36290호 (2026-04-30, in force 2026-05-01): both are new holidays.
	"노동절": {true, Date(2026, 5, 1)},
	"제헌절": {true, Date(2026, 5, 1)},
}

// Year returns every public holiday of Gregorian year y in date order,
// Sundays excluded. Substitutes, election days and temporary holidays are
// included and carry a Reason.
func Year(y int) ([]Holiday, error) {
	if y < MinYear || y > MaxYear {
		return nil, ErrUnsupportedYear
	}
	base, err := baseHolidays(y)
	if err != nil {
		return nil, err
	}
	all := append(base, elections(y, base)...)
	all = append(all, temporaries(y)...)
	all = append(all, substitutes(all)...)
	sort.SliceStable(all, func(i, j int) bool {
		if !all[i].Date.Equal(all[j].Date) {
			return all[i].Date.Before(all[j].Date)
		}
		return all[i].Kind < all[j].Kind
	})
	return all, nil
}

// On returns the holidays falling on civil date d, or nil for a working day
// or a plain weekend. Years outside the supported range yield nil.
func On(d time.Time) []Holiday {
	hs, err := Year(d.Year())
	if err != nil {
		return nil
	}
	d = day(d)
	var out []Holiday
	for _, h := range hs {
		if h.Date.Equal(d) {
			out = append(out, h)
		}
	}
	return out
}

// IsWorkday reports whether d is a Monday–Friday that is not a public
// holiday. Saturday is not a 공휴일 in the decree, but the five-day week has
// been the rule for public offices since 2005, so it is not a workday here.
func IsWorkday(d time.Time) bool {
	wd := d.Weekday()
	return wd != time.Saturday && wd != time.Sunday && On(d) == nil
}

// baseHolidays is 제2조 items 2–10 for year y: fixed dates plus the lunar ones.
func baseHolidays(y int) ([]Holiday, error) {
	fixed := func(m time.Month, d int, name, en, basis string) Holiday {
		return Holiday{Date: Date(y, m, d), Name: name, NameEN: en, Kind: Fixed, Basis: basis}
	}
	out := []Holiday{
		fixed(1, 1, "1월 1일", "New Year's Day", "제2조제3호"),
		fixed(3, 1, "3·1절", "Independence Movement Day", "제2조제2호 (국경일)"),
		fixed(5, 5, "어린이날", "Children's Day", "제2조제7호"),
		fixed(6, 6, "현충일", "Memorial Day", "제2조제8호"),
		fixed(8, 15, "광복절", "Liberation Day", "제2조제2호 (국경일)"),
		fixed(10, 3, "개천절", "National Foundation Day", "제2조제2호 (국경일)"),
		fixed(10, 9, "한글날", "Hangul Day", "제2조제2호 (국경일)"),
		fixed(12, 25, "기독탄신일", "Christmas Day", "제2조제10호"),
	}
	if y >= 2026 { // 법률 제21338호 / 대통령령 제36290호
		out = append(out,
			fixed(5, 1, "노동절", "Labour Day", "제2조제6호"),
			fixed(7, 17, "제헌절", "Constitution Day", "제2조제2호 (국경일)"),
		)
	}
	lunar := func(m, d int) (time.Time, error) {
		sy, sm, sd, err := amlich.LunarToSolar(amlich.LunarDate{Year: y, Month: m, Day: d}, amlich.Korea)
		return Date(sy, time.Month(sm), sd), err
	}
	seollal, err := lunar(1, 1)
	if err != nil {
		return nil, err
	}
	buddha, err := lunar(4, 8)
	if err != nil {
		return nil, err
	}
	chuseok, err := lunar(8, 15)
	if err != nil {
		return nil, err
	}
	block := func(mid time.Time, name, en, basis string) []Holiday {
		return []Holiday{
			{Date: mid.AddDate(0, 0, -1), Name: name + " 전날", NameEN: "Day before " + en, Kind: Lunar, Basis: basis},
			{Date: mid, Name: name, NameEN: en, Kind: Lunar, Basis: basis},
			{Date: mid.AddDate(0, 0, 1), Name: name + " 다음 날", NameEN: "Day after " + en, Kind: Lunar, Basis: basis},
		}
	}
	out = append(out, block(seollal, "설날", "Seollal", "제2조제4호")...)
	out = append(out, Holiday{Date: buddha, Name: "부처님 오신 날", NameEN: "Buddha's Birthday", Kind: Lunar, Basis: "제2조제5호"})
	out = append(out, block(chuseok, "추석", "Chuseok", "제2조제9호")...)
	return out, nil
}

// substitutes applies 제3조 to the year's holidays. Each date that carries an
// eligible holiday and is "lost" — to a Sunday, to a Saturday for the
// Saturday class, or to another 제2조 holiday on a weekday — produces one
// substitute: the first following day that is neither a holiday nor a
// weekend. Assigned substitutes count as holidays for the next search, which
// is exactly the 제2항 same-day rule, and skipping Saturday is 제3항.
//
// One lost date gives one substitute even when two holidays share it
// (2025-05-05: 어린이날 and 부처님 오신 날, one substitute on 05-06), which is
// how the government has applied the clause; two lost dates give two.
func substitutes(all []Holiday) []Holiday {
	byDate := map[time.Time][]Holiday{}
	for _, h := range all {
		byDate[h.Date] = append(byDate[h.Date], h)
	}
	dates := make([]time.Time, 0, len(byDate))
	for d := range byDate {
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })

	taken := map[time.Time]bool{}
	var out []Holiday
	for _, d := range dates {
		trigger, why := lost(d, byDate[d])
		if why == "" {
			continue
		}
		s := d.AddDate(0, 0, 1)
		for s.Weekday() == time.Saturday || s.Weekday() == time.Sunday || len(byDate[s]) > 0 || taken[s] {
			s = s.AddDate(0, 0, 1)
		}
		taken[s] = true
		out = append(out, Holiday{
			Date: s, Name: "대체공휴일", NameEN: "Substitute holiday", Kind: Substitute,
			Basis:  "제3조",
			Reason: fmt.Sprintf("%s (%s) %s", trigger.Name, ISO(d), why),
		})
	}
	return out
}

// lost decides whether date d is lost under 제3조제1항 and returns the
// holiday that triggers the substitute and the clause that applies.
func lost(d time.Time, hs []Holiday) (Holiday, string) {
	for _, h := range hs {
		r, ok := substitutable[h.Name]
		if !ok || d.Before(r.from) {
			continue
		}
		switch d.Weekday() {
		case time.Sunday:
			if r.saturday {
				return h, "fell on Sunday (제3조제1항제1호)"
			}
			return h, "fell on Sunday (제3조제1항제2호)"
		case time.Saturday:
			if r.saturday {
				return h, "fell on Saturday (제3조제1항제1호)"
			}
		default:
			for _, o := range hs {
				// Only 제2조제2호–제10호 count as "다른 공휴일"; election days and
				// temporary holidays sit outside that range.
				if o.Name != h.Name && (o.Kind == Fixed || o.Kind == Lunar) {
					return h, fmt.Sprintf("coincided with %s (제3조제1항제3호)", o.Name)
				}
			}
		}
	}
	return Holiday{}, ""
}
