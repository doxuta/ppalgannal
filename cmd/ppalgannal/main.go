// Command ppalgannal prints South Korea's public holidays for a year, or
// explains a single date.
//
//	ppalgannal            # this year, Asia/Seoul
//	ppalgannal 2027
//	ppalgannal 2026-10-05
//	ppalgannal -json 2026
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/doxuta/ppalgannal"
)

type row struct {
	Date   string `json:"date"`
	Name   string `json:"name"`
	NameEN string `json:"name_en"`
	Kind   string `json:"kind"`
	Basis  string `json:"basis"`
	Reason string `json:"reason,omitempty"`
}

func toRows(hs []ppalgannal.Holiday) []row {
	out := make([]row, 0, len(hs))
	for _, h := range hs {
		out = append(out, row{ppalgannal.ISO(h.Date), h.Name, h.NameEN, h.Kind.String(), h.Basis, h.Reason})
	}
	return out
}

func main() {
	asJSON := flag.Bool("json", false, "emit JSON")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: ppalgannal [-json] [YYYY | YYYY-MM-DD]")
	}
	flag.Parse()

	seoul, _ := time.LoadLocation("Asia/Seoul")
	arg := fmt.Sprint(time.Now().In(seoul).Year())
	if flag.NArg() > 0 {
		arg = flag.Arg(0)
	}

	if d, err := time.Parse("2006-01-02", arg); err == nil {
		// On reports "no holiday" for a year outside the supported window;
		// the CLI should refuse it like the year form does.
		if _, err := ppalgannal.Year(d.Year()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		hs := ppalgannal.On(d)
		if *asJSON {
			json.NewEncoder(os.Stdout).Encode(toRows(hs))
			return
		}
		switch {
		case len(hs) > 0:
			for _, h := range hs {
				fmt.Println(h)
			}
		case ppalgannal.IsWorkday(d):
			fmt.Printf("%s %s workday\n", arg, d.Weekday().String()[:3])
		default:
			fmt.Printf("%s %s weekend\n", arg, d.Weekday().String()[:3])
		}
		return
	}

	// A malformed date (2026-1-5, 2026-02-30) must not be read as the year
	// 2026: Atoi, unlike Sscanf, rejects trailing text.
	y, err := strconv.Atoi(arg)
	if err != nil {
		flag.Usage()
		os.Exit(2)
	}
	hs, err := ppalgannal.Year(y)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *asJSON {
		json.NewEncoder(os.Stdout).Encode(toRows(hs))
		return
	}
	for _, h := range hs {
		fmt.Println(h)
	}
}
