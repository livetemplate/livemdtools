// Command livemdtools is the CLI tool for creating and serving interactive documentation.
package main

import (
	"fmt"
	"os"

	"github.com/livetemplate/livemdtools/cmd/livemdtools/commands"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	case "serve":
		err = commands.ServeCommand(args)
	case "validate":
		err = commands.ValidateCommand(args)
	case "fix":
		err = commands.FixCommand(args)
	case "new":
		err = commands.NewCommand(args)
	case "blocks":
		err = commands.BlocksCommand(args)
	case "version":
		fmt.Printf("livemdtools version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("livemdtools - Interactive documentation made easy")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  livemdtools serve [directory]     Start development server")
	fmt.Println("  livemdtools validate [directory]  Validate markdown files")
	fmt.Println("  livemdtools fix [directory]       Auto-fix common issues")
	fmt.Println("  livemdtools blocks [directory]    Inspect code blocks")
	fmt.Println("  livemdtools new <name>            Create new tutorial")
	fmt.Println("  livemdtools version               Show version")
	fmt.Println("  livemdtools help                  Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  livemdtools serve                 # Serve current directory")
	fmt.Println("  livemdtools serve ./tutorials     # Serve tutorials directory")
	fmt.Println("  livemdtools serve --watch         # Serve with live reload")
	fmt.Println("  livemdtools validate              # Validate current directory")
	fmt.Println("  livemdtools validate examples/    # Validate specific directory")
	fmt.Println("  livemdtools fix                   # Auto-fix issues in current directory")
	fmt.Println("  livemdtools fix --dry-run         # Preview fixes without applying")
	fmt.Println("  livemdtools blocks examples/      # Inspect blocks in examples/")
	fmt.Println("  livemdtools blocks . --verbose    # Show detailed block info")
	fmt.Println("  livemdtools new my-tutorial       # Create new tutorial")
	fmt.Println()
	fmt.Println("Documentation: https://github.com/livetemplate/livemdtools")
}
