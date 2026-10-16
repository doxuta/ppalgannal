package main

import (
	"os/exec"
	"strings"
	"testing"
)

func run(t *testing.T, arg string) (string, error) {
	t.Helper()
	out, err := exec.Command("go", "run", ".", arg).Output()
	return string(out), err
}

// A date that time.Parse rejects must not fall through to fmt.Sscanf("%d"),
// which reads the leading digits and prints the whole year with exit 0.
// README: usage is "ppalgannal [-json] [YYYY | YYYY-MM-DD]".
func TestMalformedDateIsRefused(t *testing.T) {
	for _, arg := range []string{"2026-1-5", "2026-02-30", "2026-13-01", "2026abc"} {
		out, err := run(t, arg)
		if err == nil {
			t.Errorf("ppalgannal %s: exit 0 with %d lines of output, want a usage error",
				arg, strings.Count(out, "\n"))
		}
	}
}

// README, Limitations: "Years before 2014 are refused." The year form is
// (ppalgannal 2013 exits 1), the date form is not: 2013-12-25 prints
// "2013-12-25 Wed workday" and exits 0, because On swallows ErrUnsupportedYear.
func TestDateOutsideSupportedYearsIsRefused(t *testing.T) {
	for _, arg := range []string{"2013-12-25", "2200-01-01"} {
		out, err := run(t, arg)
		if err == nil {
			t.Errorf("ppalgannal %s: exit 0 with %q, want a refusal", arg, strings.TrimSpace(out))
		}
	}
}
