package main

import (
	"strings"
	"testing"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		expr string
		want float64
	}{
		{"3 4 +", 7},
		{"10 2 /", 5},
		{"2 3 4 * +", 14},
	}
	for _, tt := range tests {
		got, err := Evaluate(strings.Fields(tt.expr))
		if err != nil {
			t.Fatalf("%s: %v", tt.expr, err)
		}
		if got != tt.want {
			t.Errorf("%s: want %v got %v", tt.expr, tt.want, got)
		}
	}
}
