package driver

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ScarfazeCompany/scar-script/internal/checker"
	"github.com/ScarfazeCompany/scar-script/internal/codegen"
	"github.com/ScarfazeCompany/scar-script/internal/lexer"
	"github.com/ScarfazeCompany/scar-script/internal/parser"
)

// Compile reads a .scar file and returns generated, gofmt-formatted Go source.
func Compile(srcPath string) ([]byte, error) {
	src, err := os.ReadFile(srcPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", srcPath, err)
	}
	return CompileSource(string(src), srcPath)
}

// CompileSource compiles ScarScript source text to Go source.
func CompileSource(src, filename string) ([]byte, error) {
	toks, lerr := lexer.Lex(src, filename)
	if lerr != nil {
		return nil, lerr
	}
	prog, perr := parser.Parse(toks, filename)
	if perr != nil {
		return nil, perr
	}
	res, cerr := checker.Check(prog)
	if cerr != nil {
		return nil, cerr
	}
	out, gerr := codegen.Generate(prog, res)
	if gerr != nil {
		return nil, gerr
	}
	return []byte(out), nil
}

// Run compiles a .scar file to a temporary directory and executes it with
// `go run`, sending program output to the given writers.
func Run(srcPath string, stdout, stderr io.Writer) error {
	gen, err := Compile(srcPath)
	if err != nil {
		return err
	}
	dir, cleanup, err := writeTemp(gen)
	if err != nil {
		return err
	}
	defer cleanup()

	cmd := exec.Command("go", "run", filepath.Join(dir, "main.go"))
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("program exited with code %d", ee.ExitCode())
		}
		return fmt.Errorf("go run failed: %w", err)
	}
	return nil
}

// Build compiles a .scar file into a native binary at outPath.
func Build(srcPath, outPath string) error {
	gen, err := Compile(srcPath)
	if err != nil {
		return err
	}
	absOut, err := filepath.Abs(outPath)
	if err != nil {
		return fmt.Errorf("invalid output path %q: %w", outPath, err)
	}
	dir, cleanup, err := writeTemp(gen)
	if err != nil {
		return err
	}
	defer cleanup()

	cmd := exec.Command("go", "build", "-o", absOut, filepath.Join(dir, "main.go"))
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("go build failed with exit code %d", ee.ExitCode())
		}
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}

// DefaultOutputPath derives a binary name from the source file name.
func DefaultOutputPath(srcPath string) string {
	base := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	if strings.HasSuffix(strings.ToLower(os.Getenv("OS")), "windows") {
		return base + ".exe"
	}
	return base
}

func writeTemp(gen []byte) (dir string, cleanup func(), err error) {
	dir, err = os.MkdirTemp("", "scar-*")
	if err != nil {
		return "", nil, fmt.Errorf("cannot create temp dir: %w", err)
	}
	cleanup = func() { os.RemoveAll(dir) }
	if err := os.WriteFile(filepath.Join(dir, "main.go"), gen, 0o644); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("cannot write generated code: %w", err)
	}
	return dir, cleanup, nil
}
