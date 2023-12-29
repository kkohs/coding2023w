package main

import (
	"fmt"
	"math"
	"os"
)

// Log log message to stderr
func Log(messages ...any) {
	_, _ = fmt.Fprintln(os.Stderr, messages...)
}

// Calculate distance between to grid points, return as int
func distance(x1, y1, x2, y2 int) int {
	return int(math.Sqrt(float64((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))))
}

func normalizeVector(vx, vy int) (float64, float64) {
	mag := math.Sqrt(float64(vx*vx + vy*vy))
	return float64(vx) / mag, float64(vy) / mag
}

func rotateVector(vx, vy int, angleDeg int) (int, int) {
	angleRad := float64(angleDeg) * math.Pi / 180
	cosAngle, sinAngle := math.Cos(angleRad), math.Sin(angleRad)
	rotatedVx := float64(vx)*cosAngle - float64(vy)*sinAngle
	rotatedVy := float64(vx)*sinAngle + float64(vy)*cosAngle
	return int(rotatedVx), int(rotatedVy)
}

func dotProduct(vx1, vy1, vx2, vy2 int) int {
	return vx1*vx2 + vy1*vy2
}
