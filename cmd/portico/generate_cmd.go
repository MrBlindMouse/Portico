package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

//go:generate cp ../../internal/plans/landing.go planembed/landing.go.txt
//go:generate cp ../../internal/plans/guestbook.go planembed/guestbook.go.txt
//go:generate cp ../../internal/plans/chat.go planembed/chat.go.txt
//go:generate cp ../../internal/plans/register.go planembed/register.go.txt
//go:generate cp ../../internal/plans/form.go planembed/form.go.txt
//go:generate cp ../../internal/plans/blog.go planembed/blog.go.txt
//go:generate cp ../../internal/plans/lobby.go planembed/lobby.go.txt
//go:generate cp ../../internal/plans/rest.go planembed/rest.go.txt

//go:embed planembed/*
var planSources embed.FS

type planPiece struct {
	embed string // file name under planembed/, without directory
	dest  string
	label string // header comment name
}

type planDef struct {
	name  string
	files []planPiece
	call  string
}

func planCatalog() []planDef {
	rest := planPiece{embed: "rest.go.txt", dest: "rest.go", label: "rest"}
	return []planDef{
		{name: "landing", files: []planPiece{{embed: "landing.go.txt", dest: "landing.go", label: "landing"}}, call: landingCall},
		{name: "guestbook", files: []planPiece{{embed: "guestbook.go.txt", dest: "guestbook.go", label: "guestbook"}, rest}, call: "portico.New().Merge(Guestbook()).Build()"},
		{name: "chat", files: []planPiece{{embed: "chat.go.txt", dest: "chat.go", label: "chat"}, rest}, call: "portico.New().Merge(Chat()).Build()"},
		{name: "register", files: []planPiece{{embed: "register.go.txt", dest: "register.go", label: "register"}}, call: "portico.New().Merge(Registration()).Build()"},
		{name: "form", files: []planPiece{{embed: "form.go.txt", dest: "form.go", label: "form"}}, call: "portico.New().Merge(Form()).Build()"},
		{name: "blog", files: []planPiece{{embed: "blog.go.txt", dest: "blog.go", label: "blog"}}, call: "portico.New().Merge(Blog()).Build()"},
		{name: "lobby", files: []planPiece{{embed: "lobby.go.txt", dest: "lobby.go", label: "lobby"}}, call: "portico.New().Merge(Lobby()).Build()"},
	}
}

const landingCall = `portico.New().Merge(Landing(map[string]components.SceneID{
    // "guestbook": "guestbook",
})).Build()`

func planNames() []string {
	catalog := planCatalog()
	names := make([]string, len(catalog))
	for i := range catalog {
		names[i] = catalog[i].name
	}
	return names
}

func generateCmd(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	dir := fs.String("dir", ".", "directory to write the generated file")
	pkg := fs.String("pkg", "", "package clause (default: package in -dir, or main)")
	force := fs.Bool("force", false, "overwrite existing files")
	list := fs.Bool("list", false, "list plan names")
	_ = fs.Parse(args)
	if *list {
		for _, p := range planCatalog() {
			fmt.Println(p.name)
		}
		return
	}
	if fs.NArg() != 1 {
		log.Fatal("usage: portico generate [-dir .] [-pkg name] [-force] <name>\n       portico generate -list")
	}
	call, err := generatePlan(*dir, *pkg, fs.Arg(0), *force)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(call)
}

func generatePlan(dir, pkg, name string, force bool) (string, error) {
	catalog := planCatalog()
	var def *planDef
	for i := range catalog {
		if catalog[i].name == name {
			def = &catalog[i]
			break
		}
	}
	if def == nil {
		return "", fmt.Errorf("unknown plan %q (portico generate -list)", name)
	}
	target := filepath.Clean(dir)
	if err := os.MkdirAll(target, 0o755); err != nil {
		return "", err
	}
	if pkg == "" {
		detected, err := detectPackage(target)
		if err != nil {
			return "", err
		}
		pkg = detected
	}
	if !validPackageName(pkg) {
		return "", fmt.Errorf("invalid package name %q", pkg)
	}
	for _, piece := range def.files {
		src, err := planSources.ReadFile("planembed/" + piece.embed)
		if err != nil {
			return "", err
		}
		body, err := renderPlan(src, pkg, piece.label)
		if err != nil {
			return "", err
		}
		dest := filepath.Join(target, piece.dest)
		if _, err := os.Stat(dest); err == nil && !force {
			return "", fmt.Errorf("%s exists (use -force)", dest)
		} else if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		if err := os.WriteFile(dest, body, 0o644); err != nil {
			return "", err
		}
	}
	return def.call, nil
}

func renderPlan(src []byte, pkg, label string) ([]byte, error) {
	body := string(src)
	const old = "package plans\n"
	if !strings.HasPrefix(body, old) {
		return nil, fmt.Errorf("plan %s: missing package plans", label)
	}
	body = "package " + pkg + "\n" + strings.TrimPrefix(body, old)
	header := fmt.Sprintf("// Copied from Portico plan %s. Edit freely.\n", label)
	return []byte(header + body), nil
}

func detectPackage(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			return "", err
		}
		pkg, err := packageClause(f)
		f.Close()
		if err != nil {
			return "", err
		}
		if pkg != "" {
			return pkg, nil
		}
	}
	return "main", nil
}

func packageClause(r fs.File) (string, error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if rest, ok := strings.CutPrefix(line, "package "); ok {
			return rest, nil
		}
		return "", nil
	}
	return "", sc.Err()
}

func validPackageName(pkg string) bool {
	if pkg == "" || pkg == "_" {
		return false
	}
	for i, r := range pkg {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}
