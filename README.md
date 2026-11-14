<p align="center">
  <img src="scar_script.png" alt="ScarScript Logo" width="200"/>
</p>

<h1 align="center">ScarScript</h1>

<p align="center">
  <strong>Pakistan's First High-Performance Compiled Language with Express-Like Syntax</strong>
</p>

<p align="center">
  <a href="https://scarscript.scarfaze.com/">
    <img src="https://img.shields.io/badge/Website-scarscript.scarfaze.com-0A66C2?style=flat-square" alt="Website"/>
  </a>
  <a href="https://scarscript.scarfaze.com/playground">
    <img src="https://img.shields.io/badge/Playground-Wasm%20Beta-00ADD8?style=flat-square" alt="Wasm Playground"/>
  </a>
  <a href="https://github.com/ScarfazeCompany/scar-script/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/ScarfazeCompany/scar-script/ci.yml?branch=main&style=flat-square" alt="CI"/>
  </a>
  <a href="https://github.com/ScarfazeCompany/scar-script/releases">
    <img src="https://img.shields.io/github/v/release/ScarfazeCompany/scar-script?style=flat-square" alt="Release"/>
  </a>
  <a href="https://github.com/ScarfazeCompany/scar-script/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/license-MIT%20OR%20Apache--2.0-blue.svg?style=flat-square" alt="MIT or Apache-2.0 License"/>
  </a>
</p>

---

<p align="center">
  <img src="https://scarscript.scarfaze.com/api/media/mw800shr-595fe8e70c.png" alt="ScarScript v1.0.3 Development Banner" width="100%"/>
</p>

## 📢 Language Development Status

