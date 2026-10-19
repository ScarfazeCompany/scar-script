# ScarScript Standard Library Reference

> All standard library functions are accessed through built-in modules: `math`, `file`, `sys`, `time`.

---

## math

Mathematical operations. All numeric functions work with `dec` (float64) unless noted.

### `math.sqrt(x: dec) dec`
Returns the square root of `x`.
```scar
val root: dec = math.sqrt(16)    // 4.0
val root2 = math.sqrt(2)         // 1.414...
```

### `math.pow(x: dec, y: dec) dec`
Returns `x` raised to the power `y`.
```scar
val squared = math.pow(5, 2)     // 25.0
val cubed = math.pow(3, 3)       // 27.0
```

### `math.abs(x: int) int`
Returns the absolute value of `x` as `int`.
```scar
val n = math.abs(-45)            // 45
val m = math.abs(0)              // 0
```

### `math.round(x: dec) int`
Rounds `x` to the nearest integer (half away from zero).
```scar
val r1 = math.round(4.7)         // 5
val r2 = math.round(4.3)         // 4
val r3 = math.round(-2.5)        // -3
```

### `math.rand(lo: int, hi: int) int`
Returns a cryptographically **non-secure** random integer in the inclusive range `[lo, hi]`.
```scar
val dice = math.rand(1, 6)       // 1..6
val idx = math.rand(0, len-1)    // random index
```
> **Note:** Uses Go's `math/rand` (global source, auto-seeded since Go 1.20). Not suitable for cryptographic use.

---

## file

File system operations. All functions panic on I/O errors with descriptive messages.

### `file.write(path: str, content: str)`
Writes `content` to `path`, **overwriting** any existing file. Creates parent directories if needed.
```scar
file.write("config.json", `{"port": 8080}`)
```

### `file.append(path: str, content: str)`
Appends `content` to `path`. Creates file if it doesn't exist.
```scar
file.append("server.log", "Request received\n")
```

### `file.read(path: str) str`
Reads entire file and returns contents as string. Panics if file missing or unreadable.
```scar
val config = file.read("config.json")
```

### `file.exists(path: str) bool`
Returns `true` if path exists (file or directory).
```scar
if file.exists("data.bin") {
    val data = file.read("data.bin")
}
```

### `file.delete(path: str)`
Deletes the file at `path`. Panics on error (e.g., permission denied).
```scar
if file.exists("tmp.tmp") {
    file.delete("tmp.tmp")
}
```

---

## sys

System interaction.

### `sys.env(name: str) str`
Returns the value of environment variable `name`. Returns empty string if not set.
```scar
val port = sys.env("PORT") or "8080"
val db = sys.env("DATABASE_URL")
```

---

## time

Time and sleep utilities.

### `time.now() str`
Returns current timestamp as formatted string (`2006-01-02 15:04:05`).
```scar
val started = time.now()
print("Server started at {started}")
```

### `time.sleep(ms: int)`
Blocks the current thread for `ms` milliseconds.
```scar
print("Waiting...")
time.sleep(1000)   // 1 second
print("Done")
```

---

## Built-in Functions

These are **not** in modules — available globally.

### `print(args ...any)`
Prints arguments to stdout with space separation and newline. Supports string interpolation in string arguments.
```scar
print("Hello World")
print(name, age, true)
print("User {name} is {age}")
```

### `log.info(args ...any)`
Logs to stderr with `[INFO]` prefix and timestamp.
```scar
log.info("Server started on port 8080")
```

### `log.error(args ...any)`
Logs to stderr with `[ERROR]` prefix and timestamp.
```scar
log.error("Database connection failed")
```

---

## String Methods

Available on `str` values (including string literals and interpolated results).

| Method | Signature | Description |
|--------|-----------|-------------|
| `trim()` | `() str` | Remove leading/trailing whitespace |
| `toUpper()` | `() str` | Convert to uppercase |
| `contains(sub: str)` | `(str) bool` | Substring check |
| `len()` | `() int` | Byte length (UTF-8 code units) |

