package client

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/yourname/watsontcp-go/message"
)

type Callbacks struct {
	OnConnect    func()
	OnDisconnect func()
	OnMessage    func(msg *message.Message, data []byte)
}

type Client struct {
	Addr      string
	TLSConfig *tls.Config

	callbacks Callbacks

	conn    net.Conn
	writeMu sync.Mutex
	respMap sync.Map
	done    chan struct{}
	dcOnce  sync.Once
}

type response struct {
	msg  *message.Message
	data []byte
	err  error
}

func New(addr string, tlsConf *tls.Config, cb Callbacks) *Client {
	return &Client{
		Addr:      addr,
		TLSConfig: tlsConf,
		callbacks: cb,
		done:      make(chan struct{}),
	}
}

func (c *Client) Connect() error {
	if c.conn != nil {
		return errors.New("already connected")
	}
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return err
	}
	if c.TLSConfig != nil {
		tlsConn := tls.Client(conn, c.TLSConfig)
		if err := tlsConn.Handshake(); err != nil {
			conn.Close()
			return err
		}
		conn = tlsConn
	}
	c.conn = conn
	if c.callbacks.OnConnect != nil {
		go c.callbacks.OnConnect()
	}
	go c.readLoop()
	return nil
}

func (c *Client) Disconnect() {
	c.dcOnce.Do(func() {
		close(c.done)
		if c.conn != nil {
			c.conn.Close()
		}
		if c.callbacks.OnDisconnect != nil {
			c.callbacks.OnDisconnect()
		}
	})
}

func (c *Client) Send(msg *message.Message, data []byte) error {
	if c.conn == nil {
		return errors.New("not connected")
	}
	msg.ContentLength = int64(len(data))
	msg.TimestampUtc = time.Now().UTC()
	header, err := message.BuildHeader(msg)
	if err != nil {
		return err
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	if len(data) > 0 {
		if _, err := c.conn.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) SendSync(ctx context.Context, msg *message.Message, data []byte) (*message.Message, []byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	guid := msg.ConversationGUID
	if guid == "" {
		guid = newGUID()
		msg.ConversationGUID = guid
	}
	msg.SyncRequest = true
	ch := make(chan *response, 1)
	c.respMap.Store(guid, ch)
	if err := c.Send(msg, data); err != nil {
		c.respMap.Delete(guid)
		return nil, nil, err
	}
	select {
	case resp := <-ch:
		return resp.msg, resp.data, resp.err
	case <-ctx.Done():
		c.respMap.Delete(guid)
		return nil, nil, ctx.Err()
	}
}

func (c *Client) readLoop() {
	defer c.Disconnect()
	for {
		select {
		case <-c.done:
			return
		default:
		}
		msg, err := message.ParseHeader(c.conn)
		if err != nil {
			if err != io.EOF {
				// handle error
			}
			return
		}
		payload := make([]byte, msg.ContentLength)
		if _, err := io.ReadFull(c.conn, payload); err != nil {
			return
		}
		if msg.SyncResponse && msg.ConversationGUID != "" {
			if val, ok := c.respMap.Load(msg.ConversationGUID); ok {
				ch := val.(chan *response)
				c.respMap.Delete(msg.ConversationGUID)
				ch <- &response{msg: msg, data: payload}
				close(ch)
				continue
			}
		}
		if c.callbacks.OnMessage != nil {
			go c.callbacks.OnMessage(msg, payload)
		}
	}
}

func newGUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
