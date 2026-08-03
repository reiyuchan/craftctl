package server

import "testing"

func TestParseTPSLine(t *testing.T) {
	cases := []struct {
		line string
		want float64
		ok   bool
	}{
		{"TPS from last 1m, 5m, 15m: 20.0, 20.0, 20.0", 20.0, true},
		{"TPS from last 1m, 5m, 15m: 19.5, 19.8, 20.0", 19.5, true},
		{"TPS from last 1m, 5m, 15m: 0.0, 0.0, 0.0", 0.0, true},
		{"[12:00:00 INFO]: TPS from last 1m, 5m, 15m: 20.0, 20.0, 20.0", 20.0, true},
		{"[12:00:00 INFO]: TPS from last 1m, 5m, 15m: 14.1, 15.2, 16.3", 14.1, true},
		{"There are 0/20 players online", 0, false},
		{"TPS: nothing", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		got, ok := parseTPSLine(c.line)
		if ok != c.ok {
			t.Errorf("parseTPSLine(%q) ok=%v, want %v", c.line, ok, c.ok)
			continue
		}
		if ok && got != c.want {
			t.Errorf("parseTPSLine(%q) = %v, want %v", c.line, got, c.want)
		}
	}
}
