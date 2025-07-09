package server

import "time"

// Options mirrors a subset of the configuration options available to the C#
// WatsonTcp server implementation.
type Options struct {
	// IdleTimeout is the amount of time a connection can remain idle before it
	// is terminated. Zero disables the check.
	IdleTimeout time.Duration

	// CheckInterval controls how often idle connections are evaluated.
	CheckInterval time.Duration

	// KeepAlive defines TCP keepalive behavior.
	KeepAlive KeepAlive

	// PresharedKey expected from clients.
	PresharedKey string
}

// KeepAlive mirrors WatsonTcp keepalive settings.
type KeepAlive struct {
	Enable     bool
	Interval   time.Duration
	Time       time.Duration
	RetryCount int
}

// DefaultOptions returns Options with the same defaults used in the current
// Go server implementation.
func DefaultOptions() Options {
	return Options{
		IdleTimeout:   30 * time.Second,
		CheckInterval: 5 * time.Second,
		KeepAlive: KeepAlive{
			Enable:     false,
			Interval:   5 * time.Second,
			Time:       5 * time.Second,
			RetryCount: 5,
		},
	}
}
