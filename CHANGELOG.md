# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2026-10-19

### Fixed
- **Global variable access in functions**: `const` and `val` with literal initializers (int, dec, str, bool) are now accessible inside function scopes via parent-scope fallback chain. Previously all top-level variables were invisible to functions.
- **Mutable variable isolation**: `mut` variables and `val` with non-literal initializers correctly remain inaccessible in functions (emit compile-time error).
- **Documentation**: Removed `func` from reserved/prohibited keywords tables — now properly documented as a valid core function declaration keyword.
- **Codegen**: Package-level emission of `const` and literal-initialized `val` ensures global constants/immutable vars are visible to generated Go functions.

### Technical Details
- Checker: Added `isLiteral` tracking in `varInfo`; scope lookup allows parent-scope access only for `const` or literal-initialized `val` (not `mut`).
- Codegen: `isPkgLevel()` now returns true for `const` and `val` with literal initializers (`IntLit`, `FloatLit`, `BoolLit`, `StrLit`).
- All existing tests pass; new behavior verified with integration tests.

---

## [1.0.0] - 2026-10-14

### Added

- **Language Features**
  - Express-like syntax: `func`, `val`/`mut`/`const`, `for ... in`, `if/else`
  - Static type system: `int`, `dec`, `str`, `bool`, `void`, tuples
  - Type inference with optional annotations
  - Multiple return values with destructuring: `val a, b = f()`
  - Anonymous functions / closures with variable capture
  - Recursion with forward references (phase-1 function registry)
  - `dec` type with int-literal widening (`val x: dec = 4` → `4.0`)
  - Range operator: `1..10` (end-exclusive)
  - String interpolation: `"Hello {name}"`
  - JS-style `or` fallback: `val x = maybe() or "default"`
  - Reserved name protection (builtins, modules, Go keywords, `_` prefix)

- **Built-in Modules**
  - `math`: `sqrt`, `pow`, `abs`, `round`, `rand`
  - `file`: `write`, `append`, `read`, `exists`, `delete`
  - `sys`: `env`
  - `time`: `now`, `sleep`

- **Built-in Functions**
  - `print(value)` — formatted output with interpolation
  - `log(level, message)` — structured logging with timestamp

- **Compiler Pipeline**
  - Lexer → Parser (panic-based errors with positions)
  - Type Checker (phase-1 registry, isolated scopes, always-returns analysis)
  - Codegen (self-contained `.gen.go`, `gofmt` validated)
  - Two-pass import emission (only used modules imported)

- **CLI**
  - `scar run <file.scar>` — compile & execute
  - `scar build <file.scar> [-o out]` — build native binary
  - `scar version` / `scar help`

- **Quality Assurance**
  - Full test suite (lexer, parser, checker, codegen, driver golden tests)
  - `go test`, `go vet`, `gofmt` clean
  - CI pipeline (GitHub Actions)

- **Documentation**
  - Complete language specification
  - Standard library reference
  - Professional docs site (sidebar nav, search, responsive)

### Fixed

- Function scope isolation (named functions get isolated scope)
- Closure capture for anonymous functions
- Unreachable code detection
- Missing return detection
- Int-literal to dec assignment via `assignableVal`

### Changed

- Module path: `github.com/ScarfazeCompany/scar-script`

## [0.1.0] - 2026-09-27

### Added

- Initial compiler skeleton
- Lexer with token positions
- Pratt parser for expressions
- Basic type checker (int, str, bool)
- Codegen to Go
- `scar run` / `scar build` CLI
- Basic test infrastructure
