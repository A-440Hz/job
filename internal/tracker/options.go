package tracker

import "time"

type Option func(*options)

type options struct {
	withQuantity int
	withDeadline time.Time
	withFrequency
}

func getOpts(opt ...Option) options {
	opts := getDefaultOptions()
	for _, o := range opt {
		o(&opts)
	}
	return opts
}

// getDefaultJobAppOptions should be called when initializing the JobAppTracker
func getDefaultJobAppOptions() options {
	return options{
		withQuantity: 5,
	}
}
