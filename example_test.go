package ppalgannal_test

import (
	"fmt"

	"github.com/doxuta/ppalgannal"
)

func ExampleYear() {
	hs, _ := ppalgannal.Year(2026)
	for _, h := range hs {
		if h.Kind == ppalgannal.Substitute {
			fmt.Println(h)
		}
	}
	// Output:
	// 2026-03-02 Mon 대체공휴일 — 3·1절 (2026-03-01) fell on Sunday (제3조제1항제1호)
	// 2026-05-25 Mon 대체공휴일 — 부처님 오신 날 (2026-05-24) fell on Sunday (제3조제1항제1호)
	// 2026-08-17 Mon 대체공휴일 — 광복절 (2026-08-15) fell on Saturday (제3조제1항제1호)
	// 2026-10-05 Mon 대체공휴일 — 개천절 (2026-10-03) fell on Saturday (제3조제1항제1호)
}

func ExampleOn() {
	for _, h := range ppalgannal.On(ppalgannal.Date(2025, 5, 6)) {
		fmt.Println(h)
	}
	fmt.Println(ppalgannal.IsWorkday(ppalgannal.Date(2025, 5, 7)))
	// Output:
	// 2025-05-06 Tue 대체공휴일 — 어린이날 (2025-05-05) coincided with 부처님 오신 날 (제3조제1항제3호)
	// true
}
