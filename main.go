package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// FoldTokens applies a processing function over a slice of elements, carrying an accumulated state.
// E is the element type; A is the accumulator (state) type.
func FoldTokens[E any, A any](elements []E, initial []A, fn func([]A, E) ([]A, error)) ([]A, error) {
	curr := initial
	for _, elem := range elements {
		var err error
		curr, err = fn(curr, elem)
		if err != nil {
			return nil, err
		}
	}
	return curr, nil
}

// Evaluate computes the result of an expression given in Reverse Polish Notation (RPN).
func Evaluate(tokens []string) (float64, error) {
	// FoldTokens[string, float64] folds string tokens into a float64 accumulator stack
	// stack, err := FoldTokens[string, float64](tokens, []float64{}, TokenProcessor)
	stack, err := FoldTokens(tokens, []float64{}, TokenProcessor)

	if err != nil {
		return 0, err
	}
	if len(stack) != 1 {
		return 0, fmt.Errorf("invalid expression: leftover stack %v", stack)
	}
	return stack[0], nil
}

// binaryOps maps operator tokens to corresponding operations.
var binaryOps = map[string]func(a, b float64) (float64, error){
	"+": func(a, b float64) (float64, error) { return a + b, nil },
	"-": func(a, b float64) (float64, error) { return a - b, nil },
	"*": func(a, b float64) (float64, error) { return a * b, nil },
	"/": func(a, b float64) (float64, error) {
		if b == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return a / b, nil
	},
}

// applyBinaryOp applies a binary operator to the top two elements of the stack.
func applyBinaryOp(stack []float64, op string) ([]float64, error) {
	fn := binaryOps[op]
	if fn == nil {
		return nil, fmt.Errorf("unknown operator '%s'", op)
	}
	if len(stack) < 2 {
		return nil, fmt.Errorf("not enough operands for '%s'", op)
	}
	a := stack[len(stack)-2]
	b := stack[len(stack)-1]
	res, err := fn(a, b)
	if err != nil {
		return nil, err
	}
	return append(stack[:len(stack)-2], res), nil
}

// TokenProcessor processes a single token, updating the evaluation stack.
// TokenProcessor processes a single token, updating the evaluation stack.
func TokenProcessor(stack []float64, tok string) ([]float64, error) {
	if _, ok := binaryOps[tok]; ok {
		return applyBinaryOp(stack, tok)
	}
	num, err := strconv.ParseFloat(tok, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid token '%s'", tok)
	}
	return append(stack, num), nil
}

// main handles CLI input: either evaluates a single expression passed as args or starts a REPL.
func main() {
	if len(os.Args) > 1 {
		expr := strings.Join(os.Args[1:], " ")
		result, err := Evaluate(strings.Fields(expr))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Println(result)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("RPN Calculator")
	fmt.Println("Enter expression in Reverse Polish Notation, or 'exit' to quit.")
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}
		tokens := strings.Fields(line)
		result, err := Evaluate(tokens)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Result:", result)
		}
	}
}
