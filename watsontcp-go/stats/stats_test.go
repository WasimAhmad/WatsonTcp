package stats

import "testing"

func TestMessageSizeAverages(t *testing.T) {
	s := New()
	s.AddReceivedBytes(50)
	s.IncrementReceivedMessages()
	s.AddReceivedBytes(100)
	s.IncrementReceivedMessages()
	s.AddSentBytes(30)
	s.IncrementSentMessages()
	s.AddSentBytes(70)
	s.IncrementSentMessages()

	if s.ReceivedMessageSizeAverage() != 75 {
		t.Fatalf("expected avg 75 got %d", s.ReceivedMessageSizeAverage())
	}
	if s.SentMessageSizeAverage() != 50 {
		t.Fatalf("expected avg 50 got %d", s.SentMessageSizeAverage())
	}

	start := s.StartTime()
	s.Reset()
	if s.ReceivedBytes() != 0 || s.SentBytes() != 0 || s.ReceivedMessages() != 0 || s.SentMessages() != 0 {
		t.Fatalf("reset did not clear counters")
	}
	if !s.StartTime().Equal(start) {
		t.Fatalf("reset should not change start time")
	}
}

func TestStringOutput(t *testing.T) {
	s := New()
	out := s.String()
	if len(out) == 0 || out[0] != '-' {
		t.Fatalf("unexpected string output: %q", out)
	}
}
