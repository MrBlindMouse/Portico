package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	portico "github.com/MrBlindMouse/Portico"
	"github.com/MrBlindMouse/Portico/components"
	"github.com/MrBlindMouse/Portico/config"
	"github.com/MrBlindMouse/Portico/identity"
	"github.com/MrBlindMouse/Portico/internal/plans"
	"github.com/MrBlindMouse/Portico/runtime"
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
	case "generate":
		generateCmd(os.Args[2:])
	case "identity-export":
		identityExportCmd(os.Args[2:])
	case "account-export":
		accountExportCmd(os.Args[2:])
	case "account-import":
		accountImportCmd(os.Args[2:])
	case "wire":
		wireCmd(os.Args[2:])
	case "init":
		initCmd(os.Args[2:])
	case "docs":
		docsCmd(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: portico <init|generate|docs|inspect|wire|identity-export|account-export|account-import|serve -demo> [flags]\n")
	fmt.Fprintf(os.Stderr, "  A site is your own Go module that imports this library.\n")
	fmt.Fprintf(os.Stderr, "  Create one: go run github.com/MrBlindMouse/Portico/cmd/portico@<tag> init -module <path>\n")
	fmt.Fprintf(os.Stderr, "  Then in the site: go run github.com/MrBlindMouse/Portico/cmd/portico generate <plan>\n")
	fmt.Fprintf(os.Stderr, "  Docs: go run …/cmd/portico docs -tag <version>   (portico serve is only -demo)\n")
}

func docsCmd(args []string) {
	fs := flag.NewFlagSet("docs", flag.ExitOnError)
	tag := fs.String("tag", "main", "git tag for doc URLs (match your go.mod version)")
	_ = fs.Parse(args)
	base := "https://github.com/MrBlindMouse/Portico/blob/" + *tag
	fmt.Printf("manual:          %s/docs/manual/README.md\n", base)
	fmt.Printf("building-a-site: %s/docs/building-a-site.md\n", base)
	fmt.Printf("site-agents:     %s/docs/site-agents.md\n", base)
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
		log.Fatal("portico serve is a smoke test and requires -demo. Your site is a separate module: portico init -module <path>, then go run .")
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
		Merge(plans.Guestbook()).
		Merge(plans.Chat()).
		Merge(plans.Landing(map[string]components.SceneID{
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
		store := runtime.NewAccountStore()
		opts.AccountStore = store
		opts.Accounts = &runtime.AccountConfig{Enabled: true, Accounts: store}
	}
	if err := portico.Serve(*addr, exp, opts); err != nil {
		log.Fatal(err)
	}
}

func inspectCmd(args []string) {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	base := fs.String("url", "http://127.0.0.1:9922", "inspect base URL")
	resource := fs.String("resource", "snapshot", "snapshot|sessions|reload-preview|validate")
	method := fs.String("method", "", "HTTP method (default GET; POST for reload-preview|validate)")
	file := fs.String("file", "", "JSON body file for POST")
	_ = fs.Parse(args)

	httpMethod := *method
	if httpMethod == "" {
		switch *resource {
		case "reload-preview", "validate":
			httpMethod = http.MethodPost
		default:
			httpMethod = http.MethodGet
		}
	}
	url := *base + "/inspect/" + *resource
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()
	var body io.Reader
	if *file != "" {
		b, err := os.ReadFile(*file)
		if err != nil {
			log.Fatal(err)
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, httpMethod, url, body)
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
