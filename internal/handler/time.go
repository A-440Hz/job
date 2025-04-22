package handler

import (
	"job/internal/scheduler"
	"net/http"
	"strconv"
)

// PollUserTimezone returns the user's time.Location or the DefaultTimezone on failure.
// JS time.Date() has a .getTimezoneOffset() method that returns the offset in minutes to convert local time to UTC.
// we can store it as "X-user-timezone" in a http header and use it to convert to a storable Timezone object.
func PollUserTimezone(r *http.Request) *scheduler.Timezone {
	invOffset := r.Header.Get("X-user-timezone")
	if invOffset == "" || invOffset == "NaN" {
		return scheduler.GetDefaultTimezone()
	}
	io, err := strconv.Atoi(invOffset)
	if err != nil {
		return scheduler.GetDefaultTimezone()
	}
	// convert to seconds
	return scheduler.NewTimezone(io * 60)
}
