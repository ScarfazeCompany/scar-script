# ScarScript Language Specification (v1)

## 1. Lexical Structure

### 1.1 Source Encoding
UTF-8 encoded text files with `.scar` extension.

### 1.2 Keywords
Reserved and cannot be used as identifiers:
```
val mut const if else for in and or true false map
str int bool any dec func return
```

### 1.3 Built-in Names
Cannot be redeclared (shadowing is an error):
```
print log math file sys time
```

### 1.4 Go Keywords
All Go keywords are reserved:
```
break case chan continue default defer else fallthrough
for func go goto if import map package range return select
struct switch type var
```

### 1.5 Identifiers
- Must start with letter or `_` (but `_` prefix is **forbidden** for user declarations)
- May contain letters, digits, `_`
- Case-sensitive

### 1.6 Literals
| Type | Syntax | Example |
|------|--------|---------|
| Integer | Decimal digits | `42`, `-5` |
| Float (dec) | Digits `.` digits (`.digit` required) | `3.14`, `-0.5`, `4.0` |
| Boolean | `true` / `false` | `true` |
| String | Double quotes, supports escapes & interpolation | `"Hello {name}"` |

**Note:** `1..5` is a range operator (two INT tokens), not a float.

### 1.7 Operators & Punctuation
```
= == != < > <= >= + - * / % ++ --
.. ... ? : , ; ( ) { } [ ] .
```

---

## 2. Types

### 2.1 Primitive Types
| Keyword | Go Type | Description |
|---------|---------|-------------|
| `str` | `string` | UTF-8 string |
| `int` | `int` | Signed integer (platform width) |
| `bool` | `bool` | `true` / `false` |
| `dec` | `float64` | IEEE-754 double precision |
| `any` | `any` | Top type (interface{}) |
| `void` | — | No value (function return only) |

### 2.2 Composite Types
```
str[]          // array of strings
map[str: int]  // map with string keys, int values
(int, bool)    // tuple (function return only)
```

### 2.3 Type Compatibility
- **Assignable:** exact match OR `T → any` widening
- **Int literal → dec:** allowed in declarations, arguments, returns
  - `val d: dec = 4` ✓
  - `val d: dec = someIntVar` ✗
- **Numeric promotion:** `int` + `dec` → `dec` (codegen emits `float64()`)

---

## 3. Variables & Constants

### 3.1 Declarations
```scar
val name: str = "Zawiyar"   // immutable (default)
mut count: int = 0          // mutable
const PI: dec = 3.14159     // compile-time constant
```

### 3.2 Type Inference
Annotation optional when initializer present:
```scar
val x = 42          // int
val y = 3.14        // dec
val s = "hello"     // str
```

### 3.3 Multiple Declaration (tuple destructuring)
```scar
val a, b = divide(10, 2)  // requires function returning tuple
mut x, y = f()
```
- Only `val`/`mut` (no `const`)
- No type annotation allowed
- RHS must be a single call expression returning a tuple

### 3.4 Constants
- Types allowed: `str`, `int`, `bool` only
- Initializer must be a literal (or unary `-` on int literal)

---

## 4. Expressions

### 4.1 Precedence (highest to lowest)
1. Primary: literals, identifiers, calls, indexing, member access, `(expr)`
2. Unary: `-` `!`
3. Multiplicative: `*` `/` `%`
4. Additive: `+` `-`
5. Range: `..`
6. Comparison: `<` `>` `<=` `>=`
7. Equality: `==` `!=`
8. Logical AND: `and`
9. Logical OR / Default: `or`
10. Ternary: `? :`

### 4.2 Arithmetic
- `int` + `int` → `int`
- `dec` + `dec` → `dec`
- `int` + `dec` → `dec` (int side wrapped in `float64()`)
- `%` requires `int` operands only

### 4.3 Comparison
- `int`/`dec`/`str` comparable with same type
- Mixed numeric: promoted to `dec`
- Result: `bool`

### 4.4 String Concatenation
- `str` + `str` → `str`

### 4.5 Logical Operators
- `and`: `bool and bool` → `bool` (short-circuit `&&`)
- `or`: three modes:
  1. `bool or bool` → `bool` (`||`)
  2. `any or T` → `T` (type-assertion fallback)
  3. `T or T` where `T` ∈ {str, int, dec} → `T` (zero-value check)

### 4.6 Ternary
```scar
val result = condition ? valueIfTrue : valueIfFalse
```
- Branches must have identical type
- Compiles to IIFE: `func() T { if c { return a }; return b }()`

### 4.7 Calls & Member Access
```scar
print("hello")           // builtin
math.sqrt(16)            // module function
arr.add("x")             // method
map.key                  // map field access (sugar for map["key"])
arr[0]                   // indexing
```

---

## 5. Statements

### 5.1 Expression Statement
Only function calls allowed (result discarded):
```scar
greet("Zawiyar")
divide(10, 2)            // tuple result discarded
```

### 5.2 Assignment
```scar
mut x: int = 1
x = 2                    // allowed
x++                      // inc/dec (int only)
```
- Immutable (`val`/`const`) cannot be assigned
- Type checked against declared type

