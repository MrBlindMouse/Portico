package runtime

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// ServerConfig configures the SSH server.
type ServerConfig struct {
	Addr         string
	DataDir      string
	HostKeyPath  string
	IdleTimeout   time.Duration
	MaxConns      int
	MaxConnsPerIP int
	Logger        *slog.Logger
	Accounts     *AccountConfig
	AllowKeys    []string
	DenyKeys     []string
}

// AccountConfig enables password authentication (phase 8).
type AccountConfig struct {
	Enabled  bool
	Users    map[string]string // username -> password (dev only; hash in production)
	Accounts *AccountStore
}

// DefaultConfig returns dev-friendly defaults.
func DefaultConfig() ServerConfig {
	return ServerConfig{
		Addr:        ":2222",
		IdleTimeout: 30 * time.Minute,
		MaxConns:    64,
		Logger:      slog.Default(),
	}
}

// Server serves an experience over SSH.
type Server struct {
	cfg       ServerConfig
	engine    *Engine
	ssh       *ssh.Server
	mu        sync.Mutex
	conns     int
	connsByIP map[string]int
}

// NewServer creates an SSH server for the engine.
func NewServer(cfg ServerConfig, engine *Engine) (*Server, error) {
	if engine == nil {
		return nil, fmt.Errorf("portico: engine is required")
	}
	s := &Server{cfg: cfg, engine: engine}
	hostKey, err := s.loadOrGenerateHostKey()
	if err != nil {
		return nil, err
	}
	s.ssh = &ssh.Server{
		Addr:    cfg.Addr,
		Handler: s.handleSession,
	}
	s.ssh.AddHostKey(hostKey)
	if cfg.Accounts != nil && cfg.Accounts.Enabled {
		s.ssh.PasswordHandler = s.passwordHandler
	}
	return s, nil
}

func (s *Server) loadOrGenerateHostKey() (gossh.Signer, error) {
	path := s.cfg.HostKeyPath
	if path == "" && s.cfg.DataDir != "" {
		path = filepath.Join(s.cfg.DataDir, "ssh_host_rsa")
	}
	if path != "" {
		if b, err := os.ReadFile(path); err == nil {
			return gossh.ParsePrivateKey(b)
		}
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	if path != "" {
		_ = os.MkdirAll(filepath.Dir(path), 0o750)
		pemBytes := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		})
		_ = os.WriteFile(path, pemBytes, 0o600)
	}
	return gossh.NewSignerFromKey(key)
}

func (s *Server) passwordHandler(ctx ssh.Context, pass string) bool {
	if s.cfg.Accounts == nil || !s.cfg.Accounts.Enabled {
		return false
	}
	user := ctx.User()
	if s.cfg.Accounts.Users != nil {
		if want, ok := s.cfg.Accounts.Users[user]; ok && want == pass {
			return true
		}
	}
	if s.cfg.Accounts.Accounts != nil {
		return s.cfg.Accounts.Accounts.Verify(user, pass)
	}
	return false
}

func (s *Server) handleSession(sess ssh.Session) {
	remote := sess.RemoteAddr().String()
	if !s.acquireConn(remote) {
		s.cfg.Logger.Warn("connection_rejected", "reason", "limit", "remote", remote)
		_, _ = ioWriteString(sess, "connection limit reached\n")
		return
	}
	defer s.releaseConn(remote)

	var idn identity.Identity
	sshUser := sess.User()
	if sshUser != "" && s.cfg.Accounts != nil && s.cfg.Accounts.Enabled {
		idn = identity.FromAccount(sshUser)
	} else if pk := sess.PublicKey(); pk != nil {
		idn = identity.FromPublicKey(pk)
		if s.denied(idn.Fingerprint()) {
			s.cfg.Logger.Warn("connection_denied", "fingerprint", idn.Fingerprint(), "remote", remote)
			_, _ = ioWriteString(sess, "access denied\n")
			return
		}
	} else {
		idn = identity.Anonymous()
	}

	id := s.engine.NewSessionID()
	root := s.engine.Experience().Root()
	child := newSession(id, idn, root)
	s.engine.RegisterSession(child)
	s.cfg.Logger.Info("session_connected", "session", id, "identity", idn.String(), "remote", sess.RemoteAddr().String())
	defer func() {
		for _, sp := range child.Spaces() {
			_, _ = s.engine.Applier().Apply(child, s.engine.Experience(), []experience.Effect{
				experience.LeaveSpaceEffect(sp),
			})
		}
		s.engine.UnregisterSession(id)
		s.cfg.Logger.Info("session_disconnected", "session", id, "identity", idn.String())
	}()

	term := NewSSHTerminal(sess)
	term.OnResize(func(w, h int) {
		_ = s.engine.Render(child, term)
	})

	var activity chan struct{}
	if s.cfg.IdleTimeout > 0 {
		activity = make(chan struct{}, 1)
		go s.idleWatch(sess.Context(), sess, activity)
	}

	_ = s.engine.RunLoop(child, term, activity)
}

func (s *Server) acquireConn(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg.MaxConns > 0 && s.conns >= s.cfg.MaxConns {
		return false
	}
	if s.cfg.MaxConnsPerIP > 0 {
		if s.connsByIP == nil {
			s.connsByIP = make(map[string]int)
		}
		if s.connsByIP[host] >= s.cfg.MaxConnsPerIP {
			return false
		}
		s.connsByIP[host]++
	}
	s.conns++
	return true
}

func (s *Server) releaseConn(remote string) {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns--
	if s.connsByIP != nil && s.cfg.MaxConnsPerIP > 0 {
		s.connsByIP[host]--
		if s.connsByIP[host] <= 0 {
			delete(s.connsByIP, host)
		}
	}
}

func (s *Server) denied(fp string) bool {
	for _, d := range s.cfg.DenyKeys {
		if d == fp {
			return true
		}
	}
	if len(s.cfg.AllowKeys) == 0 {
		return false
	}
	for _, a := range s.cfg.AllowKeys {
		if a == fp {
			return false
		}
	}
	return true
}

func (s *Server) idleWatch(ctx ssh.Context, sess ssh.Session, activity <-chan struct{}) {
	t := time.NewTimer(s.cfg.IdleTimeout)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-activity:
			if !t.Stop() {
				select {
				case <-t.C:
				default:
				}
			}
			t.Reset(s.cfg.IdleTimeout)
		case <-t.C:
			_ = sess.Close()
			return
		}
	}
}

// ListenAndServe starts the SSH server and blocks until ctx is done or error.
func (s *Server) ListenAndServe(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return err
	}
	return s.ListenAndServeOn(ctx, ln)
}

// ListenAndServeOn serves on an existing listener (tests).
func (s *Server) ListenAndServeOn(ctx context.Context, ln net.Listener) error {
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()
	s.cfg.Logger.Info("portico listening", "addr", ln.Addr().String())
	err := s.ssh.Serve(ln)
	if errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}

// EngineRef returns the underlying engine.
func (s *Server) Engine() *Engine { return s.engine }

// ReloadExperience validates and swaps the experience.
func (s *Server) ReloadExperience(b *experience.Builder) error {
	exp, err := b.Build()
	if err != nil {
		return err
	}
	return s.engine.SetExperience(exp)
}

func ioWriteString(w ioStringWriter, v string) (int, error) {
	return w.Write([]byte(v))
}

type ioStringWriter interface {
	Write([]byte) (int, error)
}
