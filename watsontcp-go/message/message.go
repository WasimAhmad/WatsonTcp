package message

import "time"

// MessageStatus represents the status of a WatsonTcp message.
type MessageStatus string

const (
	StatusNormal         MessageStatus = "Normal"
	StatusSuccess        MessageStatus = "Success"
	StatusFailure        MessageStatus = "Failure"
	StatusAuthRequired   MessageStatus = "AuthRequired"
	StatusAuthRequested  MessageStatus = "AuthRequested"
	StatusAuthSuccess    MessageStatus = "AuthSuccess"
	StatusAuthFailure    MessageStatus = "AuthFailure"
	StatusRemoved        MessageStatus = "Removed"
	StatusShutdown       MessageStatus = "Shutdown"
	StatusHeartbeat      MessageStatus = "Heartbeat"
	StatusTimeout        MessageStatus = "Timeout"
	StatusRegisterClient MessageStatus = "RegisterClient"
)

// Message mirrors the WatsonMessage class used in the C# implementation.
type Message struct {
	ContentLength    int64          `json:"len"`
	PresharedKey     []byte         `json:"psk,omitempty"`
	Status           MessageStatus  `json:"status"`
	Metadata         map[string]any `json:"md,omitempty"`
	SyncRequest      bool           `json:"syncreq"`
	SyncResponse     bool           `json:"syncresp"`
	TimestampUtc     time.Time      `json:"ts"`
	ExpirationUtc    *time.Time     `json:"exp,omitempty"`
	ConversationGUID string         `json:"convguid"`
	SenderGUID       string         `json:"-"`
}
