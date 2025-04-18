package handler

import (
	"job/internal/scheduler"
	"net/http"
	"strconv"
	"time"
)

// PollUserTimezone returns the user's time.Location or the DefaultTimezone on failure.
// js time.Date() has a .getTimezoneOffset() method that returns the offset in minutes to convert local time to UTC
// we can store it in "X-user-timezone" in a http header and use it to convert to a time.Location
func PollUserTimezone(r *http.Request) *time.Location {
	s := r.Header.Get("X-user-timezone")
	if s == "" || s == "NaN" {
		return scheduler.DefaultTimezone
	}
	offset, err := strconv.Atoi(s)
	if err != nil {
		return scheduler.DefaultTimezone
	}
	return time.FixedZone("UTC", -offset*60)
}
