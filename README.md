# Calculator (RPN, Functional Style)

This is a simple Reverse Polish Notation (RPN) calculator implemented in Go using a functional programming paradigm.

Features:
- Pure functions for token processing and evaluation
- Higher-order `FoldTokens` function for reducing tokens over a stack
- Interactive REPL or single-expression evaluation via command-line arguments

Requirements:
- Go 1.20 or later

Usage:
1. Evaluate a single expression:
   ```sh
   go run main.go "3 4 +"
   # Output: 7
   ```
   Note: wrap expressions containing `*` or other shell-special characters in quotes.

2. Start the interactive REPL:
   ```sh
   go run main.go
   # RPN Calculator
   # Enter expression in Reverse Polish Notation, or 'exit' to quit.
   > 3 4 + 2 *
   Result: 14
   > exit
   ```

Building:
```sh
go build -o calculator
./calculator "5 1 2 + 4 * + 3 -"
# Output: 14
```