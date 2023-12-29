package main

import (
	"fmt"
	"math"
)

const (
	MinLightDepth      = 2500
	MinBatteryLevel    = 5
	DroneMovement      = 600
	MonsterMinDistance = 1500
)

// Move moves drone to target if monster is in way tries to avoid it
func (drone *Drone) Move(state *GameState) {

	points := state.CalculatePotentialPoints()
	Log("Points", points)
	if points > 63 {
		x, y := drone.X, 500
		if drone.IsMonstersNearby(state) {
			x, y = drone.CalculateEscapePath(state)
		}
		command := fmt.Sprintf("MOVE %d %d %d ASCENDIIING! TOSCOOOORE", x, y, drone.GetLightPower(state))
		fmt.Println(command)
		return
	}

	if drone.Target != nil {
		if drone.Target.IsScanned(state) || drone.Target.IsDelivered(state) || drone.Target.IsTargeted(state, drone) {
			drone.Target = nil
		}
	}
	if drone.Target == nil {
		drone.Target = drone.FindTarget(state)
	}

	// If no target found, ascend to surface
	if drone.Target == nil {
		x, y := drone.X, 500
		if drone.IsMonstersNearby(state) {
			x, y = drone.CalculateEscapePath(state)
		}
		command := fmt.Sprintf("MOVE %d %d %d ASCENDIIING!", x, y, drone.GetLightPower(state))
		fmt.Println(command)
		return
	}

	if drone.IsMonstersNearby(state) {
		// Calculate escape path
		x, y := drone.CalculateEscapePath(state)
		command := fmt.Sprintf("MOVE %d %d %d ESCAPE!", x, y, drone.GetLightPower(state))
		fmt.Println(command)
		return
	}

	// Light
	lightPower := drone.GetLightPower(state)
	lightMessage := ""
	if lightPower == 1 {
		lightMessage = "LIGHT"
	}
	if drone.Target != nil {
		// Move to target
		x, y := drone.GetNextPoint(state)
		drone.PrevTargetDirectionX = x
		drone.PrevTargetDirectionY = y
		command := fmt.Sprintf("MOVE %d %d %d TARGETING %d %s", x, y, lightPower, drone.Target.Id, lightMessage)
		fmt.Println(command)
	} else {
		// Wait for target
		command := fmt.Sprintf("WAIT %d NO_TARGET %s", lightPower, lightMessage)
		fmt.Println(command)
	}
}

func (drone *Drone) MoveTowardsTarget(state *GameState) {
	if drone.Target == nil {
		drone.MoveTowardsSurface()
		return
	}

	nearbyMonsters := drone.GetNearbyMonsters(state, 1000)
	targetX, targetY := drone.Target.X, drone.Target.Y
	safePathFound := false

	for angle := 0; angle < 360; angle += 10 {
		pathCrossesMonster := false
		for _, monster := range nearbyMonsters {
			if pathIntersectsMonster(drone.X, drone.Y, targetX, targetY, monster) {
				pathCrossesMonster = true
				break
			}
		}

		if !pathCrossesMonster {
			safePathFound = true
			break
		}

		// Rotate the path
		rotatedVx, rotatedVy := rotateVector(targetX-drone.X, targetY-drone.Y, angle)
		targetX, targetY = drone.X+rotatedVx, drone.Y+rotatedVy
	}

	if safePathFound {
		drone.MoveTo(targetX, targetY)
	} else {
		drone.MoveTowardsSurface()
	}
}

func (drone *Drone) MoveTowardsSurface() {
	// Move towards the surface at the current x-coordinate
	fmt.Println(fmt.Sprintf("MOVE %d %d", drone.X, 500))
}

func (drone *Drone) MoveTo(x, y int) {
	fmt.Println(fmt.Sprintf("MOVE %d %d", x, y))
}
func (drone *Drone) CalculateEscapePath(state *GameState) (int, int) {
	nearestMonster, nearestDistance := drone.FindNearestMonster(state, MonsterMinDistance)
	if nearestMonster == nil {
		return drone.X, drone.Y
	}

	// Convert coordinates to float64 for precision
	escapeVx := float64(drone.X - nearestMonster.X)
	escapeVy := float64(drone.Y - nearestMonster.Y)

	// Normalize the escape vector
	normalizedVx, normalizedVy := normalizeVector(int(escapeVx), int(escapeVy))

	// Scale the normalized vector to the drone's movement range
	scaledVx := normalizedVx * float64(DroneMovement)
	scaledVy := normalizedVy * float64(DroneMovement)

	// Calculate the potential escape position
	newX := float64(drone.X) + scaledVx
	newY := float64(drone.Y) + scaledVy

	// If the nearest monster is more than 1000 units away, consider rotating the escape direction
	if nearestDistance > 1000 {
		angle := 90
		if drone.ShouldRotateLeft(state, nearestMonster) {
			angle = -90
		}

		// Rotate the escape vector
		rotatedVx, rotatedVy := rotateVector(int(scaledVx), int(scaledVy), angle)

		// Adjust the position with the rotated vector
		newX, newY = float64(drone.X)+float64(rotatedVx), float64(drone.Y)+float64(rotatedVy)

		// Check if the new position is safe from other monsters
		for _, otherMonster := range state.GetMonsters() {
			if otherMonster.Id != nearestMonster.Id && distance(int(newX), int(newY), otherMonster.X, otherMonster.Y) < MonsterMinDistance {
				newX, newY = float64(drone.X)+scaledVx, float64(drone.Y)+scaledVy
				break
			}
		}
	}

	// Convert the final coordinates back to int, ensuring they are within bounds
	return clamp(int(newX), 0, 10000), clamp(int(newY), 0, 10000)
}