```scar
val s = "  hello  "
s.trim()           // "hello"
s.toUpper()        // "  HELLO  "
s.contains("ell")  // true
s.len()            // 10
```

---

## Array Methods

Available on array values (`T[]`).

| Method | Signature | Description |
|--------|-----------|-------------|
| `add(val: T)` | `() void` | Append (requires `mut`) |
| `remove(idx: int)` | `() void` | Remove at index (requires `mut`) |
| `len()` | `() int` | Element count |

```scar
mut items: str[] = ["a", "b"]
items.add("c")        // ["a", "b", "c"]
items.remove(0)       // ["b", "c"]
items.len()           // 2
```

---

## Map Methods

Available on map values (`map[K: V]`).

| Method | Signature | Description |
|--------|-----------|-------------|
| `len()` | `() int` | Entry count |
| Field access | `m.key` | Sugar for `m["key"]` (string keys only) |

```scar
val scores: map[str: int] = {"a": 1}
scores.len()     // 1
scores.a         // 1 (same as scores["a"])
```

---

## `or` Operator (Default/Fallback)

JS-style nullish/zero coalescing:
```scar
val a: str = "" or "default"     // "default"
val b: int = 0 or 42             // 42
val c: bool = false or true      // true
val d: dec = 0.0 or 3.14         // 3.14
val e: any = nil or "fallback"   // "fallback"
val f = getConfig() or "prod"    // any → string
```
- Left `any` → type-assertion to right's type
- Left `str`/`int`/`dec`/`bool` → zero-value check
- Left `bool` with right `bool` → logical `||`

---

## String Interpolation

Embed expressions directly in double-quoted strings:
```scar
val name = "Zawiyar"
val age = 25
print("User {name} is {age} years old")
print("2 + 2 = {2 + 2}")
```

- Expressions can be any valid expression
- Automatically converted via `_scar_str()` (handles str/int/bool/nil/any)
- Nested braces allowed: `"{a + (1 == 1 ? 2 : 3)}"`

---

## Escape Sequences (in strings)

| Sequence | Meaning |
|----------|---------|
| `\n` | Newline |
| `\t` | Tab |
| `\r` | Carriage return |
| `\\` | Backslash |
| `\"` | Double quote |
| `\'` | Single quote |
| `\{` | Literal `{` |
| `\}` | Literal `}` |

---

## Collections

### Array Literals
```scar
val nums: int[] = [1, 2, 3]
val names: str[] = ["a", "b"]
val empty: str[] = []           // requires annotation
```

### Map Literals
```scar
val scores: map[str: int] = {"alice": 100, "bob": 95}
val empty: map[str: int] = {}   // requires annotation
```
- Keys must be string literals or identifiers (converted to string)
- Trailing comma allowed

---

## Type Annotations

### Optional Inference
```scar
val x = 42              // int
val y = 3.14            // dec
val s = "hello"         // str
val b = true            // bool
```

### Explicit Types
```scar
val x: int = 42
val y: dec = 3.14
val s: str = "hello"
val b: bool = true
val any = anything
```

### Composite Types
```scar
val arr: int[] = [1, 2]
val map: map[str: int] = {"a": 1}
val tuple: (int, bool) = divide(10, 2)  // function return only
```

---

## Quick Reference Card

| Task | Syntax |
|------|--------|
| Immutable var | `val x = 1` |
| Mutable var | `mut x = 1` |
| Constant | `const PI = 3.14` |
| Function | `func add(a: int, b: int) int { return a + b }` |
| Void function | `func greet(name: str) { print(name) }` |
| Multi-return | `func div(a, b: int) (int, bool) { ... }` |
| Destructuring | `val q, r = div(10, 3)` |
| Anonymous fn | `val f = func(x: int) int { return x * 2 }` |
| If/else | `if x > 0 { ... } else { ... }` |
| Range loop | `for i in 1..5 { ... }` |
| Array loop | `for x in arr { ... }` |
| Map loop | `for k, v in m { ... }` |
| While-like | `for cond { ... }` |
| Interpolation | `"Hello {name}"` |
| Default value | `x or "fallback"` |
| Module call | `math.sqrt(16)` |

