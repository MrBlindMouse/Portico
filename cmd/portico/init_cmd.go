package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed templates/*
var siteTemplates embed.FS

func initCmd(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	dir := fs.String("dir", ".", "directory to write site files")
	module := fs.String("module", "", "go module path (required if go.mod missing)")
	force := fs.Bool("force", false, "overwrite existing files")
	_ = fs.Parse(args)

	target := filepath.Clean(*dir)
	if err := os.MkdirAll(target, 0o755); err != nil {
		log.Fatal(err)
	}

	goMod := filepath.Join(target, "go.mod")
	if _, err := os.Stat(goMod); os.IsNotExist(err) {
		if *module == "" {
			log.Fatal("go.mod missing: pass -module github.com/you/my-site")
		}
		cmd := exec.Command("go", "mod", "init", *module)
		cmd.Dir = target
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}

	get := exec.Command("go", "get", "github.com/MrBlindMouse/Portico@latest")
	get.Dir = target
	get.Stdout = os.Stdout
	get.Stderr = os.Stderr
	if err := get.Run(); err != nil {
		log.Fatal(err)
	}

	files := map[string]string{
		"main.go":   "main.tmpl",
		"AGENTS.md": "AGENTS.md",
		"README.md": "README.site.md",
	}
	for out, tmpl := range files {
		if err := writeTemplate(target, out, tmpl, *force); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Fprintf(os.Stdout, "wrote %s. Run: cd %s && go run .\n", target, target)
	fmt.Fprintf(os.Stdout, "Next: go run github.com/MrBlindMouse/Portico/cmd/portico generate <plan>   (names: %s)\n", strings.Join(planNames(), ", "))
}

func writeTemplate(dir, outName, tmplName string, force bool) error {
	dest := filepath.Join(dir, outName)
	if _, err := os.Stat(dest); err == nil && !force {
		return fmt.Errorf("%s exists (use -force)", dest)
	}
	data, err := siteTemplates.ReadFile("templates/" + tmplName)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0o644)
}
