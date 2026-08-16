# i-shall User Manual

Welcome to `i-shall`, a lightweight, modern UNIX shell designed with clean, intuitive syntax. This guide will walk you through setting up and using the shell.

## 1. Getting Started

### Installation & Execution
Ensure you have Go installed on your system.
```bash
git clone https://github.com/Octahedron-apple/i-shall.git
cd i-shall
go run .
```

### Profile Sourcing
Upon startup, `i-shall` automatically looks for a configuration file located at `~/.ishallrc`. You can put your custom aliases, exported variables, and functions in this file, and they will load every time you open the shell!

## 2. Shell Basics

`i-shall` supports the core functionality you expect from a daily driver:
- **Command Execution**: Run standard UNIX binaries (`ls`, `grep`, `cat`).
- **Pipelining**: Chain commands together using `|`.
- **I/O Redirection**: Write outputs using `>`, append using `>>`, and read inputs using `<`.
- **Logical Operators**: Chain execution using `&&` and `||`.
- **Subshells**: Run isolated commands using `( command )`.

## 3. The `i-shall` Grammar

Where `i-shall` shines is its custom, strict, and explicit typing grammar.

### Variables & Data Structures
You define variables cleanly with `=` (spaces around it are allowed!).
```bash
name = "Alice"
age = 30
arr = ("apples", "bananas")
```

Unlike traditional shells, `i-shall` explicitly types arguments using variable prefixes when referencing them:
- **Strings (`$`)**: Retrieve a variable as a string (e.g., `$name`).
- **Numbers (`#`)**: Retrieve a variable as a number (e.g., `#age`).
- **Arrays**: Access array elements using brackets (e.g., `$arr[1]` returns `"bananas"`).

### Exporting Variables
To pass a variable to child processes (like Python or Node), simply add the `export` keyword before assignment:
```bash
export API_KEY = "12345"
```

## 4. Control Flow (The "Better Syntax")

`i-shall` abandons archaic Bash keywords like `then` and `do`. Blocks are inferred natively via semicolons `;` (or line-breaks in scripts).

### If/Else Statements
```bash
if #age > 18 ;
    echo "Adult"
elif #age > 12 ;
    echo "Teen"
else
    echo "Child"
fi
```

### While Loops
```bash
count = 0
while #count < 5 ;
    echo "Count:" $count
    count = #count + 1
done
```

### C-Style For Loops
`i-shall` features a built-in mathematical evaluator for loops!
```bash
for ( #i = 0 ; #i < 5 ; #i = #i + 1 ) ;
    echo "Loop index:" $i
done
```

## 5. Functions & Aliases

### Functions
You can group commands into functions and access arguments using the `$args` array.
```bash
fn greet ;
    echo "Hello" $args[0]
end

greet "Alice"
# Output: Hello Alice
```

### Aliases
Under the hood, aliases are just functions! They automatically pass arguments forward.
```bash
alias ll = "ls -la"
ll /tmp
# Executes: ls -la /tmp
```

## 6. Advanced Features

`i-shall` implements powerful command expansions and substitution natively!

### Command Substitution `$(...)`
You can execute a command and capture its standard output directly into a variable or as an argument to another command.
```bash
cwd_files = $(ls)
echo "Current files are:" $cwd_files
```

### Variable Interpolation & Quotes
Single quotes (`'...'`) treat everything as literal text. Double quotes (`"..."`) allow you to seamlessly inject variables into your strings!
```bash
name = "World"
echo "Hello $name! 1 + 1 is #two"
```

### Path & Brace Expansions
- **Tilde (`~`)**: Instantly expands to your home directory (`cd ~` or `ls ~/Documents`).
- **Brace Expansion (`{}`)**: Rapidly generate string permutations.
```bash
echo file_{1,2,3}.txt
# Output: file_1.txt file_2.txt file_3.txt
```

### Trapping Signals
You can catch OS signals (like `SIGINT` from pressing Ctrl+C) and run a command instead of crashing your script.
```bash
trap "echo 'You pressed Ctrl+C!'" SIGINT
```

## 7. Interactive Shell (REPL)
If you are typing interactively and leave an `if` block, a loop, or a quote `"` open when pressing Enter, `i-shall` will automatically buffer your input and change your prompt to `> `, allowing you to seamlessly type multi-line code directly in the terminal!
