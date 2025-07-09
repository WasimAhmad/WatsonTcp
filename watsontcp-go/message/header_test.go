package message

import (
	"bytes"
	"testing"
)

func TestBuildParseHeader(t *testing.T) {
	msg := &Message{Status: StatusNormal, ContentLength: 5}
	hdr, err := BuildHeader(msg)
	if err != nil {
		t.Fatalf("build header: %v", err)
	}
	if !bytes.HasSuffix(hdr, []byte{'\r', '\n', '\r', '\n'}) {
		t.Fatalf("header missing delimiter")
	}
	parsed, err := ParseHeader(bytes.NewReader(hdr))
	if err != nil {
		t.Fatalf("parse header: %v", err)
	}
	if parsed.Status != msg.Status || parsed.ContentLength != msg.ContentLength {
		t.Fatalf("parsed message mismatch")
	}
}

func TestParseHeaderPeerDisconnect(t *testing.T) {
	data := []byte("\x00\x00\x00\x00")
	_, err := ParseHeader(bytes.NewReader(data))
	if err == nil {
		t.Fatalf("expected error")
	}
}
