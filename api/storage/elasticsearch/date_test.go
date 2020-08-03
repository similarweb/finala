package elasticsearch

import (
	"testing"
	"time"
)

func TestGetDayAfterDate(t *testing.T) {

	dt := time.Date(2020, 01, 01, 10, 00, 0, 0, time.UTC)
	tomorrow := getDayAfterDate(dt, time.UTC)

	if tomorrow.Year() != 2020 {
		t.Fatalf("unexpected year, got %d expected %d", tomorrow.Year(), 2020)
	}

	if tomorrow.Month() != 01 {
		t.Fatalf("unexpected month, got %d expected %d", tomorrow.Month(), 01)
	}

	if tomorrow.Day() != 02 {
		t.Fatalf("unexpected day, got %d expected %d", tomorrow.Day(), 02)
	}

	if tomorrow.Hour() != 00 {
		t.Fatalf("unexpected hour, got %d expected %d", tomorrow.Hour(), 00)
	}

	if tomorrow.Minute() != 00 {
		t.Fatalf("unexpected Minute, got %d expected %d", tomorrow.Hour(), 00)
	}

}
