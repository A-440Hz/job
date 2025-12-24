package handler

import (
	"job/internal/scheduler"
	"net/http"
	"strconv"
)

// PollUserTimezone returns the user's time.Location or the DefaultTimezone on failure.
// JS time.Date() has a .getTimezoneOffset() method that returns the offset in minutes to convert local time to UTC.
// we can store it as "X-Timezone-Offset" in a http header and use it to convert to a storable Timezone object.
func PollUserTimezone(r *http.Request) *scheduler.Timezone {
	invOffset := r.Header.Get("X-Timezone-Offset")
	if invOffset == "" || invOffset == "NaN" {
		return scheduler.GetDefaultTimezone()
	}
	io, err := strconv.Atoi(invOffset)
	if err != nil {
		return scheduler.GetDefaultTimezone()
	}
	// convert to seconds: if X-Timezone-Offset is -60, the user timezone is 60 minutes ahead of UTC,
	// so we need to invert it and convert to seconds, to get seconds east of UTC for time.FixedZone input
	return scheduler.NewTimezoneWithOffset(int64(io * -60))
}
