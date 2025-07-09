package server

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/yourname/watsontcp-go/message"
	"github.com/yourname/watsontcp-go/stats"
)

type Callbacks struct {
	OnConnect    func(id string, conn net.Conn)
	OnDisconnect func(id string)
	OnMessage    func(id string, msg *message.Message, data []byte)
}

type Server struct {
	Addr      string
	TLSConfig *tls.Config

	callbacks Callbacks

	options Options
	stats   *stats.Statistics

	listener net.Listener
	conns    map[string]*clientConn
	mu       sync.Mutex

	idleTimeout   time.Duration
	checkInterval time.Duration

	done chan struct{}
}

type clientConn struct {
	conn       net.Conn
	lastActive time.Time
}

// Statistics returns runtime counters for the server.
func (s *Server) Statistics() *stats.Statistics {
	return s.stats
}

func New(addr string, tlsConf *tls.Config, cb Callbacks, opts *Options) *Server {
	if opts == nil {
		defaultOpts := DefaultOptions()
		opts = &defaultOpts
	}
	return &Server{
		Addr:          addr,
		TLSConfig:     tlsConf,
		callbacks:     cb,
		options:       *opts,
		stats:         stats.New(),
		conns:         make(map[string]*clientConn),
		idleTimeout:   opts.IdleTimeout,
		checkInterval: opts.CheckInterval,
		done:          make(chan struct{}),
	}
}

func (s *Server) Start() error {
	if s.listener != nil {
		return errors.New("server already started")
	}
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	if s.TLSConfig != nil {
		ln = tls.NewListener(ln, s.TLSConfig)
	}
	s.listener = ln
	go s.acceptLoop()
	go s.monitorLoop()
	return nil
}

func (s *Server) Stop() {
	close(s.done)
	if s.listener != nil {
		s.listener.Close()
	}
	s.mu.Lock()
	for id, c := range s.conns {
		c.conn.Close()
		delete(s.conns, id)
	}
	s.mu.Unlock()
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
			}
			continue
		}
		if s.options.KeepAlive.Enable {
			if tcp, ok := conn.(*net.TCPConn); ok {
				tcp.SetKeepAlive(true)
				if s.options.KeepAlive.Interval > 0 {
					tcp.SetKeepAlivePeriod(s.options.KeepAlive.Interval)
				}
			}
		}
		id := conn.RemoteAddr().String()
		s.mu.Lock()
		s.conns[id] = &clientConn{conn: conn, lastActive: time.Now()}
		s.mu.Unlock()
		if s.callbacks.OnConnect != nil {
			go s.callbacks.OnConnect(id, conn)
		}
		go s.handleConn(id)
	}
}

func (s *Server) handleConn(id string) {
	c := func() *clientConn {
		s.mu.Lock()
		defer s.mu.Unlock()
		return s.conns[id]
	}()
	if c == nil {
		return
	}
	defer func() {
		c.conn.Close()
		s.mu.Lock()
		delete(s.conns, id)
		s.mu.Unlock()
		if s.callbacks.OnDisconnect != nil {
			s.callbacks.OnDisconnect(id)
		}
	}()
	for {
		msg, err := message.ParseHeader(c.conn)
		if err != nil {
			if err != io.EOF && s.callbacks.OnDisconnect != nil {
				// connection error
			}
			return
		}
		payload := make([]byte, msg.ContentLength)
		if _, err := io.ReadFull(c.conn, payload); err != nil {
			return
		}
		s.stats.IncrementReceivedMessages()
		s.stats.AddReceivedBytes(int64(len(payload)))
		s.mu.Lock()
		c.lastActive = time.Now()
		s.mu.Unlock()
		if s.callbacks.OnMessage != nil {
			s.callbacks.OnMessage(id, msg, payload)
		}
	}
}

func (s *Server) monitorLoop() {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			var toClose []string
			s.mu.Lock()
			for id, c := range s.conns {
				if now.Sub(c.lastActive) > s.idleTimeout {
					toClose = append(toClose, id)
				}
			}
			s.mu.Unlock()
			for _, id := range toClose {
				s.mu.Lock()
				c := s.conns[id]
				s.mu.Unlock()
				if c != nil {
					c.conn.Close()
				}
			}
		case <-s.done:
			return
		}
	}
}
