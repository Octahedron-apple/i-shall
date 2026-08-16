# i-shall

![Go Tests](https://github.com/Octahedron-apple/i-shall/actions/workflows/test.yml/badge.svg)

`i-shall` is a lightweight, custom-built UNIX shell written in Go. It implements a fully functioning Lexer, Parser, and Abstract Syntax Tree (AST) architecture to handle modern shell grammars.

## Features

While minimal, `i-shall` currently supports a core set of standard POSIX shell features:

- **Command Execution**: Standard binary execution.
- **Pipelining**: Chain processes together using `|` (e.g., `ls -la | grep main`).
- **I/O Redirection**: Supports writing (`>`), appending (`>>`), and reading (`<`) standard input/output to and from files.
- **Logical Operators**: Conditional chaining of commands using `&&` (AND) and `||` (OR) (e.g., `build && run || echo "Failed"`).
- **Quote Parsing**: Correctly groups strings with spaces using single (`'`) and double (`"`) quotes.
- **Built-ins**: Currently supports `cd` (changes directory) and `exit`.
- **Dynamic Colored Prompt**: Displays a colorful `user@device@ishall:filepath>` prompt using ANSI escape codes.

## What it lacks

`i-shall` is an educational project and lacks many quality-of-life and advanced features found in daily-driver shells like Bash, Zsh, or Fish. Currently, it **does not** support:

- **Variables & Environment Manipulation**: No support for `$VAR` expansion, `export`, or `.bashrc`/profile sourcing.
- **Tab Completion & History**: Pressing the up-arrow or tab key will not work, as it currently reads raw bytes via `bufio` rather than utilizing a readline library.
- **Background Tasks & Job Control**: No support for running commands in the background (`&`), nor handling `Ctrl+Z` to suspend jobs (`fg`/`bg`).
- **Globbing/Wildcards**: Using `*` or `?` (e.g., `rm *.txt`) will not expand to file lists. It will be passed as a literal `*` to the command.
- **Advanced Control Flow**: No `if`, `while`, `for`, or subshells `()`.
- **Aliases**: No custom command aliases.

## Architecture & Testing

`i-shall` relies on a multi-stage architecture:
1. **Lexer**: Converts raw input strings into a sequence of Tokens.
2. **Parser**: Consumes tokens to build an Abstract Syntax Tree (specifically, a `Sequence` of `Pipelines` containing `Commands`).
3. **Executor**: Walks the AST, setting up OS file descriptors (`io.Pipe`, `os.File`) and conditionally executing `os/exec` processes based on exit codes.

**Automated Testing**: This repository has automated unit tests for both the Lexer and Parser components located in the `/tests` directory, which are hooked up to GitHub Actions.

## Getting Started

To run `i-shall`, ensure you have Go installed:

```bash
# Clone the repository
git clone https://github.com/Octahedron-apple/i-shall.git
cd i-shall

# Run the tests
go test -v ./tests

# Build and run the shell
go run .
```
