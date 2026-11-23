package driver

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ScarfazeCompany/scar-script/internal/ast"
	"github.com/ScarfazeCompany/scar-script/internal/checker"
	"github.com/ScarfazeCompany/scar-script/internal/codegen"
	"github.com/ScarfazeCompany/scar-script/internal/lexer"
	"github.com/ScarfazeCompany/scar-script/internal/parser"
)

// Compile reads a .scar file and returns generated, gofmt-formatted Go source.
func Compile(srcPath string) ([]byte, error) {
	absPath, err := filepath.Abs(srcPath)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve path %q: %w", srcPath, err)
	}

	// Resolve all modules and detect cycles
	modules, err := resolveModules(absPath, make(map[string]*moduleInfo), nil)
	if err != nil {
		return nil, err
	}

	// Bundle all modules into a single program
	prog, err := bundleModules(modules, absPath)
	if err != nil {
		return nil, err
	}

	return CompileSourceWithProgram(prog)
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
	return CompileSourceWithProgram(prog)
}

// CompileSourceWithProgram compiles a pre-parsed AST program to Go source.
func CompileSourceWithProgram(prog *ast.Program) ([]byte, error) {
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

// moduleInfo holds a parsed module and its exported symbols.
type moduleInfo struct {
	path     string
	prog     *ast.Program
	exported map[string]bool // set of exported symbol names
	deps     []string        // module paths this module imports
}

// resolveModules recursively reads and parses all modules, detecting cycles.
func resolveModules(path string, visited map[string]*moduleInfo, stack []string) (map[string]*moduleInfo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve path %q: %w", path, err)
	}

	// Cycle detection
	for _, s := range stack {
		if s == absPath {
			var cycle []string
			cycle = append(cycle, stack...)
			cycle = append(cycle, absPath)
			return nil, fmt.Errorf("[Import Error] Circular dependency detected: %s", strings.Join(cycle, " -> "))
		}
	}

	// Already visited
	if _, ok := visited[absPath]; ok {
		return visited, nil
	}

	src, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read module %q: %w", absPath, err)
	}

	toks, lerr := lexer.Lex(string(src), absPath)
	if lerr != nil {
		return nil, lerr
	}
	prog, perr := parser.Parse(toks, absPath)
	if perr != nil {
		return nil, perr
	}

	mi := &moduleInfo{
		path:     absPath,
		prog:     prog,
		exported: make(map[string]bool),
	}

	// Collect exports and imports
	newStack := append(stack, absPath)
	for _, s := range prog.Stmts {
		if ed, ok := s.(*ast.ExportDecl); ok {
			mi.collectExports(ed.Decl)
		} else if d, ok := s.(*ast.DeclStmt); ok {
			mi.exported[d.Name] = false // not exported
		} else if fd, ok := s.(*ast.FuncDecl); ok {
			mi.exported[fd.Name] = false
		} else if sd, ok := s.(*ast.StructDecl); ok {
			mi.exported[sd.Name] = false
		} else if imp, ok := s.(*ast.ImportStmt); ok {
			mi.deps = append(mi.deps, imp.From)
			// Verify imported symbols are exported
			impPath := resolveImportPath(absPath, imp.From)
			impVisited, err := resolveModules(impPath, visited, newStack)
			if err != nil {
				return nil, err
			}
			impMod, impOk := impVisited[impPath]
			if !impOk {
				return nil, fmt.Errorf("[Import Error] cannot import module %q", imp.From)
			}
			for _, name := range imp.Names {
				if !impMod.exported[name] {
					return nil, fmt.Errorf("[Import Error] line %d:col %d: Symbol %q is not exported by %q",
						imp.FromPos.Line, imp.FromPos.Col, name, imp.From)
				}
				mi.exported[name] = true // mark as available for re-export
			}
		}
	}

	visited[absPath] = mi
	return visited, nil
}

// collectExports collects exported symbol names from an export declaration.
func (mi *moduleInfo) collectExports(decl ast.Stmt) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		mi.exported[d.Name] = true
	case *ast.StructDecl:
		mi.exported[d.Name] = true
	case *ast.DeclStmt:
		mi.exported[d.Name] = true
	case *ast.MultiDeclStmt:
		for _, n := range d.Names {
			mi.exported[n] = true
		}
	}
}

// resolveImportPath resolves a relative import path against the importing file's directory.
func resolveImportPath(fromPath, from string) string {
	fromDir := filepath.Dir(fromPath)
	resolved := filepath.Join(fromDir, from)
	// Ensure .scar extension
	if !strings.HasSuffix(resolved, ".scar") {
		resolved += ".scar"
	}
	abs, err := filepath.Abs(resolved)
	if err != nil {
		return resolved
	}
	return abs
}

// bundleModules combines all module ASTs into a single program with symbol mangling.
func bundleModules(modules map[string]*moduleInfo, mainPath string) (*ast.Program, error) {
	mainMod, ok := modules[mainPath]
	if !ok {
		return nil, fmt.Errorf("main module %q not found", mainPath)
	}

	prog := &ast.Program{File: mainPath}

	// Add main module's statements first (no mangling for the entry module).
	for _, s := range mainMod.prog.Stmts {
		if s != nil {
			prog.Stmts = append(prog.Stmts, s)
		}
	}

	// Add dependencies' statements
	visited := make(map[string]bool)
	var addDeps func(path string)
	addDeps = func(path string) {
		if visited[path] {
			return
		}
		visited[path] = true
		mod, ok := modules[path]
		if !ok {
			return
		}
		// Add dependencies first
		for _, dep := range mod.deps {
			depPath := resolveImportPath(path, dep)
			addDeps(depPath)
		}
		// Add this module's statements (mangled)
		for _, s := range mod.prog.Stmts {
			mangled := mangleStmt(s, path, mod.exported)
			if mangled != nil {
				prog.Stmts = append(prog.Stmts, mangled)
			}
		}
	}

	for _, dep := range mainMod.deps {
		depPath := resolveImportPath(mainPath, dep)
		addDeps(depPath)
	}

	return prog, nil
}

// mangleStmt prefixes non-exported top-level declarations with the module path
// to avoid symbol collisions between modules.
func mangleStmt(s ast.Stmt, modulePath string, exported map[string]bool) ast.Stmt {
	modName := moduleBaseName(modulePath)

	switch d := s.(type) {
	case *ast.DeclStmt:
		if !exported[d.Name] {
			d.Name = modName + "_" + d.Name
		}
		return d
	case *ast.FuncDecl:
		if !exported[d.Name] {
			d.Name = modName + "_" + d.Name
		}
		return d
	case *ast.StructDecl:
		if !exported[d.Name] {
			d.Name = modName + "_" + d.Name
		}
		return d
	case *ast.ExportDecl:
		// Recursively mangle the inner declaration
		d.Decl = mangleStmt(d.Decl, modulePath, exported)
		return d
	case *ast.ImportStmt:
		// Import statements are removed after bundling
		return nil
	default:
		return s
	}
}

// moduleBaseName returns a short identifier for a module path.
func moduleBaseName(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	// Replace non-alphanumeric characters with underscore
	var sb strings.Builder
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	return sb.String()
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
