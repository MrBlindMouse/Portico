package fakessh_test

import (
	"bufio"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/runtime"
	gossh "golang.org/x/crypto/ssh"
)

func TestEmbeddedSSH(t *testing.T) {
	exp, err := experience.NewBuilder().
		Root("home").
		Scene("home", experience.Scene{
			Render: func(experience.RenderContext) string { return "hello-ssh" },
		}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	eng := runtime.NewEngine(exp, runtime.NewApplier(nil, identity.NewMemoryStore()))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	cfg := runtime.DefaultConfig()
	cfg.Addr = ln.Addr().String()
	srv, err := runtime.NewServer(cfg, eng)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = srv.ListenAndServeOn(ctx, ln)
	}()
	time.Sleep(50 * time.Millisecond)

	client, err := gossh.Dial("tcp", ln.Addr().String(), &gossh.ClientConfig{
		User:            "test",
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
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
readLoop:
	for {
		select {
		case <-deadline:
			break readLoop
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				break readLoop
			}
			b.WriteString(line)
			if strings.Contains(b.String(), "hello-ssh") {
				break readLoop
			}
		}
	}
	if !strings.Contains(b.String(), "hello-ssh") {
		t.Fatalf("output=%q", b.String())
	}
}
