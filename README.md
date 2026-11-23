<p align="center">
  <img src="scar_script.png" alt="ScarScript logo" width="200">
</p>

<h1 align="center">ScarScript</h1>

<p align="center">
  <strong>Pakistan's First High-Performance Compiled Language with Express-Like Syntax</strong>
</p>

<p align="center">
  <a href="https://scarscript.scarfaze.com/">
    <img src="https://img.shields.io/badge/Website-scarscript.scarfaze.com-0A66C2?style=flat-square" alt="Official website">
  </a>
  <a href="https://scarscript.scarfaze.com/playground">
    <img src="https://img.shields.io/badge/Playground-Wasm%20Beta-00ADD8?style=flat-square" alt="WebAssembly playground">
  </a>
  <a href="https://github.com/ScarfazeCompany/scar-script/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/ScarfazeCompany/scar-script/ci.yml?branch=main&style=flat-square" alt="CI status">
  </a>
  <a href="https://github.com/ScarfazeCompany/scar-script/releases">
    <img src="https://img.shields.io/github/v/release/ScarfazeCompany/scar-script?style=flat-square" alt="Latest release">
  </a>
  <a href="https://github.com/ScarfazeCompany/scar-script/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/license-MIT%20OR%20Apache--2.0-blue.svg?style=flat-square" alt="MIT or Apache-2.0 license">
  </a>
</p>

<p align="center">
  <img src="https://scarscript.scarfaze.com/api/media/mw800shr-595fe8e70c.png" alt="ScarScript v1.0.3 release banner" width="100%">
</p>

---

## Language Status

**ScarScript v1.0.3 is officially released by Team Scarfaze.**

Building on v1.0.2, which introduced initial Language Server Protocol (LSP) support for VS Code, v1.0.3 adds:

- Loop control statements: `continue`, `break`, and `stop`
- User-defined `struct` types and field mutation
- `Option<T>` and `Result<T, E>` for explicit optional-value and error handling
- A file-based module system using `import`, `export`, and `from`
- Circular-dependency detection for imported modules

