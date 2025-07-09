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
	OnStream     func(id string, msg *message.Message, r io.Reader)
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

	maxConnections int
	permittedIPs   []string
	blockedIPs     []string

	done chan struct{}
}

func (s *Server) logf(format string, args ...any) {
	if s.options.Logger != nil && s.options.DebugMessages {
		s.options.Logger(format, args...)
	}
}

type clientConn struct {
	conn       net.Conn
	lastActive time.Time
	mu         sync.Mutex
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
		Addr:           addr,
		TLSConfig:      tlsConf,
		callbacks:      cb,
		options:        *opts,
		stats:          stats.New(),
		conns:          make(map[string]*clientConn),
		idleTimeout:    opts.IdleTimeout,
		checkInterval:  opts.CheckInterval,
		maxConnections: opts.MaxConnections,
		permittedIPs:   opts.PermittedIPs,
		blockedIPs:     opts.BlockedIPs,
		done:           make(chan struct{}),
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
		remoteHost, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		if !s.ipAllowed(remoteHost) {
			conn.Close()
			continue
		}
		s.mu.Lock()
		if s.maxConnections > 0 && len(s.conns) >= s.maxConnections {
			s.mu.Unlock()
			conn.Close()
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
	if s.options.PresharedKey != "" {
		msg, err := message.ParseHeader(c.conn)
		if err != nil {
			return
		}
		payload := make([]byte, msg.ContentLength)
		if _, err := io.ReadFull(c.conn, payload); err != nil {
			return
		}
		if msg.Status != message.StatusAuthRequested || string(msg.PresharedKey) != s.options.PresharedKey {
			resp := &message.Message{Status: message.StatusAuthFailure}
			if hdr, err := message.BuildHeader(resp); err == nil {
				c.conn.Write(hdr)
			}
			return
		}
		resp := &message.Message{Status: message.StatusAuthSuccess}
		if hdr, err := message.BuildHeader(resp); err == nil {
			c.conn.Write(hdr)
		}
	}
	for {
		msg, err := message.ParseHeader(c.conn)
		if err != nil {
			if err != io.EOF && s.callbacks.OnDisconnect != nil {
				// connection error
			}
			return
		}
		s.logf("received from %s: %+v", id, msg)
		if s.callbacks.OnStream != nil && s.callbacks.OnMessage == nil {
			lr := &io.LimitedReader{R: c.conn, N: msg.ContentLength}
			s.stats.IncrementReceivedMessages()
			s.stats.AddReceivedBytes(msg.ContentLength)
			s.mu.Lock()
			c.lastActive = time.Now()
			s.mu.Unlock()
			s.callbacks.OnStream(id, msg, lr)
			if lr.N > 0 {
				io.CopyN(io.Discard, c.conn, lr.N)
			}
		} else {
			payload := make([]byte, msg.ContentLength)
			if _, err := io.ReadFull(c.conn, payload); err != nil {
				return
			}
			s.logf("received %d bytes from %s", len(payload), id)
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
}

func (s *Server) SendStream(id string, msg *message.Message, r io.Reader, length int64) error {
	if r == nil {
		return errors.New("reader nil")
	}
	s.mu.Lock()
	c := s.conns[id]
	s.mu.Unlock()
	if c == nil {
		return errors.New("unknown client")
	}
	s.logf("sending to %s: %+v length=%d", id, msg, length)
	msg.ContentLength = length
	msg.TimestampUtc = time.Now().UTC()
	header, err := message.BuildHeader(msg)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	if length > 0 {
		if _, err := io.CopyN(c.conn, r, length); err != nil {
			return err
		}
	}
	s.stats.IncrementSentMessages()
	s.stats.AddSentBytes(int64(len(header)) + length)
	s.logf("sent %d bytes to %s", int64(len(header))+length, id)
	return nil
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

func (s *Server) ipAllowed(addr string) bool {
	ip := net.ParseIP(addr)
	if ip == nil {
		return false
	}
	for _, b := range s.blockedIPs {
		if ipMatch(ip, b) {
			return false
		}
	}
	if len(s.permittedIPs) > 0 {
		for _, p := range s.permittedIPs {
			if ipMatch(ip, p) {
				return true
			}
		}
		return false
	}
	return true
}

func ipMatch(ip net.IP, pattern string) bool {
	if ip2 := net.ParseIP(pattern); ip2 != nil {
		return ip.Equal(ip2)
	}
	if _, netw, err := net.ParseCIDR(pattern); err == nil && netw != nil {
		return netw.Contains(ip)
	}
	return false
}
