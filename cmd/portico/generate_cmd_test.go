package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanEmbedMatchesSource(t *testing.T) {
	for _, name := range []string{"landing", "guestbook", "chat", "register", "form", "blog", "lobby", "rest"} {
		embedded, err := planSources.ReadFile("planembed/" + name + ".go.txt")
		if err != nil {
			t.Fatal(err)
		}
		src, err := os.ReadFile(filepath.Join("..", "..", "internal", "plans", name+".go"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(embedded, src) {
			t.Fatalf("%s embed is stale; run go generate ./cmd/portico", name)
		}
	}
}

func TestGenerateGuestbook(t *testing.T) {
	dir := t.TempDir()
	call, err := generatePlan(dir, "", "guestbook", false)
	if err != nil {
		t.Fatal(err)
	}
	if call != "portico.New().Merge(Guestbook()).Build()" {
		t.Fatalf("call %q", call)
	}
	gb, err := os.ReadFile(filepath.Join(dir, "guestbook.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(gb)
	if !strings.HasPrefix(text, "// Copied from Portico plan guestbook. Edit freely.\npackage main\n") {
		t.Fatalf("header/package: %s", text[:80])
	}
	if !strings.Contains(text, "func Guestbook(") || !strings.Contains(text, "installation.NewBuilder") {
		t.Fatal("guestbook body missing")
	}
	if _, err := os.Stat(filepath.Join(dir, "rest.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := generatePlan(dir, "", "guestbook", false); err == nil {
		t.Fatal("expected error when file exists")
	}
}

func TestGenerateChatSharesRestAfter(t *testing.T) {
	dir := t.TempDir()
	if _, err := generatePlan(dir, "main", "guestbook", false); err != nil {
		t.Fatal(err)
	}
	if _, err := generatePlan(dir, "main", "chat", true); err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			ast.Inspect(f, func(node ast.Node) bool {
				fn, ok := node.(*ast.FuncDecl)
				if ok && fn.Name.Name == "restAfter" {
					n++
				}
				return true
			})
		}
	}
	if n != 1 {
		t.Fatalf("restAfter declarations: %d", n)
	}
}

func TestGenerateDetectsPackage(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "doc.go"), []byte("package site\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := generatePlan(dir, "", "lobby", false); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "lobby.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "package site\n") {
		t.Fatalf("package: %s", b[:60])
	}
}

func TestGenerateGuestbookBuilds(t *testing.T) {
	dir := t.TempDir()
	if _, err := generatePlan(dir, "main", "guestbook", false); err != nil {
		t.Fatal(err)
	}
	mainSrc := `package main

import (
	"os"

	portico "github.com/MrBlindMouse/Portico"
)

func main() {
	exp, err := portico.New().Merge(Guestbook()).Build()
	if err != nil || exp.Root() == "" {
		os.Exit(1)
	}
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	mod := "module example.com/generated\n\ngo 1.23\n\nrequire github.com/MrBlindMouse/Portico v0.0.0\n\nreplace github.com/MrBlindMouse/Portico => " + root + "\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
		t.Fatal(err)
	}
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = dir
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, "site"), ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
}
