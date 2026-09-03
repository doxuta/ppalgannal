package ppalgannal

import (
	"strings"
	"testing"
	"time"
)

func mustYear(t *testing.T, y int) []Holiday {
	t.Helper()
	hs, err := Year(y)
	if err != nil {
		t.Fatalf("Year(%d): %v", y, err)
	}
	return hs
}

func ofKind(hs []Holiday, k Kind) []string {
	var out []string
	for _, h := range hs {
		if h.Kind == k {
			out = append(out, ISO(h.Date))
		}
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Every substitute holiday the government has gazetted since the system
// began, plus 2027 from the 월력요항 published by KASI. Sources: 인사혁신처
// 고시 per year as reproduced by kholidayz.com (2024–2026), officeholidays.com
// and the vacanza/holidays test-suite (2014–2023), KB 2027 월력요항 summary.
var goldenSubstitutes = map[int][]string{
	2014: {"2014-09-10"},
	2015: {"2015-09-29"},
	2016: {"2016-02-10"},
	2017: {"2017-01-30", "2017-10-06"},
	2018: {"2018-05-07", "2018-09-26"},
	2019: {"2019-05-06"},
	2020: {"2020-01-27"},
	2021: {"2021-08-16", "2021-10-04", "2021-10-11"},
	2022: {"2022-09-12", "2022-10-10"},
	2023: {"2023-01-24", "2023-05-29"},
	2024: {"2024-02-12", "2024-05-06"},
	2025: {"2025-03-03", "2025-05-06", "2025-10-08"},
	2026: {"2026-03-02", "2026-05-25", "2026-08-17", "2026-10-05"},
	2027: {"2027-02-09", "2027-05-03", "2027-07-19", "2027-08-16", "2027-10-04", "2027-10-11", "2027-12-27"},
}

func TestSubstitutesGolden(t *testing.T) {
	total := 0
	for y := MinYear; y <= 2027; y++ {
		got := ofKind(mustYear(t, y), Substitute)
		if !equal(got, goldenSubstitutes[y]) {
			t.Errorf("%d substitutes: got %v, want %v", y, got, goldenSubstitutes[y])
		}
		total += len(got)
	}
	if total != 32 { // README claim
		t.Errorf("substitutes 2014–2027: got %d, want 32", total)
	}
}

// The complete 2026 list as gazetted (인사혁신처), Sundays excluded.
func TestYear2026Golden(t *testing.T) {
	want := []string{
		"2026-01-01", "2026-02-16", "2026-02-17", "2026-02-18", "2026-03-01", "2026-03-02",
		"2026-05-01", "2026-05-05", "2026-05-24", "2026-05-25", "2026-06-03", "2026-06-06",
		"2026-07-17", "2026-08-15", "2026-08-17", "2026-09-24", "2026-09-25", "2026-09-26",
		"2026-10-03", "2026-10-05", "2026-10-09", "2026-12-25",
	}
	var got []string
	for _, h := range mustYear(t, 2026) {
		got = append(got, ISO(h.Date))
	}
	if !equal(got, want) {
		t.Errorf("2026:\n got %v\nwant %v", got, want)
	}
}

func TestElections(t *testing.T) {
	want := map[int][]string{
		2014: {"2014-06-04"}, 2015: nil, 2016: {"2016-04-13"}, 2017: nil, // 2017-05-09 was a vacancy election
		2018: {"2018-06-13"}, // first Wednesday was 현충일 — 제34조제2항 shift
		2019: nil, 2020: {"2020-04-15"}, 2021: nil,
		2022: {"2022-03-09", "2022-06-01"},         // 03-02 sits next to 3·1절 — shifted
		2023: nil, 2024: {"2024-04-10"}, 2025: nil, // 2025-06-03 was a vacancy election
		2026: {"2026-06-03"}, 2027: nil, 2028: {"2028-04-12"},
	}
	for y, w := range want {
		if got := ofKind(mustYear(t, y), Election); !equal(got, w) {
			t.Errorf("%d elections: got %v, want %v", y, got, w)
		}
	}
	// Assembly and local cycles are exact; only a presidential day beyond the
	// last known term expiry is a projection.
	for _, y := range []int{2026, 2028} {
		for _, h := range mustYear(t, y) {
			if h.Kind == Election && strings.Contains(h.Reason, "projected") {
				t.Errorf("%d: %s must not be projected", y, h)
			}
		}
	}
	found := false
	for _, h := range mustYear(t, 2035) {
		if h.Kind == Election && h.Name == "대통령선거" {
			found = true
			if !strings.Contains(h.Reason, "projected") {
				t.Errorf("2035 presidential election must be projected: %s", h)
			}
		}
	}
	if !found {
		t.Error("2035 should carry a projected presidential election")
	}
}

func TestNeverSubstituted(t *testing.T) {
	// 1월 1일 and 현충일 are outside 제3조 — on a weekend they simply vanish.
	for _, d := range []string{
		"2022-01-01", "2023-01-01", "2028-01-01", // Sat, Sun, Sat
		"2020-06-06", "2021-06-06", "2027-06-06", // Sat, Sun, Sun
	} {
		for _, h := range On(parse(t, d).AddDate(0, 0, 1)) {
			if h.Kind == Substitute {
				t.Errorf("%s must not produce a substitute, got %s", d, h)
			}
		}
		for _, h := range On(parse(t, d).AddDate(0, 0, 2)) {
			if h.Kind == Substitute {
				t.Errorf("%s must not produce a substitute, got %s", d, h)
			}
		}
	}
}

func TestSeollalChuseokSaturdayIsNotLost(t *testing.T) {
	// 제3조제1항제2호: the lunar blocks are substituted for Sunday only.
	if got := ofKind(mustYear(t, 2026), Substitute); contains(got, "2026-09-28") {
		t.Errorf("추석 다음 날 2026-09-26 (Sat) must not be substituted: %v", got)
	}
	// 2020: 설날 on Saturday 01-25 is not lost; Sunday 01-26 is → one substitute.
	if got := ofKind(mustYear(t, 2020), Substitute); !equal(got, []string{"2020-01-27"}) {
		t.Errorf("2020: %v", got)
	}
}

func TestSameDayOverlap(t *testing.T) {
	// 2025-05-05: 어린이날 and 부처님 오신 날 on one Monday → one substitute, as gazetted.
	got := ofKind(mustYear(t, 2025), Substitute)
	if !equal(got, []string{"2025-03-03", "2025-05-06", "2025-10-08"}) {
		t.Errorf("2025: %v", got)
	}
	// 2039: 추석 on Sunday 10-02 and 추석 다음 날 + 개천절 on Monday 10-03 — two
	// lost dates, and 제3조제2항 pushes the second substitute to 10-05.
	got = ofKind(mustYear(t, 2039), Substitute)
	if !contains(got, "2039-10-04") || !contains(got, "2039-10-05") {
		t.Errorf("2039: want 10-04 and 10-05, got %v", got)
	}
	// 2028: 개천절 falls inside the 추석 block on a Tuesday → 10-05, the same
	// shape as 2017-10-06. Nothing else in 2028 is lost.
	if got = ofKind(mustYear(t, 2028), Substitute); !equal(got, []string{"2028-10-05"}) {
		t.Errorf("2028: %v", got)
	}
}

func TestEligibilityFollowsTheAmendmentDates(t *testing.T) {
	cases := []struct {
		date string
		sub  bool
		why  string
	}{
		{"2020-03-01", false, "3·1절 on a Sunday before 2021-08-04"},
		{"2021-08-15", true, "광복절 on a Sunday after 2021-08-04"},
		{"2022-05-08", false, "부처님 오신 날 on a Sunday before 2023-05-04"},
		{"2023-05-27", true, "부처님 오신 날 on a Saturday after 2023-05-04"},
		{"2022-12-25", false, "기독탄신일 on a Sunday before 2023-05-04"},
		{"2027-12-25", true, "기독탄신일 on a Saturday"},
		{"2027-05-01", true, "노동절 on a Saturday"},
		{"2027-07-17", true, "제헌절 on a Saturday"},
	}
	for _, c := range cases {
		d := parse(t, c.date)
		found := false
		for _, h := range mustYear(t, d.Year()) {
			if h.Kind == Substitute && strings.Contains(h.Reason, "("+c.date+")") {
				found = true
			}
		}
		if found != c.sub {
			t.Errorf("%s: substitute=%v, want %v (%s)", c.date, found, c.sub, c.why)
		}
	}
	// 노동절 and 제헌절 do not exist as holidays before 2026.
	for _, h := range mustYear(t, 2025) {
		if h.Name == "노동절" || h.Name == "제헌절" {
			t.Errorf("2025 must not contain %s", h.Name)
		}
	}
}

func TestOnAndIsWorkday(t *testing.T) {
	hs := On(Date(2026, 10, 5))
	if len(hs) != 1 || hs[0].Kind != Substitute || !strings.Contains(hs[0].Reason, "개천절 (2026-10-03) fell on Saturday") {
		t.Errorf("On(2026-10-05) = %v", hs)
	}
	if hs := On(Date(2025, 5, 5)); len(hs) != 2 {
		t.Errorf("On(2025-05-05) should list both holidays, got %v", hs)
	}
	if On(Date(2026, 10, 6)) != nil || !IsWorkday(Date(2026, 10, 6)) {
		t.Error("2026-10-06 is a workday")
	}
	if IsWorkday(Date(2026, 10, 4)) || IsWorkday(Date(2026, 9, 26)) || IsWorkday(Date(2026, 10, 5)) {
		t.Error("Sunday, Saturday and a substitute are not workdays")
	}
	if On(Date(2013, 12, 25)) != nil || !IsWorkday(Date(2013, 12, 25)) {
		t.Error("unsupported years yield no holidays and, being a Wednesday, a workday")
	}
	// A time of day in any zone maps to its civil date.
	seoul, _ := time.LoadLocation("Asia/Seoul")
	if On(time.Date(2026, 10, 5, 23, 59, 0, 0, seoul)) == nil {
		t.Error("On must truncate to the civil date")
	}
}

func TestUnsupportedYear(t *testing.T) {
	for _, y := range []int{2013, 1999, 2200} {
		if _, err := Year(y); err != ErrUnsupportedYear {
			t.Errorf("Year(%d) err = %v", y, err)
		}
	}
}

func TestEveryHolidayHasABasisAndSorted(t *testing.T) {
	for y := MinYear; y <= 2040; y++ {
		hs := mustYear(t, y)
		for i, h := range hs {
			if h.Basis == "" || h.Name == "" || h.NameEN == "" {
				t.Errorf("%d: incomplete holiday %+v", y, h)
			}
			if h.Kind != Fixed && h.Kind != Lunar && h.Reason == "" {
				t.Errorf("%d: %s needs a reason", y, h)
			}
			if i > 0 && h.Date.Before(hs[i-1].Date) {
				t.Errorf("%d: not sorted at %s", y, h)
			}
		}
	}
}

func contains(xs []string, x string) bool {
	for _, s := range xs {
		if s == x {
			return true
		}
	}
	return false
}

func parse(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// 제3조제3항 — a substitute that would land on Saturday moves on. No real
// year up to 2199 exercises it (a weekday overlap only ever happens inside
// the 추석 block or on 어린이날, and the following Saturday is never free),
// so it is pinned on a synthetic pair of holidays on a Friday.
func TestSubstituteSkipsSaturday(t *testing.T) {
	fri := Date(2026, 7, 10)
	in := []Holiday{
		{Date: fri, Name: "어린이날", Kind: Fixed},
		{Date: fri, Name: "부처님 오신 날", Kind: Lunar},
	}
	got := substitutes(in)
	if len(got) != 1 || !got[0].Date.Equal(Date(2026, 7, 13)) {
		t.Errorf("want one substitute on Monday 2026-07-13, got %v", got)
	}
}
