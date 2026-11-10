package runtime

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/MrBlindMouse/Portico/components"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/installation"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// passwordAuthKey is set on the SSH context only after password verification succeeds.
type passwordAuthKey struct{}

// ServerConfig configures the SSH server.
type ServerConfig struct {
	Addr              string
	DataDir           string
	HostKeyPath       string
	IdleTimeout       time.Duration
	MaxConns          int
	MaxConnsPerIP     int
	Logger            *slog.Logger
	Accounts          *AccountConfig
	AllowKeys         []string
	DenyKeys          []string
	AllowAccounts     []string
	DenyAccounts      []string
	RequireAccount    bool
	ShutdownGrace     time.Duration
	KeepAliveInterval time.Duration
	KeepAliveCountMax int
}

// AccountConfig enables password authentication (phase 8).
type AccountConfig struct {
	Enabled  bool
	Users    map[string]string // username -> password (dev only)
	Accounts identity.AccountStore
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

// Server serves an installation over SSH.
type Server struct {
	cfg       ServerConfig
	engine    *Engine
	ssh       *ssh.Server
	mu        sync.Mutex
	conns     int
	connsByIP map[string]int

	draining   bool
	sessionsWG sync.WaitGroup
	liveSSH    sync.Map // session id -> ssh.Session
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
		Addr:             cfg.Addr,
		Handler:          s.handleSession,
		PublicKeyHandler: s.publicKeyHandler,
	}
	s.ssh.AddHostKey(hostKey)
	if cfg.Accounts != nil && cfg.Accounts.Enabled {
		s.ssh.PasswordHandler = s.passwordHandler
	}
	// SSH clients probe "none" first and stop when it succeeds, so key and
	// password auth only run when none is disabled.
	if cfg.allowsNoAuth() {
		s.ssh.ServerConfigCallback = func(ssh.Context) *gossh.ServerConfig {
			return &gossh.ServerConfig{NoClientAuth: true}
		}
	}
	return s, nil
}

// allowsNoAuth is the README dev loop: ssh with no key. It is off when a key
// list, password accounts, or RequireAccount is set, because a successful
// "none" probe would skip those methods.
func (cfg ServerConfig) allowsNoAuth() bool {
	if cfg.RequireAccount || len(cfg.AllowKeys) > 0 || len(cfg.DenyKeys) > 0 {
		return false
	}
	if cfg.Accounts != nil && cfg.Accounts.Enabled {
		return false
	}
	return true
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

func (s *Server) publicKeyHandler(_ ssh.Context, key ssh.PublicKey) bool {
	if key == nil {
		return false
	}
	return !s.denied(identity.FromPublicKey(key).Fingerprint())
}

func (s *Server) passwordHandler(ctx ssh.Context, pass string) bool {
	if s.cfg.Accounts == nil || !s.cfg.Accounts.Enabled {
		return false
	}
	user := ctx.User()
	if s.cfg.Accounts.Users != nil {
		if want, ok := s.cfg.Accounts.Users[user]; ok && want == pass {
			s.acceptPassword(ctx, user)
			return true
		}
	}
	if s.cfg.Accounts.Accounts != nil {
		ok, err := s.cfg.Accounts.Accounts.Verify(user, pass)
		if err == nil && ok {
			s.acceptPassword(ctx, user)
			return true
		}
	}
	return false
}

func (s *Server) acceptPassword(ctx ssh.Context, user string) {
	ctx.SetValue(passwordAuthKey{}, true)
	if s.cfg.Logger != nil {
		s.cfg.Logger.Info("account_login", "account", user)
	}
}

// Draining reports whether the server is shutting down.
func (s *Server) Draining() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.draining
}

// ActiveSessions returns the number of tracked SSH sessions.
func (s *Server) ActiveSessions() int {
	n := 0
	s.liveSSH.Range(func(_, _ any) bool {
		n++
		return true
	})
	return n
}

