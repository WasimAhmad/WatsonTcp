package message

import (
	"encoding/json"
	"errors"
	"io"
)

// BuildHeader serializes msg to JSON and appends \r\n\r\n.
func BuildHeader(msg *Message) ([]byte, error) {
	if msg == nil {
		return nil, errors.New("msg nil")
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	data = append(data, '\r', '\n', '\r', '\n')
	return data, nil
}

// ParseHeader reads from r until \r\n\r\n and unmarshals the header.
func ParseHeader(r io.Reader) (*Message, error) {
	if r == nil {
		return nil, errors.New("reader nil")
	}

	header := make([]byte, 0, 24)
	buf := make([]byte, 1)
	for {
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		header = append(header, buf[0])
		if len(header) >= 4 {
			end := header[len(header)-4:]
			if end[0] == 0 && end[1] == 0 && end[2] == 0 && end[3] == 0 {
				return nil, errors.New("null header indicates peer disconnected")
			}
			if end[0] == '\r' && end[1] == '\n' && end[2] == '\r' && end[3] == '\n' {
				break
			}
		}
	}

	jsonPart := header[:len(header)-4]
	var msg Message
	if err := json.Unmarshal(jsonPart, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
