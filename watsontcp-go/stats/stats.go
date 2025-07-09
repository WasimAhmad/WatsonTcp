package stats

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Statistics tracks counts of bytes and messages sent and received.
type Statistics struct {
	startTime     time.Time
	receivedBytes int64
	receivedMsgs  int64
	sentBytes     int64
	sentMsgs      int64
}

// New creates a new Statistics value with the start time set to now.
func New() *Statistics {
	return &Statistics{startTime: time.Now().UTC()}
}

// StartTime returns the time at which statistics were created.
func (s *Statistics) StartTime() time.Time { return s.startTime }

// UpTime returns the duration since StartTime.
func (s *Statistics) UpTime() time.Duration { return time.Since(s.startTime) }

// ReceivedBytes returns the total bytes received.
func (s *Statistics) ReceivedBytes() int64 { return atomic.LoadInt64(&s.receivedBytes) }

// ReceivedMessages returns the total messages received.
func (s *Statistics) ReceivedMessages() int64 { return atomic.LoadInt64(&s.receivedMsgs) }

// SentBytes returns the total bytes sent.
func (s *Statistics) SentBytes() int64 { return atomic.LoadInt64(&s.sentBytes) }

// SentMessages returns the total messages sent.
func (s *Statistics) SentMessages() int64 { return atomic.LoadInt64(&s.sentMsgs) }

// ReceivedMessageSizeAverage returns the average size in bytes of received messages.
func (s *Statistics) ReceivedMessageSizeAverage() int64 {
	msgs := s.ReceivedMessages()
	if msgs == 0 {
		return 0
	}
	return s.ReceivedBytes() / msgs
}

// SentMessageSizeAverage returns the average size in bytes of sent messages.
func (s *Statistics) SentMessageSizeAverage() int64 {
	msgs := s.SentMessages()
	if msgs == 0 {
		return 0
	}
	return s.SentBytes() / msgs
}

// AddReceivedBytes increments the received byte counter.
func (s *Statistics) AddReceivedBytes(n int64) { atomic.AddInt64(&s.receivedBytes, n) }

// IncrementReceivedMessages increments the received message counter.
func (s *Statistics) IncrementReceivedMessages() { atomic.AddInt64(&s.receivedMsgs, 1) }

// AddSentBytes increments the sent byte counter.
func (s *Statistics) AddSentBytes(n int64) { atomic.AddInt64(&s.sentBytes, n) }

// IncrementSentMessages increments the sent message counter.
func (s *Statistics) IncrementSentMessages() { atomic.AddInt64(&s.sentMsgs, 1) }

// Reset sets counters back to zero preserving the start time.
func (s *Statistics) Reset() {
	atomic.StoreInt64(&s.receivedBytes, 0)
	atomic.StoreInt64(&s.receivedMsgs, 0)
	atomic.StoreInt64(&s.sentBytes, 0)
	atomic.StoreInt64(&s.sentMsgs, 0)
}

// String returns a formatted human-readable representation of the statistics.
func (s *Statistics) String() string {
	return fmt.Sprintf("--- Statistics ---%s    Started     : %s%s    Uptime      : %s%s    Received    : %s       Bytes    : %d%s       Messages : %d%s       Average  : %d bytes%s    Sent        : %s       Bytes    : %d%s       Messages : %d%s       Average  : %d bytes%s",
		"\n", s.startTime.Format(time.RFC3339), "\n",
		s.UpTime().String(), "\n",
		"\n",
		s.ReceivedBytes(), "\n",
		s.ReceivedMessages(), "\n",
		s.ReceivedMessageSizeAverage(), "\n",
		"\n",
		s.SentBytes(), "\n",
		s.SentMessages(), "\n",
		s.SentMessageSizeAverage(), "\n")
}
