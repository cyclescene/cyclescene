package main

import "testing"

func TestTodayForCity(t *testing.T) {
	for _, city := range []string{"la", "pdx", "slc"} {
		if got := todayForCity(city); len(got) != len("2006-01-02") {
			t.Fatalf("todayForCity(%q) = %q, want YYYY-MM-DD", city, got)
		}
	}
}