See the [Version Notes & Migration Guide](https://scarscript.scarfaze.com/docs/reference/migration) and the [Engineering Blog](https://scarscript.scarfaze.com/blog) for more information.

**Performance benchmarks:** [View the official v1.0.2 benchmark results](./BENCHMARKS.md).

---

## Try ScarScript Online

You can experiment with ScarScript in your browser without installing the CLI.

The interactive playground is powered by WebAssembly (Wasm) and includes built-in Emmet support.

**[Open the ScarScript Playground](https://scarscript.scarfaze.com/playground)**

---

## What Is ScarScript?

ScarScript is a statically typed, compiled programming language that transpiles directly to Go.

It combines clean, Express.js-style syntax with Go's execution performance, type safety, and ability to produce standalone native binaries. ScarScript does not run through a language interpreter or virtual machine; its generated Go code is compiled using the Go toolchain.

## Key Features

| Feature | Description |
|---|---|
| Express-like syntax | Readable constructs such as `func`, `val`, `mut`, `const`, `for ... in`, and brace-based `if`/`else` statements. |
| Static type checking | Compile-time type inference and type-safety checks. |
| Multiple return values | Go-style multiple returns, such as `func() (int, bool)`, with destructuring using `val a, b = f()`. |
| Anonymous functions | First-class anonymous functions and closures with variable capture. |
| `dec` type | A 64-bit floating-point type, including integer-literal widening (for example, `val x: dec = 4`). |
| Built-in modules | Standard modules for `math`, `file`, `sys`, and `time`. |
| Go transpilation | Generates self-contained `.gen.go` files and validates generated code with `gofmt`. |
| Native compilation | Uses the Go toolchain to build native binaries without a ScarScript interpreter or VM. |
| v1.0.3 language features | Loop control, user-defined structs, `Option<T>`, `Result<T, E>`, and file-based modules with cycle detection. |

---

## Performance & Benchmarks

ScarScript transpiles to Go, which is then compiled to native machine code.

The following benchmark compares a loop containing 10,000,000 iterations:

```scar
mut sum = 0

for i in 1..10_000_000 {
    sum = sum + i
}

print("Sum: {sum}")
```

### Reported benchmark results

| Language | Execution time | Relative speed | Notes |
|---|---:|---:|---|
| **ScarScript** | **0.165 s** | **1.00× (baseline)** | Native binary, no ScarScript runtime |
| Go (reference) | ~0.155 s | 1.06× faster | Direct Go implementation |
| Node.js (v20) | ~0.95 s | 5.8× slower | V8 JIT; warm-up required |
| Python (3.11) | ~3.8 s | 23× slower | CPython interpreter |

**Benchmark takeaway:** In this test, ScarScript's execution time was less than 7% higher than the direct Go implementation. It was also faster than the listed Node.js and Python runs.

These figures describe the benchmark above and may vary with hardware, operating system, compiler versions, and benchmark methodology.

### Reproduce the benchmark

Save the following as `bench.scar`:

```scar
mut sum = 0

for i in 1..10_000_000 {
    sum = sum + i
}

print("Sum: {sum}")
```

Build and run it:

```bash
scar build bench.scar -o bench
./bench
```

On Windows, run the generated executable using its `.exe` filename if applicable.

---

## Quick Start

### 1. Install the Go toolchain

The ScarScript CLI provides the compiler frontend, but running or building a ScarScript program also requires the Go toolchain.

ScarScript invokes Go commands under the hood:

- `scar run` uses the Go toolchain to run the generated program.
- `scar build` uses the Go toolchain to build a native executable.

Install **Go 1.19 or later** and make sure the `go` command is available in your system `PATH`.

**Debian / Ubuntu**

```bash
sudo apt-get update
sudo apt-get install -y golang-go
```

**macOS**

```bash
brew install go
```

**Windows / other platforms**

Download and install Go from the [official Go website](https://golang.org/).

Verify the installation:

```bash
go version
```

### 2. Install ScarScript

Choose one of the following methods.

#### Option A — Download a pre-built CLI binary

Download the latest release for your operating system from the [GitHub Releases page](https://github.com/ScarfazeCompany/scar-script/releases).

| Platform | Release binary |
|---|---|
| Windows (x64) | `scar-windows-amd64.exe` |
| Linux (x64) | `scar-linux-amd64` |
| macOS (Intel) | `scar-darwin-amd64` |
| macOS (Apple Silicon) | `scar-darwin-arm64` |

After downloading, make the binary available from your terminal (for example, by adding its directory to `PATH`).

#### Option B — Install using Go

```bash
go install github.com/ScarfazeCompany/scar-script/cmd/scar@latest
```

Make sure your Go binary directory is included in your system `PATH`.

#### Option C — Build from source

Clone the repository, enter the project directory, and run the build script for your platform.

**Windows (PowerShell)**

```powershell
.\scripts\build.ps1
```

**Linux / macOS**

```bash
./scripts/build.sh
```

### 3. Create your first ScarScript program

Create a file named `hello.scar`:

```scar
func greet(name: str) {
    print("Hello, {name}!")
}

greet("Zawiyar")
```

Run the program:

```bash
scar run hello.scar
```

Expected output:

```text
Hello, Zawiyar!
```

### 4. Build a standalone executable

```bash
scar build hello.scar -o hello
```

Run the generated executable:

```bash
./hello
```

On Windows, run the generated `.exe` file.

---

## Language Syntax Overview

The examples below demonstrate the core ScarScript syntax.

### Variables and type inference

```scar
// Immutable by default
val name: str = "Zawiyar"

// Mutable variable
mut count: int = 0

// Constant
const API_URL: str = "https://api.scar.com"

// Type inference
val inferred = 42       // int
val price = 99.99       // dec
val flag = true         // bool
```

### Functions

```scar
func add(a: int, b: int) int {
    return a + b
}
```

### Multiple return values

ScarScript supports Go-style multiple return values.

```scar
func divide(a: int, b: int) (int, bool) {
    if b == 0 {
        return 0, false
    }

    return a / b, true
}

val result, ok = divide(10, 2)
```

### Anonymous functions and closures

```scar
val multiply = func(x: int, y: int) int {
    return x * y
}
```

### Conditional statements

```scar
if age >= 18 and active {
    print("Access granted")
} else {
    print("Denied")
}
```

### Loops

ScarScript supports several `for` loop forms:

```scar
// Range loop
for i in 1..10 {
    print(i)
}

// Iterate over a collection
for item in items {
    print(item)
}

// Iterate over a map
for k, v in map {
    print("{k}: {v}")
}

// Condition-based loop
for condition {
    // Loop body
}
```

### String interpolation

```scar
print("User {name} is {age} years old")
```

### Default values with `or`

```scar
val port = sys.env("PORT") or "8080"
val status = data.status or "unknown"
```

### Built-in module examples

```scar
val root = math.sqrt(16)
val randomNumber = math.rand(1, 100)

file.write("log.txt", "started")
val config = file.read("config.json")

time.sleep(1000)
val currentTime = time.now()
```

---

## v1.0.3 Release Features

This release adds loop control, structs, safer value handling, and a file-based module system.

### 1. Loop control: `continue`, `break`, and `stop`

Use `continue` to skip the current iteration. Use `break` or `stop` to exit a loop.

Using these statements outside a loop is reported as a compile-time semantic error.

```scar
for i in 1..10 {
    if i == 2 {
        continue
    }

    if i == 5 {
        break
    }

    if i == 8 {
        stop
    }

    print("i = {i}")
}
```

### 2. User-defined structs and field mutation

Declare a struct with typed fields, create an instance using a struct literal, and access or update its fields.

Field mutation is supported on `mut` bindings. Integer-to-`dec` widening is also supported where applicable.

```scar
struct User {
    id: int
    name: str
    score: dec
}

mut user = User {
    id: 1,
    name: "Zawiyar",
    score: 95
}

user.score = 98.5

print("{user.name} (#{user.id}) -> {user.score}")
```

### 3. Optional values and error handling: `Option<T>` and `Result<T, E>`

ScarScript provides `Option<T>` and `Result<T, E>` for representing optional values and successful or failed operations.

`Option<T>` uses:

- `Some(value)` when a value is present
- `None` when no value is available

`Result<T, E>` uses:

- `Ok(value)` for a successful result
- `Err(error)` for an error result

```scar
func findUser(id: int) Option<str> {
    if id == 1 {
        return Some("Zawiyar")
    }

    return None
}

func parsePort(raw: str) Result<int, str> {
    if raw == "8080" {
        return Ok(8080)
    }

    return Err("Invalid port: {raw}")
}

val found = findUser(1)

if found.isSome() {
    print("Found: {found.unwrap()}")
}

print("Fallback: {findUser(99).unwrapOr(\"Guest\")}")

val portRes = parsePort("invalid")

if portRes.isErr() {
    print("Error: {portRes.unwrapErr()}")
}

print("Port: {parsePort(\"invalid\").unwrapOr(3000)}")
```

### 4. File-based module system

ScarScript v1.0.3 supports organizing programs across `.scar` files using `export`, `import`, and `from`.

The module system includes circular-dependency detection and export validation.

**`math_utils.scar`**

```scar
export func square(n: int) int {
    return n * n
}
```

**`main.scar`**

```scar
import { square } from "./math_utils.scar"

print("Square of 6 is {square(6)}")
```

---

## Built-in Modules

ScarScript currently includes the following standard modules.

| Module | Available functions | Description |
|---|---|---|
| `math` | `sqrt(dec) dec`, `pow(dec, dec) dec`, `abs(int) int`, `round(dec) int`, `rand(int, int) int` | Mathematical operations and random number generation |
| `file` | `write(str, str)`, `append(str, str)`, `read(str) str`, `exists(str) bool`, `delete(str)` | File operations |
| `sys` | `env(str) str` | Environment variable access |
| `time` | `now() str`, `sleep(int)` | Current time and sleep operations |

---

## CLI Reference

| Command | Description |
|---|---|
| `scar run <file.scar>` | Transpile, compile, and run a ScarScript file |
| `scar build <file.scar> [-o out]` | Build a standalone native executable |
| `scar version` | Display the installed ScarScript version |
| `scar help` | Display CLI help and available flags |

Examples:

```bash
# Run a program
scar run hello.scar

# Build a standalone executable
scar build hello.scar -o hello

# Display the installed version
scar version

# Display help
scar help
```

---

## Compiler Architecture

ScarScript follows a compilation pipeline that converts source code into generated Go code and then uses the Go toolchain to run or build it.

```text
ScarScript source
       |
       v
+----------------------+
|   Lexer and Parser   |
|                      |
| - Tokenization       |
| - Syntax parsing     |
| - Line/column errors |
+----------+-----------+
           |
           v
+----------------------+
|     Type Checker     |
|                      |
| - Function registry  |
| - Type inference     |
| - Return checking    |
| - Unused tracking    |
+----------+-----------+
           |
           v
+----------------------+
|       Codegen        |
|                      |
| - Generates Go code  |
| - Emits prelude      |
| - Validates with     |
|   gofmt              |
+----------+-----------+
           |
           v
+----------------------+
|     Go Toolchain     |
|                      |
|   go run / go build  |
+----------+-----------+
           |
           v
     Native executable
```

---

## Documentation & Resources

- **[Official Website](https://scarscript.scarfaze.com/)** — Project overview and ecosystem links.
- **[Interactive Playground](https://scarscript.scarfaze.com/playground)** — Try ScarScript in your browser.
- **[Version Notes & Migration Guide](https://scarscript.scarfaze.com/docs/reference/migration)** — Release and migration information.
- **[Engineering Blog](https://scarscript.scarfaze.com/blog)** — Engineering articles and project updates.
- **[GitHub Repository](https://github.com/ScarfazeCompany/scar-script)** — Source code and issue tracking.
- **[GitHub Releases](https://github.com/ScarfazeCompany/scar-script/releases)** — Download release binaries.
- **[Benchmark Results](./BENCHMARKS.md)** — ScarScript v1.0.2 performance benchmarks.
- **[License](https://github.com/ScarfazeCompany/scar-script/blob/main/LICENSE)** — License terms.

---

## Roadmap

### v1.0.3 — Released

- Loop control: `continue`, `break`, and `stop`
- User-defined structs and field mutation
- `Option<T>` and `Result<T, E>`
- File-based module system with `import`, `export`, and `from`
- Circular-dependency detection

### v1.0.4 — Planned

- Zero-setup embedded database engine
- `db` module with CRUD operations
- HTTP web server and routing
- `web` module with `route`, `need`, `guard`, `limit`, and `reply`
- `env.get`
- Recurring tasks

### v1.1.0 — Planned

- Additional backend and standard-library capabilities
- Further performance optimizations

---

## Contributing

Contributions, bug reports, and feature suggestions are welcome.

Please read [`CONTRIBUTING.md`](./CONTRIBUTING.md) for the project's contribution guidelines, including information about submitting pull requests and reporting issues.

---

## License

ScarScript is dual-licensed. You may choose to use it under the terms of either:

- [MIT License](https://github.com/ScarfazeCompany/scar-script/blob/main/LICENSE)
- [Apache License 2.0](https://github.com/ScarfazeCompany/scar-script/blob/main/LICENSE)

Choose the license that best fits your use case.

---

## Author

**Zawiyar Awan**  
Founder & CEO, **Scarfaze**

Building Pakistan's First High-Performance Compiled Language with Express-Like Syntax.

<p align="center">
  <a href="https://scarscript.scarfaze.com/">ScarScript Official Website</a>
</p>
