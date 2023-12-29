package main

import (
	"fmt"
	"math"
	"os"
)

// Log log message to stderr
func Log(messages ...any) {
	fmt.Fprintln(os.Stderr, messages)
}

// Calculate distance between to grid points, return as int
func distance(x1, y1, x2, y2 int) int {
	return int(math.Sqrt(float64((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))))
}
