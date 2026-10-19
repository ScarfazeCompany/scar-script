---
name: Bug Report
about: Report a bug in the ScarScript compiler
title: "[BUG] "
labels: bug
assignees: ''
---

## Bug Description

A clear and concise description of what the bug is.

## Steps to Reproduce

1. Create a file `test.scar` with:
```scar
// Your code here
```

2. Run:
```bash
scar run test.scar
# or
scar build test.scar -o test
./test
```

3. Observe the error

## Expected Behavior

What you expected to happen.

## Actual Behavior

What actually happened. Include the full error output:

```
paste error output here
```

## Environment

- OS: [e.g., Windows 11, Ubuntu 22.04, macOS 14]
- Go version: `go version`
- ScarScript version: `scar version`
- Installation method: [go install / binary download / built from source]

## Additional Context

Any other context, logs, or screenshots.

## Checklist

- [ ] I have searched existing issues
- [ ] I have tested with the latest version
- [ ] I have included a minimal reproduction case
