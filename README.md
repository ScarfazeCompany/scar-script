<p align="center">
  <img src="scar_script.png" alt="ScarScript Logo" width="200"/>
</p>

<h1 align="center">ScarScript</h1>

<p align="center">
  <strong>Pakistan's First High-Performance Compiled Language with Express-Like Syntax</strong>
</p>

<p align="center">
  <a href="https://github.com/ScarfazeCompany/scar-script/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/ScarfazeCompany/scar-script/ci.yml?branch=main" alt="CI"/>
  </a>
  <a href="https://github.com/ScarfazeCompany/scar-script/releases">
    <img src="https://img.shields.io/github/v/release/ScarfazeCompany/scar-script" alt="Release"/>
  </a>
  <a href="https://github.com/ScarfazeCompany/scar-script/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/ScarfazeCompany/scar-script" alt="License"/>
  </a>
</p>

---

## 📢 Language Development Announcement

> [!IMPORTANT]
Language Development Status

ScarScript is under active development by Team Scarfaze. Version 1.0.2 adds an initial LSP server entry point for VS Code integration, following the scope and build-cache fixes in v1.0.1.

As we continue expanding our compiler features and standard library, you may encounter new experimental features or updates. We deeply appreciate your contributions and bug reports as we build Pakistan's first high-performance compiled language!

— Team Scarfaze

---

**ScarScript** is a statically-typed, compiled scripting language that transpiles to Go. It combines the readability of Express.js-style syntax with the performance and safety of Go.

### ✨ Key Features

| Feature | Description |
|---------|-------------|
| **Express-like Syntax** | `func`, `val`/`mut`/`const`, `for ... in`, `if/else` with braces |
| **Static Type Checking** | Full compile-time type inference with strict rules |
| **Multiple Returns** | Go-style `func() (int, bool)` with `val a, b = f()` |
| **Anonymous Functions** | Closures with capture: `val h = func(x) { ... }` |
| **`dec` Type** | 64-bit float with int-literal widening (`val x: dec = 4`) |
| **Built-in Modules** | `math`, `file`, `sys`, `time` — zero dependencies |
| **Transpiles to Go** | Self-contained `.gen.go` → `go build` / `go run` |
| **Zero Runtime** | No interpreter, no VM — pure Go binary |

---

### ⚡ Performance & Benchmarks

ScarScript **transpiles directly to Go** and compiles to a native binary — delivering **Go-level performance** with a cleaner syntax.

**Benchmark:** 10,000,000 loop iterations (`for i in 1..10_000_000 { }`)

| Language | Execution Time | Relative Speed | Notes |
|----------|---------------|----------------|-------|
| **ScarScript** | **0.165 s** | **1.00×** (baseline) | Native binary, zero runtime |
| **Go** (reference) | ~0.155 s | 1.06× faster | Direct Go implementation |
| **Node.js** (v20) | ~0.95 s | **5.8× slower** | V8 JIT, warmup required |
| **Python** (3.11) | ~3.8 s | **23× slower** | CPython interpreter |

> **Key Insight:** ScarScript adds **< 7% overhead** vs hand-written Go, while being **6–23× faster** than popular interpreted/JIT languages. No warmup, no JIT — just instant native speed.

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

**Option 1 — Download Binary (no Go needed):**

Grab your platform binary from [Releases](https://github.com/ScarfazeCompany/scar-script/releases):

| Platform | File |
|----------|------|
| 🪟 Windows | `scar-windows-amd64.exe` |
| 🐧 Linux | `scar-linux-amd64` |
| 🍎 macOS (Intel) | `scar-darwin-amd64` |
| 🍎 macOS (Apple Silicon) | `scar-darwin-arm64` |

**Option 2 — Go Install:**

```bash
go install github.com/ScarfazeCompany/scar-script/cmd/scar@latest
```

**Write your first script (hello.scar):**

```scar
func greet(name: str) {
    print("Hello, {name}!")
}

greet("Zawiyar")
```

**Run it:**

```bash
scar run hello.scar
# Hello, Zawiyar!

# Build a binary
scar build hello.scar -o hello
./hello
```

**Or build all platforms yourself:**

```bash
# Windows
.\scripts\build.ps1

# Linux / macOS
./scripts/build.sh
```

---

### 📖 Syntax Overview

```scar
// Variables (immutable by default)
val name: str = "Zawiyar"
mut count: int = 0
const API_URL: str = "https://api.scar.com"

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

// Modules
val root = math.sqrt(16)           // 4.0
val rnd  = math.rand(1, 100)
file.write("log.txt", "started")
val cfg  = file.read("config.json")
time.sleep(1000)
val now  = time.now()
```

---

### 📦 Built-in Modules

| Module | Functions |
|--------|-----------|
| **math** | `sqrt(dec) dec`, `pow(dec, dec) dec`, `abs(int) int`, `round(dec) int`, `rand(int, int) int` |
| **file** | `write(str, str)`, `append(str, str)`, `read(str) str`, `exists(str) bool`, `delete(str)` |
| **sys** | `env(str) str` |
| **time** | `now() str`, `sleep(int)` |

---

### 🛠 CLI Reference

```bash
scar run <file.scar>          # Compile & run (uses temp dir)
scar build <file.scar> [-o out]  # Build static binary
scar version                   # Show version
scar help                      # Show help
```

---

### 🏗 How It Works

```
scar run hello.scar
        │
        ▼
┌───────────────────┐
│  Lexer → Parser   │  (panic-based errors, precise positions)
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│    Type Checker   │  (phase-1 func registry, full static analysis)
│  • assigns types  │
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

### 📚 Documentation

- 🌐 **[Documentation Site](https://scarscript.scarfaze.com/docs/index.html)** — full docs with sidebar navigation
- [Getting Started](https://scarscript.scarfaze.com/docs/getting-started.html)
- [Language Syntax](https://scarscript.scarfaze.com/docs/syntax.html)
- [Type System](https://scarscript.scarfaze.com/docs/types.html)
- [Functions](https://scarscript.scarfaze.com/docs/functions.html)
- [Standard Library](https://scarscript.scarfaze.com/docs/stdlib.html)
- [CLI Reference](https://scarscript.scarfaze.com/docs/cli.html)

---

### 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

### 📄 License

MIT License — see [LICENSE](LICENSE) for details.

---

### 👨‍💻 Author

**Zawiyar Awan** — Founder, **Scarfaze**  
Pakistan's First High-Performance Compiled Language with Express-Like Syntax