func (s *Server) handleSession(sess ssh.Session) {
	if s.Draining() {
		_, _ = ioWriteString(sess, "server is shutting down\n")
		return
	}
	remote := sess.RemoteAddr().String()
	if !s.acquireConn(remote) {
		s.cfg.Logger.Warn("connection_rejected", "reason", "limit", "remote", remote)
		_, _ = ioWriteString(sess, "connection limit reached\n")
		return
	}
	defer s.releaseConn(remote)

	idn, ok := s.identify(sess)
	if !ok {
		return
	}

	id := s.engine.NewSessionID()
	root := s.engine.Installation().Root()
	child := newSession(id, idn, root)
	s.sessionsWG.Add(1)
	s.liveSSH.Store(id, sess)
	defer func() {
		s.liveSSH.Delete(id)
		s.sessionsWG.Done()
	}()
	s.engine.RegisterSession(child)
	s.cfg.Logger.Info("session_connected", "session", id, "identity", idn.String(), "remote", sess.RemoteAddr().String())
	exp := s.engine.Installation()
	if hooks := exp.OnSessionStart(); len(hooks) > 0 {
		if _, err := s.engine.Applier().Apply(child, exp, hooks); err != nil && s.cfg.Logger != nil {
			s.cfg.Logger.Error("session_start_failed", "session", id, "error", err)
		}
	}
	defer func() {
		exp := s.engine.Installation()
		if hooks := exp.OnSessionEnd(); len(hooks) > 0 {
			if _, err := s.engine.Applier().Apply(child, exp, hooks); err != nil && s.cfg.Logger != nil {
				s.cfg.Logger.Error("session_end_failed", "session", id, "error", err)
			}
		}
		for _, sp := range child.Spaces() {
			if _, err := s.engine.Applier().Apply(child, s.engine.Installation(), []components.Effect{
				components.LeaveSpaceEffect(sp),
			}); err != nil && s.cfg.Logger != nil {
				s.cfg.Logger.Error("session_leave_failed", "session", id, "space", string(sp), "error", err)
			}
		}
		s.engine.UnregisterSession(id)
		if s.cfg.Logger != nil {
			s.cfg.Logger.Info("session_disconnected", "session", id, "identity", idn.String())
		}
	}()
	if s.cfg.KeepAliveInterval > 0 {
		go s.keepAlive(sess)
	}

	term := NewSSHTerminal(sess)
	var lastW, lastH int
	if w, h, ok := term.WindowSize(); ok {
		child.setSize(w, h)
		lastW, lastH = w, h
	}
	term.OnResize(func(w, h int) {
		if w <= 0 || h <= 0 || (w == lastW && h == lastH) {
			return
		}
		lastW, lastH = w, h
		child.setSize(w, h)
		_ = s.engine.Render(child, term)
		term.RestoreInputLine()
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

// identify chooses key, then a verified password, then anonymous.
// A linked fingerprint becomes the account. RequireAccount rejects guests and unlinked keys.
func (s *Server) identify(sess ssh.Session) (identity.Identity, bool) {
	remote := ""
	if sess.RemoteAddr() != nil {
		remote = sess.RemoteAddr().String()
	}
	deny := func(msg string, args ...any) (identity.Identity, bool) {
		if s.cfg.Logger != nil {
			s.cfg.Logger.Warn(msg, args...)
		}
		_, _ = ioWriteString(sess, "access denied\n")
		return identity.Identity{}, false
	}
	if pk := sess.PublicKey(); pk != nil {
		idn := identity.FromPublicKey(pk)
		if s.denied(idn.Fingerprint()) {
			return deny("connection_denied", "fingerprint", idn.Fingerprint(), "remote", remote)
		}
		if s.cfg.Accounts != nil && s.cfg.Accounts.Accounts != nil {
			if rec, ok, _ := s.cfg.Accounts.Accounts.LookupByFingerprint(idn.Fingerprint()); ok {
				idn = identity.FromAccount(rec.Username)
			}
		}
		if s.cfg.RequireAccount && idn.Account() == "" {
			if s.cfg.Logger != nil {
				s.cfg.Logger.Warn("connection_denied", "reason", "require_account", "remote", remote)
			}
			_, _ = ioWriteString(sess, "account required\n")
			return identity.Identity{}, false
		}
		return withSessionLogin(idn, sess.User()), true
	}
	if sess.Context().Value(passwordAuthKey{}) != nil {
		user := sess.User()
		if s.accountDenied(user) {
			return deny("connection_denied", "account", user, "remote", remote)
		}
		return identity.FromAccount(user), true
	}
	if s.cfg.RequireAccount {
		if s.cfg.Logger != nil {
			s.cfg.Logger.Warn("connection_denied", "reason", "require_account", "remote", remote)
		}
		_, _ = ioWriteString(sess, "account required\n")
		return identity.Identity{}, false
	}
	return withSessionLogin(identity.Anonymous(), sess.User()), true
}

// withSessionLogin labels a non-account session with the SSH username.
// The label is not written to the nickname store. Account names stay the account.
func withSessionLogin(idn identity.Identity, user string) identity.Identity {
	if idn.Account() != "" {
		return idn
	}
	user = strings.TrimSpace(user)
	if user == "" {
		return idn
	}
	return idn.WithLogin(user)
}

func (s *Server) keepAlive(sess ssh.Session) {
	t := time.NewTicker(s.cfg.KeepAliveInterval)
	defer t.Stop()
	raw := sess.Context().Value(ssh.ContextKeyConn)
	conn, ok := raw.(gossh.Conn)
	if !ok || conn == nil {
		return
	}
	misses := 0
	for {
		select {
		case <-sess.Context().Done():
			return
		case <-t.C:
			_, _, err := conn.SendRequest("keepalive@openssh.com", true, nil)
			if err != nil {
				misses++
				if s.cfg.KeepAliveCountMax > 0 && misses >= s.cfg.KeepAliveCountMax {
					_ = sess.Close()
					return
				}
				continue
			}
			misses = 0
		}
	}
}

func (s *Server) accountDenied(user string) bool {
	for _, d := range s.cfg.DenyAccounts {
		if d == user {
			return true
		}
	}
	if len(s.cfg.AllowAccounts) == 0 {
		return false
	}
	for _, a := range s.cfg.AllowAccounts {
		if a == user {
			return false
		}
	}
	return true
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
		s.mu.Lock()
		s.draining = true
		s.mu.Unlock()
		_ = ln.Close()
		grace := s.cfg.ShutdownGrace
		if grace <= 0 {
			grace = 30 * time.Second
		}
		done := make(chan struct{})
		go func() {
			s.sessionsWG.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(grace):
			s.liveSSH.Range(func(key, value any) bool {
				if sshSess, ok := value.(ssh.Session); ok {
					_ = sshSess.Close()
				}
				return true
			})
			s.sessionsWG.Wait()
		}
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

// ReloadInstallation validates and swaps the installation.
func (s *Server) ReloadInstallation(b *installation.Builder) error {
	exp, err := b.Build()
	if err != nil {
		return err
	}
	if err := s.engine.SetInstallation(exp); err != nil {
		s.engine.NoteReloadFailure()
		return err
	}
	s.engine.RedrawSessions()
	return nil
}

func ioWriteString(w ioStringWriter, v string) (int, error) {
	return w.Write([]byte(v))
}

type ioStringWriter interface {
	Write([]byte) (int, error)
}
