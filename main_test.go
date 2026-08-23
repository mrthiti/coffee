package main

import (
	"testing"
	"time"
)

func TestParseTimeout(t *testing.T) {
	cases := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		{"3h", 3 * time.Hour, false},
		{"5m", 5 * time.Minute, false},
		{"3h30m", 3*time.Hour + 30*time.Minute, false},
		{"", 0, true},
		{"bogus", 0, true},
		{"-1h", 0, true},
	}

	for _, c := range cases {
		got, err := parseTimeout(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseTimeout(%q) = %v, want error", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseTimeout(%q) returned unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseTimeout(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
