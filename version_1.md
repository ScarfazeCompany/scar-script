// 1. Immutable Variables (Default - Value change nahi ho sakti)
 

mut status: str = "pending"
status = "completed" // Valid

// 3. Global Compile-Time Constants
const API_URL: str = "https://api.scar.com"

// 4. Clean & Unique Collections Syntax
val tags: str[] = ["fast", "backend", "express"]
val scores: map[str: int] = {"zawiyar": 100, "kashif": 95}

// Flexible / Any type (Jab exact type pata na ho, e.g., JSON payload)
val rawData: any = { "status": 200 }


Sahi hai! Variables aur Types ka design finalize hone ke baad, ab aate hain Printing, String Operations, Conditionals (If/Else), Loops, aur Essential Tasks par.Iska goal yeh hai ke logic banate waqt zero mental friction ho aur readable lage.1. Printing & Logging (Smart & Clean)JavaScript ke console.log() ya Go ke fmt.Println() jaise lambe naam ke bajaye direct print() aur log().print(...): Direct values, variables, ya formatted strings screen / terminal par print karne ke liye.String Interpolation: String ke andar hi {varName} likhne se auto-formatting hogi (no + concatenation mess).Code snippetval name: str = "Zawiyar"
val age: int = 25

// Simple Printing
print("Hello World")

// String Interpolation (Direct variables inside string)
print("User {name} ki age {age} hai.")

// Multiple items
print(name, age, true)

// Advanced Logging (Backend ke error/debug logs ke liye)
log.info("Server started on port 8080")
log.error("Database connection failed")
2. Conditionals (If / Else)if aur else mein braces {} lazmi hain taake code clean rahe, lekin parentheses () optional hain taake extra typing na karni pare.Code snippetval age: int = 20
val hasToken: bool = true

// Clean logic without extra brackets ()
if age >= 18 and hasToken {
    print("Access Granted")
} else if age < 18 {
    print("Underage")
} else {
    print("Access Denied")
}
3. Ternary / Inline Guard (Quick Assignments)Logic ko chota aur readable rakhne ke liye single-line conditional check:Code snippetval isAdult: bool = age >= 18 ? true : false

// Guard / Default value syntax (Agar value null/empty ho)
val status: str = rawData.status or "default_status"
4. Loops (Simple & Unified for)Aam languages mein while, do-while, for, forEach alag hotay hain jo confuse karte hain. Is language mein sirf ek keyword for hoga jo har tarah ke loop ka kaam karega:A. Range Loop (Number Indexing)Code snippet// 1 se 5 tak loop chalaye ga
for i in 1..5 {
    print("Count: {i}")
}
B. Array / List IterationCode snippetval tags: str[] = ["fast", "backend", "express"]

// Element iterate karna
for tag in tags {
    print("Tag: {tag}")
}

// Index aur Value dono chahiye hon
for index, tag in tags {
    print("{index}: {tag}")
}
C. Map / Dictionary IterationCode snippetval scores: map[str: int] = {"zawiyar": 100, "kashif": 95}

for key, value in scores {
    print("{key} scored {value}")
}
D. Infinite / Conditional Loop (while ka replacement)Code snippetmut count: int = 0

// Jab tak condition match karti hai
for count < 5 {
    print(count)
    count++
}
5. Essential Helper Tasks (Arrays & Strings Operations)Common backend tasks ke liye built-in methods jo logic build karna super fast bana dein:Code snippetval name: str = "  Zawiyar Awan  "

// String Helpers
val cleanName: str = name.trim()        // Spacing remove
val upper: str = name.toUpper()         // Upper case
val isMatch: bool = name.contains("Zaw") // Check sub-string

// Array Helpers
mut items: str[] = ["apple", "banana"]
items.add("orange")                     // Append
items.remove(0)                         // Remove by index
val length: int = items.len()           // Get total items count
Quick Summary MatrixTaskSyntax LookBenefitPrintprint("User: {name}")Interpolation se clean string generationIf / Elseif age > 18 and active { ... }No () parenthesis clutterLoopingfor item in items / for i in 1..10Universal for loop across all typesDefaultsval x = value or "fallback"Null checking visual aur fastAapko yeh Printing, Conditionals, aur Loops ka syntax kaisa laga?Agar yeh clear hai, toh kya hum ab Functions (Input/Output Parameters) design karein ya direct HTTP Endpoint & Route Logic par chalein?


