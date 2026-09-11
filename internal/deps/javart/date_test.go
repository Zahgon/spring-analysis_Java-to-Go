package javart_test

import (
	"testing"
	"time"

	"github.com/seaswalker/spring-analysis/internal/deps/javart"
)

// TestDateToString covers the format SimpleModel.toString prints a bound date
// in, and the "null" an absent one prints as.
func TestDateToString(t *testing.T) {
	if got := javart.DateToString(nil); got != "null" {
		t.Errorf("DateToString(nil) = %q, want %q", got, "null")
	}
	moment := time.Date(2020, time.January, 2, 3, 4, 5, 0, time.UTC)
	if got, want := javart.DateToString(&moment), "Thu Jan 02 03:04:05 UTC 2020"; got != want {
		t.Errorf("DateToString = %q, want %q", got, want)
	}
}
