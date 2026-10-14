# Security Policy

## Supported Versions

We release security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | ✅ Yes             |
| < 1.0   | ❌ No              |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to **security@scarfaze.com**.

### What to Include

When reporting a vulnerability, please include:

1. **Description** — A clear description of the vulnerability
2. **Impact** — What an attacker could achieve
3. **Steps to Reproduce** — Detailed steps to trigger the issue
4. **Affected Components** — Which parts of the compiler (lexer, parser, checker, codegen, driver)
5. **Proof of Concept** — If possible, a minimal `.scar` file demonstrating the issue
6. **Suggested Fix** — If you have ideas for mitigation

### Response Timeline

- **Acknowledgment**: Within 48 hours
- **Initial Assessment**: Within 7 days
- **Fix Development**: Depends on severity (target: 30 days for critical)
- **Disclosure**: Coordinated with reporter after fix is released

### Disclosure Policy

We follow **coordinated disclosure**. We will:

1. Confirm the vulnerability
2. Develop and test a fix
3. Release a patch version (e.g., 1.0.1)
4. Publish a security advisory on GitHub
5. Credit the reporter (unless they request anonymity)

## Compiler Security Considerations

ScarScript is a compiler that:

- **Does not execute user code at compile time** — No macro expansion, no compile-time execution
- **Sandboxes generated code** — Generated Go code runs with user's permissions only
- **Validates all inputs** — Lexer/parser/checker reject malformed input with clear errors
- **No dynamic loading** — No plugin system, no runtime code loading

### Known Safe Behaviors

- `scar run` uses a temporary directory (cleaned up after execution)
- Generated Go code is `gofmt` formatted and type-checked
- No eval, no reflection on user input, no arbitrary file writes beyond explicit `file.write`

## Contact

- **Security Email**: security@scarfaze.com
- **PGP Key**: Available on request
- **General Issues**: [GitHub Issues](https://github.com/ScarfazeCompany/scar-script/issues) (non-security only)