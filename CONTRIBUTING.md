# Contributing to ScarScript

Thank you for your interest in contributing! 🎉

---

## Ways to Contribute

- 🐛 **Report bugs** — [GitHub Issues](https://github.com/ScarfazeCompany/scar-script/issues)
- 💡 **Request features** — [Discussions](https://github.com/ScarfazeCompany/scar-script/discussions)
- 📝 **Improve docs** — Fix typos, add examples
- 🔧 **Submit PRs** — Code changes, tests, performance

---

## Development Setup

### Prerequisites
- Go 1.22+
- Git

### Clone & Build
```bash
git clone https://github.com/ScarfazeCompany/scar-script.git
cd scar-script
go build ./...
go test ./...
```

### Run the Compiler
```bash
go run ./cmd/scar run examples/demo.scar
```

---

## Code Style

- **Go standard** — `gofmt`, `go vet`, `golint` clean
- **Tests required** — Every change needs test coverage
- **Documentation** — Public APIs need comments
- **Commits** — Conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`

---

## Pull Request Process

1. **Fork** the repository
2. **Create branch** — `feature/amazing-feature` or `fix/issue-123`
3. **Write tests** — Cover new behavior
4. **Run checks** — `go test ./... && go vet ./... && gofmt -l .`
5. **Update docs** — If user-facing change
6. **Submit PR** — Fill the template, link related issues

### PR Requirements
- [ ] All tests pass
- [ ] `go vet ./...` clean
- [ ] `gofmt -l .` clean
- [ ] No new linter warnings
- [ ] Changelog entry (for user-facing changes)

---

## Project Structure

```
.
├── cmd/scar/           # CLI entry point
├── internal/
│   ├── ast/            # Abstract syntax tree
│   ├── checker/        # Type checker
│   ├── codegen/        # Go code generator
│   ├── diag/           # Diagnostics/errors
│   ├── driver/         # Compile/run pipeline
│   ├── lexer/          # Lexical analyzer
│   ├── parser/         # Parser (Pratt)
│   ├── token/          # Token definitions
│   └── types/          # Type system
├── scripts/            # Build scripts
├── docs/               # Documentation (source of truth)
├── examples/           # Example programs
└── go.mod
```

---

## Testing

```bash
# All tests
go test ./...

# Specific package
go test ./internal/checker/...

# With coverage
go test -cover ./...
```

### Test Types
- **Unit** — Lexer, parser, type system
- **Integration** — Driver golden tests (`testdata/*.scar`)
- **Codegen** — Generated Go compiles & runs

---

## Adding Language Features

1. **Update spec** — `docs/spec.md` first
2. **Lexer** — Token definitions
3. **Parser** — AST construction
4. **Types** — Type definitions
5. **Checker** — Static rules
6. **Codegen** — Go emission
7. **Tests** — Unit + golden
8. **Docs** — `docs/spec.md`, `docs/stdlib.md`

---

## Code of Conduct

Be respectful, inclusive, and constructive. See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

---

## Questions?

Open a [Discussion](https://github.com/ScarfazeCompany/scar-script/discussions) or ping @zawiyar on GitHub.

