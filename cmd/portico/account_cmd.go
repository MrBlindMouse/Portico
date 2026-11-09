package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/MrBlindMouse/Portico/identity"
)

func accountExportCmd(args []string) {
	fs := flag.NewFlagSet("account-export", flag.ExitOnError)
	data := fs.String("data", ".portico-data", "data directory")
	_ = fs.Parse(args)
	store, err := identity.OpenFileAccountStore(*data)
	if err != nil {
		log.Fatal(err)
	}
	out, err := store.Export()
	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

func accountImportCmd(args []string) {
	fs := flag.NewFlagSet("account-import", flag.ExitOnError)
	data := fs.String("data", ".portico-data", "data directory")
	file := fs.String("file", "", "JSON array of AccountRecord")
	_ = fs.Parse(args)
	if *file == "" {
		log.Fatal("account-import requires -file")
	}
	b, err := os.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
	}
	var records []identity.AccountRecord
	if err := json.Unmarshal(b, &records); err != nil {
		log.Fatal(err)
	}
	store, err := identity.OpenFileAccountStore(*data)
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Import(records); err != nil {
		log.Fatal(err)
	}
}
