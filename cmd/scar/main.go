package main

import (
	"fmt"
	"os"

	"github.com/ScarfazeCompany/scar-script/internal/driver"
	"github.com/ScarfazeCompany/scar-script/internal/lsp"
)

const version = "1.0.2"

const usage = `ScarScript compiler v` + version + `

Usage:
  scar run <file.scar>              Compile and run a program
  scar build <file.scar> [-o out]   Compile to a native binary
  scar version                      Print version
  scar help                         Show this help
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "run":
		if len(os.Args) != 3 {
			fail("usage: scar run <file.scar>")
		}
		if err := driver.Run(os.Args[2], os.Stdout, os.Stderr); err != nil {
			fail(err.Error())
		}

	case "build":
		args := os.Args[2:]
		out := ""
		if len(args) >= 3 && args[1] == "-o" {
			out = args[2]
			args = args[:1]
		}
		if len(args) != 1 {
			fail("usage: scar build <file.scar> [-o out]")
		}
		if out == "" {
			out = driver.DefaultOutputPath(args[0])
		}
		if err := driver.Build(args[0], out); err != nil {
			fail(err.Error())
		}
		fmt.Printf("built %s\n", out)

	case "version", "--version", "-v":
		fmt.Println("scar version " + version)

	case "--lsp", "lsp":
		err := lsp.StartServer()
		if err != nil {
			fmt.Printf("LSP Server Error: %v\n", err)
			os.Exit(1)
		}
		return

	case "help", "--help", "-h":
		fmt.Print(usage)

	default:
		fail("unknown command %q\n\n%s", os.Args[1], usage)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