func (drone *Drone) ShouldRotateLeft(state *GameState, nearestMonster *Creature) bool {
	// Temporarily rotate both ways and check which direction is safer
	leftVx, leftVy := rotateVector(drone.X-nearestMonster.X, drone.Y-nearestMonster.Y, 90)
	rightVx, rightVy := rotateVector(drone.X-nearestMonster.X, drone.Y-nearestMonster.Y, -90)

	leftDistance := drone.MinDistanceToAnyMonster(state, drone.X+leftVx, drone.Y+leftVy)
	rightDistance := drone.MinDistanceToAnyMonster(state, drone.X+rightVx, drone.Y+rightVy)

	// Rotate left if it results in a larger minimum distance to any monster, else rotate right
	return leftDistance > rightDistance
}

func (drone *Drone) MinDistanceToAnyMonster(state *GameState, x, y int) int {
	minDistance := math.MaxInt32
	for _, monster := range state.GetMonsters() {
		dist := distance(x, y, monster.X, monster.Y)
		if dist < minDistance {
			minDistance = dist
		}
	}
	return minDistance
}

func (drone *Drone) FindNearestMonster(state *GameState, radius int) (*Creature, int) {
	var nearestMonster *Creature
	nearestDistance := radius + 1 // Initialize with a value larger than the search radius

	for _, creature := range state.Creatures {
		if creature.Type == Monster {
			dist := distance(drone.X, drone.Y, creature.X, creature.Y)
			if dist < nearestDistance {
				nearestDistance = dist
				nearestMonster = creature
			}
		}
	}

	return nearestMonster, nearestDistance
}

// FindTarget Finds best target to move to, prefers deeper fish to shallow and closest to the drone while check if there are no monsters within the target path and radius of 2000
func (drone *Drone) FindTarget(state *GameState) *Creature {
	var bestTarget *Creature
	var highestDepth = -1
	var bestTargetDistance = math.MaxInt32

	for _, creature := range state.Creatures {
		if creature.Dead || creature.IsScanned(state) || creature.IsDelivered(state) || creature.IsTargeted(state, drone) {
			continue
		}

		creatureDepth := int(creature.Type)
		distanceToCreature := distance(drone.X, drone.Y, creature.X, creature.Y)

		if creatureDepth > highestDepth || (creatureDepth == highestDepth && distanceToCreature < bestTargetDistance) {
			highestDepth = creatureDepth
			bestTargetDistance = distanceToCreature
			bestTarget = creature
		}
	}

	return bestTarget
}

// GetLightPower returns  1 if light is to be allowed or 0 if not,
// Light can be used if drone is below 2500 units from the surface energy is more than 5 and at least 2 turns have passed
func (drone *Drone) GetLightPower(state *GameState) int {
	if drone.Y > MinLightDepth && drone.Battery > MinBatteryLevel && drone.LastLightTurn+2 < state.Turn && !drone.IsMonstersNearby(state) {
		drone.LastLightTurn = state.Turn
		return 1
	}
	return 0
}

// GetNextPoint calcualtes next x,y point of drone to move to target, movement is 600 units
func (drone *Drone) GetNextPoint(state *GameState) (int, int) {
	if drone.Target == nil {
		Log("No target available")
		return drone.X, drone.Y
	}

	// Calculate the difference in position using float64 for precision
	diffX := float64(drone.Target.X - drone.X)
	diffY := float64(drone.Target.Y - drone.Y)

	// Calculate the distance to the target
	distanceToTarget := math.Sqrt(diffX*diffX + diffY*diffY)

	// Normalize the difference
	normX, normY := diffX/distanceToTarget, diffY/distanceToTarget

	// Calculate move distance, which is the minimum of the distance to the target or the drone's movement range
	moveDistance := math.Min(distanceToTarget, float64(DroneMovement))

	// Calculate the next position and convert it back to int
	nextX := drone.X + int(normX*moveDistance)
	nextY := drone.Y + int(normY*moveDistance)

	Log("Moving towards target:", drone.Target.Id, "Next position:", nextX, nextY)
	return nextX, nextY
}

// IsMonstersNearby returns if at least 1 monster is within 1000 units of drone
func (drone *Drone) IsMonstersNearby(state *GameState) bool {
	for _, creature := range state.Creatures {
		if creature.Type == Monster && distance(drone.X, drone.Y, creature.X, creature.Y) < MonsterMinDistance {
			return true
		}
	}
	return false
}

func (drone *Drone) GetNearbyMonsters(state *GameState, radius int) []*Creature {
	var nearbyMonsters []*Creature
	for _, creature := range state.Creatures {
		if creature.Type == Monster && distance(drone.X, drone.Y, creature.X, creature.Y) < radius {
			nearbyMonsters = append(nearbyMonsters, creature)
		}
	}
	return nearbyMonsters
}
