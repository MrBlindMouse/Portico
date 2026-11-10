package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/MrBlindMouse/Portico/installation"
	"github.com/MrBlindMouse/Portico/internal/codegen"
)

func wireCmd(args []string) {
	if len(args) == 0 {
		log.Fatal("usage: portico wire <validate|codegen> -file site.wire.json")
	}
	switch args[0] {
	case "validate":
		wireValidateCmd(args[1:])
	case "codegen":
		wireCodegenCmd(args[1:])
	default:
		log.Fatalf("unknown wire subcommand %q", args[0])
	}
}

func wireValidateCmd(args []string) {
	fs := flag.NewFlagSet("wire validate", flag.ExitOnError)
	file := fs.String("file", "", "wire JSON file")
	_ = fs.Parse(args)
	if *file == "" {
		log.Fatal("-file required")
	}
	w, err := loadWire(*file)
	if err != nil {
		log.Fatal(err)
	}
	if err := installation.ValidateWire(w); err != nil {
		log.Fatal(err)
	}
	fmt.Println("ok")
}

func wireCodegenCmd(args []string) {
	fs := flag.NewFlagSet("wire codegen", flag.ExitOnError)
	file := fs.String("file", "", "wire JSON file")
	out := fs.String("out", "-", "output Go file (- for stdout)")
	_ = fs.Parse(args)
	if *file == "" {
		log.Fatal("-file required")
	}
	w, err := loadWire(*file)
	if err != nil {
		log.Fatal(err)
	}
	code := codegen.FromWire(w, "main")
	if *out == "-" {
		fmt.Print(code)
		return
	}
	if err := os.WriteFile(*out, []byte(code), 0o644); err != nil {
		log.Fatal(err)
	}
}

func loadWire(path string) (installation.Wire, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return installation.Wire{}, err
	}
	var w installation.Wire
	err = json.Unmarshal(b, &w)
	return w, err
}
