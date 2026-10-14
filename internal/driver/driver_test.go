package driver

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoldenRuns(t *testing.T) {
	files, err := filepath.Glob("testdata/*.scar")
	if err != nil || len(files) == 0 {
		t.Fatalf("no testdata found: %v", err)
	}
	for _, src := range files {
		src := src
		t.Run(filepath.Base(src), func(t *testing.T) {
			wantPath := strings.TrimSuffix(src, ".scar") + ".stdout"
			want, err := os.ReadFile(wantPath)
			if err != nil {
				t.Fatalf("missing golden file: %v", err)
			}
			var stdout, stderr bytes.Buffer
			if err := Run(src, &stdout, &stderr); err != nil {
				t.Fatalf("run failed: %v\nstderr:\n%s", err, stderr.String())
			}
			if stdout.String() != string(want) {
				t.Errorf("output mismatch\n--- got ---\n%s\n--- want ---\n%s", stdout.String(), want)
			}
		})
	}
}

func TestCompileErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		msg  string
		line int
		col  int
	}{
		{"missing type inference", "val x = []\n", "cannot infer type", 1, 9},
		{"type mismatch", "val x: int = \"a\"\n", "cannot use value of type", 1, 14},
		{"immutable assign", "val x: int = 1\nx = 2\n", "immutable variable", 2, 1},
		{"undefined var", "print(y)\n", "undefined variable", 1, 7},
		{"bad condition", "if 42 {\nprint(1)\n}\n", "condition must be bool", 1, 4},
		{"syntax error", "val x: int = 1\nval y: int = \n", "expected expression", 2, 14},
		{"str concat type", "val s: str = \"a\" + 1\n", "'+' is not defined", 1, 18},
		{"reserved word", "val type: int = 1\n", "reserved word", 1, 5},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := CompileSource(c.src, "t.scar")
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			s := err.Error()
			if !strings.Contains(s, c.msg) {
				t.Errorf("error %q, want contains %q", s, c.msg)
			}
			wantPos := fmt.Sprintf("t.scar:%d:%d", c.line, c.col)
			if !strings.Contains(s, wantPos) {
				t.Errorf("error %q, want position %q", s, wantPos)
			}
		})
	}
}

func TestCompileSuccess(t *testing.T) {
	out, err := CompileSource("val x: int = 1 + 2\nprint(x)\n", "t.scar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "package main") {
		t.Errorf("bad output: %s", out)
	}
}

func TestDefaultOutputPath(t *testing.T) {
	got := DefaultOutputPath("examples/demo.scar")
	base := filepath.Base(got)
	if !strings.HasPrefix(base, "demo") {
		t.Errorf("DefaultOutputPath = %q", got)
	}
}
