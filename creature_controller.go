package main

import (
	"math"
	"math/rand"
	"time"
)

const (
	NotInitialized  = -1
	MaxTurnsVisible = 10
)

// init initializes the random seed.
func init() {
	rand.Seed(time.Now().UnixNano()) // Initialize the random seed
}

// MoveAll creatures based on their type and current position vx,vy and nearby creatures and drones, only perform move if creature is visible or within max turns visible
func (state *GameState) MoveAll() {
	for _, creature := range state.Creatures {
		if creature.Dead {
			continue
		}
		if creature.LastVisibleTurn != NotInitialized || creature.LastVisibleTurn+MaxTurnsVisible > state.Turn {
			creature.Move(state)
		}
	}
}

// EstimateAll position of all game creatures based on drone blips, creature type and nearby creatures.
func (state *GameState) EstimateAll() {
	for _, creature := range state.Creatures {
		state.Estimate(creature)
	}

	// Adjust positions to ensure minimum distance between each pair of fishes
	for i := 0; i < len(state.Creatures)-1; i++ {
		for j := i + 1; j < len(state.Creatures); j++ {
			fish1 := state.Creatures[i]
			fish2 := state.Creatures[j]

			if distance(fish1.X, fish1.Y, fish2.X, fish2.Y) < 500 {
				state.AdjustPositions(fish1, fish2)
			}
		}
	}
}

func (state *GameState) AdjustPositions(fish1, fish2 *Creature) {
	// Calculate the midpoint between the two fishes
	midX := (fish1.X + fish2.X) / 2
	midY := (fish1.Y + fish2.Y) / 2

	// Move each fish away from the midpoint
	fish1 = moveAway(fish1, midX, midY)
	fish2 = moveAway(fish2, midX, midY)
}

func moveAway(fish *Creature, fromX, fromY int) *Creature {
	if fish.X < fromX {
		fish.X = max(0, fish.X-250) // move left, but not beyond 0
	} else {
		fish.X = min(10000, fish.X+250) // move right, but not beyond 10000
	}

	if fish.Y < fromY {
		dimensionBoundaries := fishDepthsByType[fish.Type]
		minY, _ := dimensionBoundaries[0], dimensionBoundaries[1]
		fish.Y = max(minY, fish.Y-250) // move up, but not beyond minY
	} else {
		dimensionBoundaries := fishDepthsByType[fish.Type]
		_, maxY := dimensionBoundaries[0], dimensionBoundaries[1]
		fish.Y = min(maxY, fish.Y+250) // move down, but not beyond maxY
	}

	return fish
}

// Estimate position of creature based on drone blips, creature type and nearby creatures.
func (state *GameState) Estimate(creature *Creature) {
	if creature == nil || creature.Dead || creature.LastVisibleTurn != NotInitialized || (creature.LastVisibleTurn > NotInitialized && creature.LastVisibleTurn+MaxTurnsVisible > state.Turn) {
		return
	}

	// Get minY and maxY for the creature type
	dimensionBoundaries := fishDepthsByType[creature.Type]
	minY := dimensionBoundaries[0]
	maxY := dimensionBoundaries[1]

	// Initialize possible position ranges
	possibleXMin, possibleXMax := 0, 10000
	possibleYMin, possibleYMax := minY, maxY

	// Adjust the possible position ranges based on each drone's radar blips
	for _, drone := range state.MyDrones {
		if blip, ok := drone.RadarBlips[creature.Id]; ok {
			switch blip {
			case TopRight:
				possibleXMin = max(possibleXMin, drone.X)
				possibleYMax = min(possibleYMax, drone.Y)
			case TopLeft:
				possibleXMax = min(possibleXMax, drone.X)
				possibleYMax = min(possibleYMax, drone.Y)
			case BottomRight:
				possibleXMin = max(possibleXMin, drone.X)
				possibleYMin = max(possibleYMin, drone.Y)
			case BottomLeft:
				possibleXMax = min(possibleXMax, drone.X)
				possibleYMin = max(possibleYMin, drone.Y)
			}
		}
	}

	// Calculate the estimated position as the center of the possible range
	creature.X = (possibleXMin + possibleXMax) / 2
	creature.Y = (possibleYMin + possibleYMax) / 2

	if creature.Vx == 0 && creature.Vy == 0 {
		creature.Vx, creature.Vy = generateRandomVelocity()

	}
	// Update creature's position based on the velocity
	newX := creature.X + creature.Vx
	newY := creature.Y + creature.Vy

	if newX < 0 || newX > 10000 {
		creature.Vx = -creature.Vx // Reverse X velocity
	}
	if newY < minY || newY > maxY {
		creature.Vy = -creature.Vy // Reverse Y velocity
	}

	// Check for collision with other fishes and adjust velocity
	for _, otherFish := range state.Creatures {
		if otherFish.Dead {
			continue
		}
		if otherFish.Id != creature.Id && distance(newX, newY, otherFish.X, otherFish.Y) < 600 {
			creature.Vx = -creature.Vx // Reverse X velocity
			creature.Vy = -creature.Vy // Reverse Y velocity
			break                      // Only need to adjust once for any collision
		}
	}

	// Update position with adjusted velocity
	creature.X += creature.Vx
	creature.Y += creature.Vy

	// Ensure the fish stays within its habitat zone
	creature.X = clamp(creature.X, 0, 10000)
	creature.Y = clamp(creature.Y, minY, maxY)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func generateRandomVelocity() (int, int) {
	angle := rand.Float64() * 2 * math.Pi // Random angle in radians
	vx := int(200 * math.Cos(angle))
	vy := int(200 * math.Sin(angle))
	return vx, vy
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
