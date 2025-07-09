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

	// MaxConnections specifies the maximum number of concurrent
	// connections the server will accept. Zero means unlimited.
	MaxConnections int

	// PermittedIPs is an optional list of IP addresses or CIDR ranges
	// that are allowed to connect. If empty, all clients are permitted
	// unless present in BlockedIPs.
	PermittedIPs []string

	// BlockedIPs specifies IP addresses or CIDR ranges that should be
	// rejected when a client attempts to connect.
	BlockedIPs []string

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
		MaxConnections: 0,
		PermittedIPs:   nil,
		BlockedIPs:     nil,
		Logger:         nil,
		DebugMessages:  false,
	}
}
