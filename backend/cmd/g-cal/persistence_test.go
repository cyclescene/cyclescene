package main

import "testing"

func TestImportDateRangeForCity(t *testing.T) {
	for _, city := range []string{"la", "pdx", "slc"} {
		start, end := importDateRangeForCity(city, 14)
		if len(start) != len("2006-01-02") || len(end) != len("2006-01-02") {
			t.Fatalf("importDateRangeForCity(%q) = (%q, %q), want YYYY-MM-DD values", city, start, end)
		}
	}
}
