package scheduler

import "time"

var defaultTimeZone = time.FixedZone("PST", 0)

func pollUserTimeZone() (*time.Location, error) {
	// do some handler stuff to request X-user-timezone from the client
	// probably relocate this function
	return nil, nil
}

// attempt to return 1am user local time zone, otherwise default to UTC+8
func GetDefaultGoalDeadline() time.Time {
	now := time.Now()
	//pollUserTimeZone()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 1, 0, 0, 0, defaultTimeZone) // UTC+8
}
