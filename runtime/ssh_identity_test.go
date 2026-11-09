package runtime_test

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/runtime"
	"github.com/MrBlindMouse/Portico/testkit"
	gossh "golang.org/x/crypto/ssh"
)

func TestSSHNoKeyUsesLogin(t *testing.T) {
	addr := serveIdentity(t, identityScene(t, false), nil)
	client := dialSSH(t, addr, &gossh.ClientConfig{
		User:            "alice",
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	})
	defer client.Close()
	out := readShell(t, client)
	if !strings.Contains(out, "anon alice") {
		t.Fatalf("output=%q", out)
	}
	if strings.Contains(out, "account") {
		t.Fatalf("guest became an account: %q", out)
	}
}

func TestSSHKeyIdentityPersists(t *testing.T) {
	signer := testSigner(t)
	fp := identity.FromPublicKey(signer.PublicKey()).Fingerprint()
	addr := serveIdentity(t, identityScene(t, true), func(cfg *runtime.ServerConfig) {
		cfg.AllowKeys = []string{fp}
	})
	client := dialSSH(t, addr, &gossh.ClientConfig{
		User:            "alice",
		Auth:            []gossh.AuthMethod{gossh.PublicKeys(signer)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	})
	defer client.Close()
	out := readShell(t, client)
	if !strings.Contains(out, "key alice") || !strings.Contains(out, "persisted") {
		t.Fatalf("output=%q", out)
	}
	if strings.Contains(out, "account") {
		t.Fatalf("key session became an account: %q", out)
	}
}

func TestSSHDenyKey(t *testing.T) {
	signer := testSigner(t)
	fp := identity.FromPublicKey(signer.PublicKey()).Fingerprint()
	addr := serveIdentity(t, identityScene(t, false), func(cfg *runtime.ServerConfig) {
		cfg.DenyKeys = []string{fp}
	})
	_, err := gossh.Dial("tcp", addr, &gossh.ClientConfig{
		User:            "alice",
		Auth:            []gossh.AuthMethod{gossh.PublicKeys(signer)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         2 * time.Second,
	})
	if err == nil {
		t.Fatal("denied key connected")
	}
}

func TestSSHAccountsKeepKeySeparateFromPassword(t *testing.T) {
	signer := testSigner(t)
	addr := serveIdentity(t, identityScene(t, false), func(cfg *runtime.ServerConfig) {
		cfg.Accounts = &runtime.AccountConfig{
			Enabled: true,
			Users:   map[string]string{"alice": "secret"},
		}
	})

	keyClient := dialSSH(t, addr, &gossh.ClientConfig{
		User:            "alice",
		Auth:            []gossh.AuthMethod{gossh.PublicKeys(signer)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	})
	keyOut := readShell(t, keyClient)
	keyClient.Close()
	if !strings.Contains(keyOut, "key alice") || strings.Contains(keyOut, "account") {
		t.Fatalf("key with accounts=%q", keyOut)
	}

	passClient := dialSSH(t, addr, &gossh.ClientConfig{
		User:            "alice",
		Auth:            []gossh.AuthMethod{gossh.Password("secret")},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	})
	defer passClient.Close()
	passOut := readShell(t, passClient)
	if !strings.Contains(passOut, "account alice") {
		t.Fatalf("password output=%q", passOut)
	}

	_, err := gossh.Dial("tcp", addr, &gossh.ClientConfig{
		User:            "alice",
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         2 * time.Second,
	})
	if err == nil {
		t.Fatal("accounts enabled accepted none-auth")
	}
}

func TestSSHRequireAccountRejectsNone(t *testing.T) {
	addr := serveIdentity(t, identityScene(t, false), func(cfg *runtime.ServerConfig) {
		cfg.RequireAccount = true
		cfg.Accounts = &runtime.AccountConfig{
			Enabled: true,
			Users:   map[string]string{"alice": "secret"},
		}
	})
	_, err := gossh.Dial("tcp", addr, &gossh.ClientConfig{
		User:            "alice",
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         2 * time.Second,
	})
	if err == nil {
		t.Fatal("require account accepted none-auth")
	}
	client := dialSSH(t, addr, &gossh.ClientConfig{
		User:            "alice",
		Auth:            []gossh.AuthMethod{gossh.Password("secret")},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	})
	defer client.Close()
	out := readShell(t, client)
	if !strings.Contains(out, "account alice") {
		t.Fatalf("output=%q", out)
	}
}

func TestSSHLinkedKeyBecomesAccount(t *testing.T) {
	signer := testSigner(t)
	fp := identity.FromPublicKey(signer.PublicKey()).Fingerprint()
	accounts, err := identity.OpenFileAccountStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := accounts.Register("alice", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := accounts.LinkKey("alice", fp); err != nil {
		t.Fatal(err)
	}
	addr := serveIdentity(t, identityScene(t, false), func(cfg *runtime.ServerConfig) {
		cfg.Accounts = &runtime.AccountConfig{Enabled: true, Accounts: accounts}
	})
	client := dialSSH(t, addr, &gossh.ClientConfig{
		User:            "someone",
		Auth:            []gossh.AuthMethod{gossh.PublicKeys(signer)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	})
	defer client.Close()
	out := readShell(t, client)
	if !strings.Contains(out, "account alice") || strings.Contains(out, "someone") {
		t.Fatalf("linked key output=%q", out)
	}
}

func identityScene(t *testing.T, persist bool) *experience.Experience {
	t.Helper()
	b := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(ctx experience.RenderContext) string {
				id := ctx.Identity()
				s := id.String()
				switch {
				case id.Account() != "":
					s = "account " + id.Account()
				case id.Anonymous():
					s = "anon " + s
				default:
					s = "key " + s
				}
				if v, ok := ctx.Persisted("token"); ok && v == "yes" {
					s += " persisted"
				}
				return s + "\n"
			},
		})
	if persist {
		b.OnSessionStart(experience.PersistEffect("token", "yes"))
	}
	exp, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	return exp
}

func serveIdentity(t *testing.T, exp *experience.Experience, mutate func(*runtime.ServerConfig)) string {
	t.Helper()
	eng := testkit.New(exp).Engine
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	cfg := runtime.DefaultConfig()
	cfg.Addr = ln.Addr().String()
	if mutate != nil {
		mutate(&cfg)
	}
	srv, err := runtime.NewServer(cfg, eng)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		_ = srv.ListenAndServeOn(ctx, ln)
	}()
	time.Sleep(50 * time.Millisecond)
	return ln.Addr().String()
}

func testSigner(t *testing.T) gossh.Signer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := gossh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

func dialSSH(t *testing.T, addr string, cfg *gossh.ClientConfig) *gossh.Client {
	t.Helper()
	if cfg.Timeout == 0 {
		cfg.Timeout = 2 * time.Second
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		client, err := gossh.Dial("tcp", addr, cfg)
		if err == nil {
			return client
		}
		if time.Now().After(deadline) || strings.Contains(err.Error(), "authenticate") {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func readShell(t *testing.T, client *gossh.Client) string {
	t.Helper()
	sess, err := client.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	out, err := sess.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := sess.Shell(); err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(out)
	var b strings.Builder
	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("timeout output=%q", b.String())
		default:
			line, err := reader.ReadString('\n')
			b.WriteString(line)
			if b.Len() > 0 && (strings.Contains(b.String(), "guest") || strings.Contains(b.String(), "key-") || strings.Contains(b.String(), "alice") || strings.Contains(b.String(), "account required")) {
				return b.String()
			}
			if err != nil {
				t.Fatalf("read output=%q err=%v", b.String(), err)
			}
		}
	}
}
