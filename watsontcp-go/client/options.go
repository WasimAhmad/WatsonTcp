package client

import "time"

// Options mirrors a subset of the configuration options exposed in the C#
// implementation for WatsonTcp clients.
type Options struct {
	// ConnectTimeout defines how long to wait when establishing a connection.
	ConnectTimeout time.Duration

	// IdleTimeout specifies the period of inactivity after which the
	// connection will be closed. Zero disables idle timeouts.
	IdleTimeout time.Duration

	// EvaluationInterval is the interval at which idle timeouts are evaluated.
	EvaluationInterval time.Duration

	// KeepAlive defines TCP keepalive behavior.
	KeepAlive KeepAlive

	// PresharedKey is required by the server for authentication.
	PresharedKey string

	// Logger is used when DebugMessages is true to output debug logs around
	// send and receive operations. The function should behave like
	// fmt.Printf.
	Logger func(format string, args ...any)

	// DebugMessages enables logging of send and receive operations when a
	// Logger is provided.
	DebugMessages bool
}

// KeepAlive mirrors WatsonTcp keepalive settings.
type KeepAlive struct {
	Enable     bool
	Interval   time.Duration
	Time       time.Duration
	RetryCount int
}

// DefaultOptions provides sensible defaults matching the original
// client implementation.
func DefaultOptions() Options {
	return Options{
		ConnectTimeout:     5 * time.Second,
		IdleTimeout:        0,
		EvaluationInterval: 1 * time.Second,
		KeepAlive: KeepAlive{
			Enable:     false,
			Interval:   5 * time.Second,
			Time:       5 * time.Second,
			RetryCount: 5,
		},
		Logger:        nil,
		DebugMessages: false,
	}
}