> [!IMPORTANT]
> **ScarScript v1.0.3 (Under Testing & Benchmarking)**
> 
> ScarScript is under active development by **Team Scarfaze**. Following the release of **v1.0.2** (which introduced initial Language Server Protocol support for VS Code), our engineering team is running rigorous benchmarking and stabilization for **v1.0.3**.
> 
> Check out the [Version Notes & Migration Guide](https://scarscript.scarfaze.com/docs/reference/migration) and [Engineering Blog](https://scarscript.scarfaze.com/blog) for the latest updates. We appreciate your bug reports, feedback, and contributions as we continue building Pakistan's first high-performance compiled language!

---

## ⚡ Try ScarScript Online (WebAssembly Playground)

Test ScarScript directly in your browser without installing anything locally—powered by **WebAssembly (Wasm)** and built-in Emmet support:

👉 **[Launch ScarScript Interactive Playground](https://scarscript.scarfaze.com/playground)**

---

**ScarScript** is a statically-typed, compiled programming language that transpiles directly to Go. It combines the clean readability of Express.js-style syntax with the raw execution speed, type safety, and standalone binary compilation of Go.

### ✨ Key Features

| Feature | Description |
|---------|-------------|
| **Express-Like Syntax** | Clean, readable constructs (`func`, `val`/`mut`/`const`, `for ... in`, `if/else` with braces) |
| **Static Type Checking** | Full compile-time type inference with strict safety rules |
| **Multiple Returns** | Go-style `func() (int, bool)` with clean destructuring `val a, b = f()` |
| **Anonymous Functions** | First-class closures with variable capture: `val h = func(x) { ... }` |
| **`dec` Type** | 64-bit floating-point type with integer-literal widening (`val x: dec = 4`) |
| **Built-in Modules** | Zero-dependency standard modules: `math`, `file`, `sys`, `time` |
| **Transpiles to Go** | Generates self-contained, `gofmt`-validated `.gen.go` code |
| **Zero Runtime Overhead** | No interpreter, no VM — compiles straight to a native static binary |

---

### ⚡ Performance & Benchmarks

ScarScript transpiles directly to Go and compiles to native machine code—delivering **Go-level performance** with minimal boilerplate.

**Benchmark:** 10,000,000 loop iterations (`for i in 1..10_000_000 { }`)

| Language | Execution Time | Relative Speed | Notes |
|----------|---------------|----------------|-------|
| **ScarScript** | **0.165 s** | **1.00×** (baseline) | Native binary, zero runtime |
| **Go** (reference) | ~0.155 s | 1.06× faster | Direct Go implementation |
| **Node.js** (v20) | ~0.95 s | **5.8× slower** | V8 JIT, warmup required |
| **Python** (3.11) | ~3.8 s | **23× slower** | CPython interpreter |

> **Key Insight:** ScarScript adds **less than 7% overhead** compared to hand-written Go, while running **6–23× faster** than popular interpreted and JIT-compiled languages with zero warmup time.

```bash
# Reproduce the benchmark
cat > bench.scar << 'EOF'
mut sum = 0
for i in 1..10_000_000 {
    sum = sum + i
}
print("Sum: {sum}")
EOF

scar build bench.scar -o bench
time ./bench
```

---

### 🚀 Quick Start

#### 1. System Prerequisite

> [!NOTE]
> While downloading the pre-built ScarScript CLI binary gives you the compiler frontend immediately, executing (`scar run`) or compiling (`scar build`) scripts requires the **Go toolchain** (`go` 1.19+) to be installed and available in your system `$PATH`, as ScarScript invokes `go run` and `go build` under the hood.
> 
> * **Debian / Ubuntu:** `sudo apt-get install -y golang-go`
> * **macOS:** `brew install go`
> * **Windows / Others:** Install from [golang.org](https://golang.org/)

#### 2. Install ScarScript

**Option A — Download Pre-built CLI Binary:**

Grab the latest binary for your OS from [GitHub Releases](https://github.com/ScarfazeCompany/scar-script/releases):

| Platform | Binary File |
|----------|-------------|
| 🪟 Windows (x64) | `scar-windows-amd64.exe` |
| 🐧 Linux (x64) | `scar-linux-amd64` |
| 🍎 macOS (Intel) | `scar-darwin-amd64` |
| 🍎 macOS (Apple Silicon) | `scar-darwin-arm64` |

**Option B — Install via Go:**

```bash
go install [github.com/ScarfazeCompany/scar-script/cmd/scar@latest](https://github.com/ScarfazeCompany/scar-script/cmd/scar@latest)
```

**Option C — Build from Source:**

```bash
# Windows
.\scripts\build.ps1

# Linux / macOS
./scripts/build.sh
```

#### 3. Write & Run Your First Script (`hello.scar`)

```scar
func greet(name: str) {
    print("Hello, {name}!")
}

greet("Zawiyar")
```

```bash
# Run directly
scar run hello.scar
# Output: Hello, Zawiyar!

# Compile to a standalone native binary
scar build hello.scar -o hello
./hello
```

---

### 📖 Syntax Overview

```scar
// Variables (immutable by default)
val name: str = "Zawiyar"
mut count: int = 0
const API_URL: str = "[https://api.scar.com](https://api.scar.com)"

// Type inference
val inferred = 42          // int
val price = 99.99          // dec
val flag = true            // bool

// Functions
func add(a: int, b: int) int {
    return a + b
}

// Multiple returns (Go-style)
func divide(a: int, b: int) (int, bool) {
    if b == 0 {
        return 0, false
    }
    return a / b, true
}

val result, ok = divide(10, 2)

// Anonymous functions / closures
val multiply = func(x: int, y: int) int {
    return x * y
}

// Control flow
if age >= 18 and active {
    print("Access granted")
} else {
    print("Denied")
}

// Universal for loop
for i in 1..10 { ... }
for item in items { ... }
for k, v in map { ... }
for condition { ... }

// String interpolation
print("User {name} is {age} years old")

// Smart defaults (JS-style 'or')
val port = sys.env("PORT") or "8080"
val status = data.status or "unknown"

// Built-in Modules
val root = math.sqrt(16)           // 4.0
val rnd  = math.rand(1, 100)
file.write("log.txt", "started")
val cfg  = file.read("config.json")
time.sleep(1000)
val now  = time.now()
```

---

### 📦 Built-in Modules

| Module | Available Functions |
|--------|---------------------|
| **`math`** | `sqrt(dec) dec`, `pow(dec, dec) dec`, `abs(int) int`, `round(dec) int`, `rand(int, int) int` |
| **`file`** | `write(str, str)`, `append(str, str)`, `read(str) str`, `exists(str) bool`, `delete(str)` |
| **`sys`** | `env(str) str` |
| **`time`** | `now() str`, `sleep(int)` |

---

### 🛠 CLI Reference

```bash
scar run <file.scar>             # Transpile, compile & run immediately
scar build <file.scar> [-o out]  # Build standalone native binary
scar version                     # Print installed ScarScript version
scar help                        # Display CLI help & flags
```

---

### 🏗 Compiler Architecture

```text
scar run hello.scar
        │
        ▼
┌───────────────────┐
│  Lexer → Parser   │  (panic-based recovery, exact line/col positions)
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│    Type Checker   │  (phase-1 func registry, static analysis)
│  • infers types   │
│  • checks returns │
│  • tracks unused  │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│     Codegen       │  (emits self-contained .gen.go)
│  • Go functions   │
│  • prelude helpers│
│  • gofmt validated│
└─────────┬─────────┘
          │
          ▼
    go run / go build  →  native binary
```

---

### 📚 Documentation & Resources

Explore the official documentation, release history, and community channels:

* 🌐 **[Official Website](https://scarscript.scarfaze.com/)** — Overview, playground, and ecosystem links
* 📖 **[Documentation Hub](https://scarscript.scarfaze.com/docs)** — Complete language reference
* 🚀 **[Getting Started](https://scarscript.scarfaze.com/docs/getting-started)** — Installation and first steps
* 🔤 **[Language Syntax](https://scarscript.scarfaze.com/docs/language/syntax)** — Variables, control flow, and functions
* 🔄 **[Version Notes & Migration](https://scarscript.scarfaze.com/docs/reference/migration)** — Changelog and upgrade notes
* ✍️ **[Engineering Blog](https://scarscript.scarfaze.com/blog)** — Architecture deep-dives and updates
* ℹ️ **[About ScarScript](https://scarscript.scarfaze.com/about)** | **[Contact Team](https://scarscript.scarfaze.com/contact)**

---

### 🤝 Contributing

Contributions, bug reports, and feature suggestions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on submitting pull requests and reporting issues.

---

### 📄 License

ScarScript is dual-licensed. You may choose to use it under the terms of either:

* [MIT License](LICENSE)
* [Apache License 2.0](LICENSE-APACHE)

Choose the license that best fits your use case.

---

### 👨‍💻 Author

**Zawiyar Awan** — Founder & CEO, **[Scarfaze](https://scarfaze.com)**  
*Building Pakistan's First High-Performance Compiled Language with Express-Like Syntax*
