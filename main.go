package main

import (
   "bufio"
   "fmt"
   "os"
   "strings"
   "strconv"
)

// FoldTokens applies a processing function over a slice of tokens, carrying an accumulated stack.
func FoldTokens(tokens []string, initial []float64, fn func([]float64, string) ([]float64, error)) ([]float64, error) {
   curr := initial
   for _, tok := range tokens {
       var err error
       curr, err = fn(curr, tok)
       if err != nil {
           return nil, err
       }
   }
   return curr, nil
}

// Evaluate computes the result of an expression given in Reverse Polish Notation (RPN).
func Evaluate(tokens []string) (float64, error) {
   stack, err := FoldTokens(tokens, []float64{}, TokenProcessor)
   if err != nil {
       return 0, err
   }
   if len(stack) != 1 {
       return 0, fmt.Errorf("invalid expression: leftover stack %v", stack)
   }
   return stack[0], nil
}

// TokenProcessor processes a single token, updating the evaluation stack.
func TokenProcessor(stack []float64, tok string) ([]float64, error) {
   switch tok {
   case "+":
       if len(stack) < 2 {
           return nil, fmt.Errorf("not enough operands for '+'")
       }
       b, a := stack[len(stack)-1], stack[len(stack)-2]
       return append(stack[:len(stack)-2], a+b), nil
   case "-":
       if len(stack) < 2 {
           return nil, fmt.Errorf("not enough operands for '-' ")
       }
       b, a := stack[len(stack)-1], stack[len(stack)-2]
       return append(stack[:len(stack)-2], a-b), nil
   case "*":
       if len(stack) < 2 {
           return nil, fmt.Errorf("not enough operands for '*' ")
       }
       b, a := stack[len(stack)-1], stack[len(stack)-2]
       return append(stack[:len(stack)-2], a*b), nil
   case "/":
       if len(stack) < 2 {
           return nil, fmt.Errorf("not enough operands for '/' ")
       }
       b, a := stack[len(stack)-1], stack[len(stack)-2]
       if b == 0 {
           return nil, fmt.Errorf("division by zero")
       }
       return append(stack[:len(stack)-2], a/b), nil
   default:
       num, err := strconv.ParseFloat(tok, 64)
       if err != nil {
           return nil, fmt.Errorf("invalid token '%s'", tok)
       }
       return append(stack, num), nil
   }
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