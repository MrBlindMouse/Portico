package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	portico "github.com/MrBlindMouse/Portico"
	"github.com/MrBlindMouse/Portico/config"
	"github.com/MrBlindMouse/Portico/experience"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/runtime"
	"github.com/MrBlindMouse/Portico/scaffold"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "serve":
		serveCmd(os.Args[2:])
	case "inspect":
		inspectCmd(os.Args[2:])
	case "scaffold":
		scaffoldCmd(os.Args[2:])
	case "identity-export":
		identityExportCmd(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: portico <inspect|scaffold|identity-export|serve -demo> [flags]\n")
	fmt.Fprintf(os.Stderr, "  Sites are Go programs — see examples/hello and README.md (not portico serve).\n")
}

func serveCmd(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	demo := fs.Bool("demo", false, "run bundled guestbook/chat demo (for smoke tests only)")
	addr := fs.String("addr", ":2222", "SSH listen address")
	data := fs.String("data", ".portico-data", "data directory")
	inspectAddr := fs.String("inspect", "127.0.0.1:9922", "inspect HTTP address")
	statusAddr := fs.String("status", "127.0.0.1:9923", "status HTTP address")
	configPath := fs.String("config", "", "optional JSON config file")
	enableAccounts := fs.Bool("accounts", false, "enable password auth (dev plaintext store)")
	logLevel := fs.String("log-level", "info", "log level: debug, info, warn, error")
	_ = fs.Parse(args)
	if !*demo {
		log.Fatal("portico serve requires -demo (operator smoke test). Build your site: go run ./examples/hello")
	}
	var fileCfg config.File
	if *configPath != "" {
		cfg, err := loadConfig(*configPath)
		if err != nil {
			log.Fatal(err)
		}
		if cfg.Listen != "" {
			*addr = cfg.Listen
		}
		if cfg.DataDir != "" {
			*data = cfg.DataDir
		}
		if cfg.InspectAddr != "" {
			*inspectAddr = cfg.InspectAddr
		}
		if cfg.StatusAddr != "" {
			*statusAddr = cfg.StatusAddr
		}
		if cfg.Accounts {
			*enableAccounts = true
		}
		if cfg.LogLevel != "" {
			*logLevel = cfg.LogLevel
		}
		fileCfg = cfg
	}

	exp, err := portico.New().
		Merge(scaffold.Guestbook()).
		Merge(scaffold.Chat()).
		Merge(scaffold.Landing(map[string]experience.SceneID{
			"guestbook": "guestbook",
			"chat":      "chat",
		})).
		Build()
	if err != nil {
		log.Fatal(err)
	}
	opts := portico.Options{
		DataDir:     *data,
		InspectAddr: *inspectAddr,
		StatusAddr:  *statusAddr,
		LogLevel:    *logLevel,
	}
	opts = opts.MergeConfig(fileCfg)
	if *enableAccounts {
		opts.AccountStore = runtime.NewAccountStore()
		opts.Accounts = &runtime.AccountConfig{Enabled: true, Accounts: opts.AccountStore}
	}
	if err := portico.Serve(*addr, exp, opts); err != nil {
		log.Fatal(err)
	}
}

func inspectCmd(args []string) {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	base := fs.String("url", "http://127.0.0.1:9922", "inspect base URL")
	resource := fs.String("resource", "snapshot", "snapshot|sessions")
	_ = fs.Parse(args)

	url := *base + "/inspect/" + *resource
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var v any
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func loadConfig(path string) (config.File, error) {
	return config.Load(path)
}

func identityExportCmd(args []string) {
	fs := flag.NewFlagSet("identity-export", flag.ExitOnError)
	data := fs.String("data", ".portico-data", "data directory")
	_ = fs.Parse(args)
	store, err := identity.OpenFileStore(*data)
	if err != nil {
		log.Fatal(err)
	}
	out, err := store.ExportAll()
	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

func scaffoldCmd(args []string) {
	if len(args) == 0 {
		log.Fatal("usage: portico scaffold <landing|guestbook|chat|register|nickname>")
	}
	switch args[0] {
	case "landing":
		fmt.Println(`portico.New().Merge(scaffold.Landing(map[string]experience.SceneID{"guestbook": "guestbook"}))`)
	case "guestbook":
		fmt.Println("portico.New().Merge(scaffold.Guestbook())")
	case "chat":
		fmt.Println("portico.New().Merge(scaffold.Chat())")
	case "register", "registration":
		fmt.Println("portico.New().Merge(scaffold.Registration()) // enable opts.Accounts")
	case "nickname":
		fmt.Println(`portico.New().Merge(scaffold.NicknameCommands()) // add Inputs["nick"] to scenes`)
	default:
		log.Fatalf("unknown scaffold %q", args[0])
	}
}