===================================================================================

Extended Features1. Math Module (math.*)Mathematics aur financial logic ke liye essential helpers:Code snippetval sq: dec = math.sqrt(16)        // Square Root -> 4.0
val pow: dec = math.pow(2, 3)      // Power -> 8.0
val absVal: int = math.abs(-45)     // Absolute -> 45
val rounded: int = math.round(4.7) // Rounding -> 5
val randomNum: int = math.rand(1, 100) // Random Number between 1 & 100
2. File System Module (file.*)Basic File Operations (Read, Write, Append, Delete, Check):Code snippet// 1. File Write (Overwrites or Creates)
file.write("notes.txt", "Hello from ScarScript!")

// 2. File Append (Adds to existing file)
file.append("notes.txt", "\nNew log line.")

// 3. File Read
val content: str = file.read("notes.txt")

// 4. File Existence & Delete
if file.exists("notes.txt") {
    print("File exists!")
    file.delete("notes.txt")
}
3. Basic System & Time Module (sys.* / time.*)Environment variables aur timing logic ke liye:Code snippetval now: str = time.now()           // Current Timestamp
val envPort: str = sys.env("PORT") or "8080" // Fetch System ENV Var
time.sleep(1000)                    // Sleep for 1000ms (1 second)
Background Implementation Blueprint (Go Transpile Mapping)Aapka compiler backend par in ScarScript functions ko Go ke native fast packages par clean syntax mapping dega:ScarScript SyntaxGo Standard Library Mappingmath.sqrt(x)math.Sqrt(float64(x))math.pow(x, y)math.Pow(float64(x), float64(y))file.read("path")os.ReadFile("path")file.write("path", "data")os.WriteFile("path", []byte("data"), 0644)file.append("path", "data")os.OpenFile("path", os.O_APPEND|os.O_WRONLY, 0644)file.delete("path")os.Remove("path")time.now()time.Now().Format(...)sys.env("PORT")os.Getenv("PORT")


Functions (func) ka scene ScarScript mein boht clean, statically-typed, aur Express/Go-style hybrid hoga!

Aap visual syntax, type-checking rules, aur backend execution model ka poora roadmap dekhein:

1. Function ka Basic Syntax
ScarScript mein function declare karne ke liye func keyword, parameters ke saath static types, aur -> (ya direct return type) use hoga:

Code snippet
// Simple function with return type
func add(a: int, b: int) int {
    return a + b
}

// Void function (jo kuch return na kare)
func greet(name: str) {
    print("Hello, {name}!")
}

// Usage
val result: int = add(10, 20)
greet("Zawiyar")
2. Advanced Function Features (V1.0 Extensions)
A. Multiple Return Values (Go-style - Super useful for Error Handling)
Code snippet
func divide(a: int, b: int) (int, bool) {
    if b == 0 {
        return 0, false // division failed
    }
    return a / b, true
}

val res, ok = divide(10, 2)
if ok {
    print("Result: {res}")
}
B. Anonymous Functions / Arrow Syntax (Express.js Handlers ke liye critical)
Code snippet
val multiply = func(x: int, y: int) int {
    return x * y
}

print(multiply(4, 5))
3. Compiler Backend Mechanics (Transpilation to Go)
Jab aap ScarScript mein function likhenge:

Code snippet
func calculateTotal(price: int, tax: int) int {
    return price + tax
}
Aapka Codegen engine isko instant Go code mein transpile kar dega:

Go
func calculateTotal(price int64, tax int64) int64 {
    return price + tax
}
4. TypeChecker Rules (Compile-Time Checks)
Function add karte waqt aapka checker.go 3 cheezein strict verify karega:

Parameter Count & Type Match: Calling time par agar add("hello", 5) pass kiya toh compile error aayega:

demo.scar:10:5: error: argument 1 expects int, got str

Return Type Enforcement: Function body ke andar return hone wali value ki type, signature ki return type se 100% match honi chahiye.

Dead Code Check: If/Else ke har path se return hona zaroori hoga agar function ki return type declared hai.

Implementation Scope
In functions ko support karne ke liye humen 3 files mein modifications karni hongi:

ast/ast.go: FuncDecl aur CallExpr nodes.

parser/parser.go: func keyword parse karne ka logic aur parameter lists.

checker/checker.go: Function symbol table aur scope resolver.

codegen/codegen.go: Go function definitions emission.
