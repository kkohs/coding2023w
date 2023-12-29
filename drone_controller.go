package main

import (
	"fmt"
	"math"
)

// Move moves drone to target if monster is in way tries to avoid it
func (drone *Drone) Move(state *GameState) {
	if drone.Target != nil {
		if distance(drone.X, drone.Y, drone.Target.X, drone.Target.Y) < 100 {
			drone.Target = nil
		}
	}
	if drone.Target == nil {
		drone.Target = drone.FindTarget(state)
	}
	if drone.Target != nil {
		fmt.Println("MOVE " + fmt.Sprint(drone.Target.X) + " " + fmt.Sprint(drone.Target.Y) + " " + fmt.Sprint(0) + " TARGETING " + fmt.Sprint(drone.Target.Id))
	} else {
		// Wait for target
		fmt.Println("WAIT 0")
	}
}

// FindTarget Finds best target to move to, prefers deeper fish over shallow and closest to the drone while check if there are no monsters within the target path and radius of 2000
func (drone *Drone) FindTarget(state *GameState) *Creature {
	var bestTarget *Creature
	var bestPriorityScore int = -1
	var bestTargetDistance int = math.MaxInt32

	for _, creature := range state.Creatures {
		if creature.Dead || creature.IsScanned(state) || creature.IsDelivered(state) {
			continue // Ignore dead, scanned, or delivered creatures
		}

		priorityScore := getPriorityScore(creature)
		distanceToCreature := distance(drone.X, drone.Y, creature.X, creature.Y)

		// Higher priority score or closer distance with the same score
		if priorityScore > bestPriorityScore ||
			(priorityScore == bestPriorityScore && distanceToCreature < bestTargetDistance) {
			if !isPathDangerous(drone, creature, state) {
				bestTarget = creature
				bestPriorityScore = priorityScore
				bestTargetDistance = distanceToCreature
			}
		}
	}

	return bestTarget
}

func getPriorityScore(creature *Creature) int {
	return int(creature.Type)
}
func isPathDangerous(drone *Drone, creature *Creature, state *GameState) bool {
	for _, monster := range state.GetMonsters() {
		if distanceToPath(drone.X, drone.Y, creature.X, creature.Y, monster.X, monster.Y) < 1000 {
			return true // Monster is too close to the path
		}
	}
	return false
}

func distanceToPath(x1, y1, x2, y2, px, py int) int {
	lineLength := distance(x1, y1, x2, y2)
	if lineLength == 0 {
		return distance(px, py, x1, y1)
	}

	ratio := ((px-x1)*(x2-x1) + (py-y1)*(y2-y1)) / (lineLength * lineLength)

	if ratio < 0 {
		return distance(px, py, x1, y1) // Closest to the first endpoint
	} else if ratio > 1 {
		return distance(px, py, x2, y2) // Closest to the second endpoint
	} else {
		projX := x1 + ratio*(x2-x1)
		projY := y1 + ratio*(y2-y1)
		return distance(px, py, int(projX), int(projY))
	}
}