### 5.3 Blocks
```scar
{
    val x = 1
    print(x)
}
```
- New scope for variables
- `return` inside block returns from enclosing function

### 5.4 If / Else
```scar
if condition {
    // ...
} else if other {
    // ...
} else {
    // ...
}
```
- Braces mandatory
- Parentheses optional around condition

### 5.5 For Loops (universal)
| Form | Syntax | Iterates |
|------|--------|----------|
| Range | `for i in 1..10` | integers inclusive |
| Array (elem) | `for x in arr` | elements |
| Array (idx, elem) | `for i, x in arr` | index + element |
| Map (key) | `for k in m` | keys |
| Map (key, val) | `for k, v in m` | key + value |
| Condition | `for cond {}` | while-like |

### 5.6 Return
```scar
return              // void function
return 42           // single value
return a, b         // tuple (matching function signature)
```
- Non-void functions **must** return on every path (checker enforces)
- Code after `return` in same block → "unreachable code" error

---

## 6. Functions

### 6.1 Named Functions (top-level only)
```scar
func add(a: int, b: int) int {
    return a + b
}

func greet(name: str) {       // void
    print("Hello, {name}!")
}

func divide(a: int, b: int) (int, bool) {
    if b == 0 { return 0, false }
    return a / b, true
}
```
- **Isolated scope:** only parameters visible (no capture of outer variables)
- Forward references allowed (phase-1 registration)
- Recursion & mutual recursion supported

### 6.2 Anonymous Functions (closures)
```scar
val multiply = func(x: int, y: int) int {
    return x * y
}

val addTen = func(x: int) int {
    return x + offset   // captures outer 'offset'
}
```
- Only valid as direct RHS of `val`/`mut` declaration
- Full closure capture (lexical scope)
- Optional return type annotation

### 6.3 Function Types
- Not directly writable in annotations
- Inferred from literals or declarations
- Used as values: `val f = func() int { return 1 }`

---

## 7. Modules

Module names are reserved keywords. Call as `module.function(...)`.

### 7.1 math
| Function | Signature | Description |
|----------|-----------|-------------|
| `sqrt` | `(x: dec) dec` | Square root |
| `pow` | `(x: dec, y: dec) dec` | Power |
| `abs` | `(x: int) int` | Absolute value |
| `round` | `(x: dec) int` | Round to nearest integer |
| `rand` | `(lo: int, hi: int) int` | Random integer in `[lo, hi]` |

### 7.2 file
| Function | Signature | Description |
|----------|-----------|-------------|
| `write` | `(path: str, content: str)` | Write (overwrites) |
| `append` | `(path: str, content: str)` | Append (creates) |
| `read` | `(path: str) str` | Read entire file |
| `exists` | `(path: str) bool` | Path exists |
| `delete` | `(path: str)` | Remove file |

### 7.3 sys
| Function | Signature | Description |
|----------|-----------|-------------|
| `env` | `(name: str) str` | Environment variable (empty if unset) |

### 7.4 time
| Function | Signature | Description |
|----------|-----------|-------------|
| `now` | `() str` | Current timestamp (RFC3339-like) |
| `sleep` | `(ms: int)` | Sleep milliseconds |

---

## 8. Error Handling

### 8.1 Compile-Time Errors
All errors reported as:
```
file.scar:line:col: error: message
```
Exit code: `1`

### 8.2 Common Error Classes
| Category | Example |
|----------|---------|
| Type mismatch | `cannot use value of type "str" where "int" expected` |
| Argument count | `expected 2 argument(s), got 1` |
| Argument type | `argument 1 expects int, got str` |
| Missing return | `missing return — function must return "int"` |
| Unreachable | `unreachable code` |
| Undefined | `undefined variable "x"` / `undefined function "foo"` |
| Redeclaration | `"x" is already declared in this scope` |
| Reserved name | `"print" is a builtin name and cannot be redeclared` |
| Module misuse | `"math" is a module — use math.<function>(...)` |

---

## 9. Compiler Internals

### 9.1 Pipeline
```
Lexer → Parser (panic errors) → Checker (first error aborts) → Codegen (gofmt) → go build/run
```

### 9.2 Generated Code Structure
```go
package main

import (
    _scarfmt "fmt"
    _scarlog "log"
    _scaros "os"
    _scarreflect "reflect"
    _scarstrconv "strconv"
    _scarstrings "strings"  // if needed
    _scarmath "math"        // if math.* used
    _scartime "time"        // if time.* used
    _scarrand "math/rand"   // if math.rand used
)

// Prelude helpers: _scar_str, _scar_or, _scar_or_default, _scar_empty,
// _scar_field, _scar_add, _scar_remove, _scar_index,
// _scar_abs, _scar_round, _scar_file_*, _scar_time_*, _scar_rand

// Top-level functions (from func declarations)

func main() {
    // Variable declarations
    // Statements
}
```

### 9.3 Naming Conventions
- All generated identifiers prefixed with `_scar_` or `_scar` import aliases
- User identifiers starting with `_` rejected at compile time

---

## 10. Version

Current: **v1.0.0** — specification frozen.

