# Migration Guide

> This guide covers breaking changes and migration paths between ScarScript versions.

---

## v0.x → v1.0.0 (Current)

First stable release. No previous version to migrate from.

### New in v1.0
- Full static type checking
- Functions with multiple returns
- Anonymous functions (closures)
- `dec` type (float64)
- Built-in modules: `math`, `file`, `sys`, `time`
- Optional type annotations with inference
- Multiple declarations from tuple returns
- `or` operator with JS-style fallback semantics

---

## Planned v1.1 (Backward Compatible)

### Additions
- `switch` expression
- `try/catch` for error handling (via FFI)
- `import` for user modules
- `struct` / `interface` types
- Generics (type parameters)

### Deprecations
None.

---

## Breaking Change Policy

| Version | Policy |
|---------|--------|
| Patch (1.0.x) | No breaking changes |
| Minor (1.x.0) | New features, deprecations with 1-cycle notice |
| Major (x.0.0) | Breaking changes allowed with migration guide |

---

## Migration Checklist (Future Major Versions)

When upgrading major versions:

1. **Run compiler on codebase** — capture all errors
2. **Fix type errors** — usually signature changes
3. **Update module calls** — renamed functions
4. **Test runtime behavior** — especially `or` semantics
5. **Rebuild binaries** — `scar build` all projects

---

## Version Support

| Version | Status | Supported Until |
|---------|--------|-----------------|
| 1.0.x | Current | TBD |

---

## Reporting Issues

Found a regression? Please report at:
https://github.com/ScarfazeCompany/scar-script/issues

Include:
- ScarScript version (`scar version`)
- Minimal reproduction
- Expected vs actual behavior

