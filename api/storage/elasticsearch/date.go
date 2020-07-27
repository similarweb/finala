package elasticsearch

import "time"

// getDayAfterDate returns the starte tomorrow date.
// Example: dt = 1.1.2020 10:00:00
// return : 1.2.2020 00:00:00
func getDayAfterDate(dt time.Time, zone *time.Location) time.Time {
	dt = dt.AddDate(0, 0, 1)
	return time.Date(dt.Year(), dt.Month(), dt.Day(), 00, 00, 0, 0, zone)
}
