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
- **Built-ins**: Currently supports `cd`, `exit`, and `source` (reads and executes commands from a file).
- **Profile Sourcing**: Automatically looks for and executes `~/.ishallrc` on startup.
- **Subshells**: Enclosing commands in parentheses `( )` launches them in an isolated child process, enabling isolated execution environments that plug seamlessly into pipelines.
- **Dynamic Colored Prompt**: Displays a colorful `user@device@ishall:filepath>` prompt using ANSI escape codes.
- **Tab Completion & History**: Supports auto-completing files in the current directory with `TAB`, saving command history across sessions, and full arrow-key cursor navigation.

## What it lacks

`i-shall` is an educational project and lacks many quality-of-life and advanced features found in daily-driver shells like Bash, Zsh, or Fish. Currently, it **does not** support:

- **Variables & Environment Manipulation**: No support for `$VAR` expansion, `export`, or environment variables.
- **Background Tasks & Job Control**: No support for running commands in the background (`&`), nor handling `Ctrl+Z` to suspend jobs (`fg`/`bg`).
- **Advanced Control Flow**: No `if`, `while`, or `for` loops.
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
