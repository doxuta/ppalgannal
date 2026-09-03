package ppalgannal

import (
	"fmt"
	"time"
)

// temporaryHolidays is 제2조제11호 — days the Cabinet designated ad hoc —
// since the start of the supported range. Presidential elections held to fill
// a vacancy (2017, 2025) are not 임기만료 elections under 제10호의2 and were
// designated this way instead.
var temporaryHolidays = []Holiday{
	temp(2015, 8, 14, "광복 70주년 (광복절 연휴)", "70th anniversary of Liberation"),
	temp(2016, 5, 6, "어린이날 연휴 (내수 진작)", "Bridge day after Children's Day"),
	temp(2017, 5, 9, "제19대 대통령선거 (궐위)", "Presidential election held to fill a vacancy"),
	temp(2017, 10, 2, "추석 연휴 (열흘 연휴)", "Bridge day before National Foundation Day and Chuseok"),
	temp(2020, 8, 17, "광복절 연휴 (코로나19)", "Bridge day after Liberation Day, COVID-19 relief"),
	temp(2023, 10, 2, "추석 연휴", "Bridge day between Chuseok and National Foundation Day"),
	temp(2024, 10, 1, "국군의 날 (건군 76주년)", "Armed Forces Day, 76th anniversary"),
	temp(2025, 1, 27, "설 연휴", "Bridge day before Seollal"),
	temp(2025, 6, 3, "제21대 대통령선거 (궐위)", "Presidential election held to fill a vacancy"),
}

func temp(y int, m time.Month, d int, reason, en string) Holiday {
	return Holiday{Date: Date(y, m, d), Name: "임시공휴일", NameEN: "Temporary holiday: " + en,
		Kind: Temporary, Basis: "제2조제11호", Reason: reason}
}

func temporaries(y int) []Holiday {
	var out []Holiday
	for _, h := range temporaryHolidays {
		if h.Date.Year() == y {
			out = append(out, h)
		}
	}
	return out
}

// Election cycles, 공직선거법 제34조제1항: the election is held on the first
// Wednesday on or after (term expiry − lead days). Terms are anchored on real
// expiry dates and continued by the term length. Assembly and local terms
// cannot be cut short, so their continuation is exact; a presidential term
// can (2017, 2025), so presidential days past the last anchor are marked
// "(projected)".
var cycles = []struct {
	name, en string
	expiries []time.Time // known term expiries; projected by term after the last
	term     int         // years
	lead     int         // days before expiry, 제34조제1항 각 호
	resets   bool        // an early election restarts the cycle (presidency only)
}{
	{"대통령선거", "Presidential election",
		[]time.Time{Date(2022, 5, 9), Date(2030, 6, 3)}, 5, 70, true},
	{"국회의원선거", "National Assembly election",
		[]time.Time{Date(2016, 5, 29)}, 4, 50, false},
	{"전국동시지방선거", "Nationwide local elections",
		[]time.Time{Date(2014, 6, 30)}, 4, 30, false},
}

// elections returns the regular election days of year y (제2조제10호의2).
// base is the year's 제2조 holidays, needed for the 제34조제2항 shift: an
// election day that is, or is adjacent to, a holiday moves a week later.
func elections(y int, base []Holiday) []Holiday {
	isHoliday := func(d time.Time) bool {
		if d.Weekday() == time.Sunday {
			return true
		}
		for _, h := range base {
			if h.Date.Equal(d) {
				return true
			}
		}
		return false
	}
	var out []Holiday
	for _, c := range cycles {
		exp := c.expiries[len(c.expiries)-1]
		expiries := append([]time.Time(nil), c.expiries...)
		for exp.Year() < y {
			exp = exp.AddDate(c.term, 0, 0)
			expiries = append(expiries, exp)
		}
		for i, e := range expiries {
			w := e.AddDate(0, 0, -c.lead)
			for w.Weekday() != time.Wednesday {
				w = w.AddDate(0, 0, 1)
			}
			for isHoliday(w) || isHoliday(w.AddDate(0, 0, -1)) || isHoliday(w.AddDate(0, 0, 1)) {
				w = w.AddDate(0, 0, 7)
			}
			if w.Year() != y {
				continue
			}
			reason := fmt.Sprintf("임기만료일 %s", ISO(e))
			if c.resets && i >= len(c.expiries) {
				reason += " (projected)"
			}
			out = append(out, Holiday{Date: w, Name: c.name, NameEN: c.en, Kind: Election,
				Basis: "제2조제10호의2 · 공직선거법 제34조", Reason: reason})
		}
	}
	return out
}
