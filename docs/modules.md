# ScarScript Module Reference

> Built-in modules provide system capabilities without external dependencies. All module names are reserved keywords.

---

## Module Overview

| Module | Purpose | Import |
|--------|---------|--------|
| `math` | Mathematical functions | Auto (on use) |
| `file` | File I/O operations | Auto (on use) |
| `sys` | System environment | Auto (on use) |
| `time` | Time & sleep | Auto (on use) |

> Modules are **not** imported explicitly — they are resolved at compile time when their functions are called. The code generator emits the necessary Go imports automatically.

---

## math — Mathematical Functions

All `math` functions work with `dec` (float64) except where noted. Int literals are automatically widened to `dec`.

### `math.sqrt(x: dec) dec`
```scar
val root = math.sqrt(16)        // 4.0
val hypotenuse = math.sqrt(3*3 + 4*4)  // 5.0
```
- Domain: `x >= 0` (negative → NaN, no panic)

### `math.pow(x: dec, y: dec) dec`
```scar
val area = math.pow(10, 2)      // 100.0
val growth = math.pow(2, 10)    // 1024.0
val root = math.pow(16, 0.5)    // 4.0 (same as sqrt)
```

### `math.abs(x: int) int`
Returns absolute value as `int`. Argument must be `int` (int literal allowed).
```scar
val diff = math.abs(-45)        // 45
val zero = math.abs(0)          // 0
// math.abs(3.14)  // ERROR: argument 1 expects int, got dec
```

### `math.round(x: dec) int`
Rounds to nearest integer (half away from zero). Returns `int`.
```scar
val a = math.round(4.7)         // 5
val b = math.round(4.3)         // 4
val c = math.round(-2.5)        // -3
val d = math.round(0.0)         // 0
```

### `math.rand(lo: int, hi: int) int`
Returns a pseudo-random integer in `[lo, hi]` (inclusive). Both args must be `int`.
```scar
val dice = math.rand(1, 6)      // 1..6
val index = math.rand(0, items.len() - 1)
```
- Uses Go's `math/rand.Intn` (global source, auto-seeded)
- **Not cryptographically secure**
- Panics if `hi < lo`

---

## file — File System Operations

All functions panic on I/O failure with a message like `scar: file.write: <error>`.

### `file.write(path: str, content: str)`
Creates/overwrites file at `path` with `content`. Creates parent directories if needed.
```scar
file.write("config.json", `{"debug": true}`)
file.write("data.txt", "Hello\nWorld")
```
- Mode: `0644`
- Atomic-ish: writes to temp then renames (Go's `WriteFile`)

### `file.append(path: str, content: str)`
Appends `content` to file. Creates file if missing.
```scar
file.append("access.log", "{time.now()} - Request\n")
```
- Opens with `O_APPEND|O_CREATE|O_WRONLY`, mode `0644`

### `file.read(path: str) str`
Reads entire file into a string.
```scar
val config = file.read("config.json")
val lines = file.read("data.txt")
```
- Panics if file doesn't exist or can't be read
- Returns entire contents (no trailing newline stripping)

### `file.exists(path: str) bool`
Returns `true` if path exists (file or directory).
```scar
if file.exists("cache.dat") {
    val data = file.read("cache.dat")
}
```
- Uses `os.Stat` — returns `false` for any error (including permission)

### `file.delete(path: str)`
Deletes the file at `path`.
```scar
file.delete("temp.tmp")
```
- Panics on error (including not found)
- Use `file.exists` first if deletion is optional

---

## sys — System Environment

### `sys.env(name: str) str`
Returns environment variable value, or empty string if unset.
```scar
val port = sys.env("PORT") or "8080"
val db = sys.env("DATABASE_URL")
val debug = sys.env("DEBUG") == "1"
```
- Direct mapping to Go's `os.Getenv`
- No caching — reads current process env each call

---

## time — Time & Sleep

### `time.now() str`
Returns current local time formatted as `2006-01-02 15:04:05`.
```scar
val started = time.now()
print("Booted at {started}")
// Output: Booted at 2026-09-27 17:48:46
```
- Uses `time.Now().Format("2006-01-02 15:04:05")`
- Local timezone (not UTC)

### `time.sleep(ms: int)`
Blocks current goroutine for `ms` milliseconds.
```scar
print("Waiting 2 seconds...")
time.sleep(2000)
print("Done")
```
- Argument must be non-negative `int`
- Uses `time.Sleep(time.Duration(ms) * time.Millisecond)`

---

## Codegen Import Mapping

| ScarScript Module | Go Import (auto-emitted) |
|-------------------|--------------------------|
| `math` | `_scarmath "math"` |
| `file` | `_scaros "os"` (already in prelude) |
| `sys` | `_scaros "os"` (already in prelude) |
| `time` | `_scartime "time"` |

> Only imported when module functions are actually used (detected during pass 1).

---

## Helper Functions (Prelude)

These are emitted in every generated program and used by module codegen:

| Helper | Used By | Description |
|--------|---------|-------------|
| `_scar_abs(x int) int` | `math.abs` | Absolute value |
| `_scar_round(x float64) int` | `math.round` | Round half-away-from-zero |
| `_scar_rand(lo, hi int) int` | `math.rand` | Random in range |
| `_scar_file_write(path, data string)` | `file.write` | Write file |
| `_scar_file_append(path, data string)` | `file.append` | Append file |
| `_scar_file_read(path string) string` | `file.read` | Read file |
| `_scar_file_exists(path string) bool` | `file.exists` | Check exists |
| `_scar_file_delete(path string)` | `file.delete` | Delete file |
| `_scar_time_now() string` | `time.now` | Formatted now |
| `_scar_time_sleep(ms int)` | `time.sleep` | Sleep ms |

---

## Error Messages

### Argument Type Errors
```
demo.scar:10:5: error: argument 1 expects dec, got str
```

### Module Not Found (reserved names)
```
demo.scar:5:1: error: "math" is a module — use math.<function>(...)
```

### Function Not Found in Module
```
demo.scar:3:12: error: math has no function "sin" (available: sqrt, pow, abs, round, rand)
```

---

## Best Practices

1. **Check existence before read/delete:**
   ```scar
   if file.exists("config.json") {
       val cfg = file.read("config.json")
   }
   ```

2. **Use `or` for env defaults:**
   ```scar
   val port = sys.env("PORT") or "8080"
   ```

3. **Avoid `math.rand` for security** — use crypto/rand via FFI if needed.

3. **File paths:** Use forward slashes (`/`) — works cross-platform in Go.

4. **Large files:** `file.read` loads entire file into memory — stream via FFI for huge files.

