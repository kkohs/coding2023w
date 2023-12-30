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
		drone.Ascend(state)
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
		drone.Ascend(state)
		return
	}

	drone.MoveToTarget(state)
}

// Ascend function for drone to ascend to surface
func (drone *Drone) Ascend(state *GameState) {
	command := fmt.Sprintf("MOVE %d %d %d ASCENDIIING!", drone.X, 500, drone.GetLightPower(state))
	fmt.Println(command)
}

// Wait function for drone to wait
func (drone *Drone) Wait(state *GameState) {
	command := fmt.Sprintf("WAIT %d", drone.GetLightPower(state))
	fmt.Println(command)
}

// MoveTo function for drone to move to x,y
func (drone *Drone) MoveTo(state *GameState, x, y int) {
	message := "Suurface"
	if drone.Target != nil {
		message = fmt.Sprintf("Target: %d", drone.Target.Id)
	}
	command := fmt.Sprintf("MOVE %d %d %d Targeting!! %s", x, y, drone.GetLightPower(state), message)
	fmt.Println(command)
}

// MoveToTarget moves drone to target
func (drone *Drone) MoveToTarget(state *GameState) {
	if drone.Target == nil {
		drone.Wait(state)
		return
	}

	monsterInPath := drone.GetMonstersInPath(state, drone.Target.X, drone.Target.Y)
	targetX, targetY := drone.GetNextPositionTowardsTarget(drone.Target.X, drone.Target.Y)
	if len(monsterInPath) > 0 {
		Log("Monster in path", monsterInPath)
		targetX, targetY = drone.CalculateBestPathToAvoidMonsters(state, targetX, targetY)
	}

	drone.MoveTo(state, targetX, targetY)

}

func (drone *Drone) GetNextPositionTowardsTarget(targetX, targetY int) (int, int) {
	// Calculate the difference in x and y coordinates between the drone and the target
	diffX := targetX - drone.X
	diffY := targetY - drone.Y

	// Calculate the distance to the target
	distanceToTarget := math.Sqrt(float64(diffX*diffX + diffY*diffY))

	// If the distance is less than or equal to the drone's movement, the drone can reach the target in one move
	if distanceToTarget <= DroneMovement {
		return targetX, targetY
	}

	// Normalize the difference vector
	normX, normY := float64(diffX)/distanceToTarget, float64(diffY)/distanceToTarget

	// Scale the normalized vector to the drone's movement distance
	nextX := drone.X + int(normX*DroneMovement)
	nextY := drone.Y + int(normY*DroneMovement)

	return nextX, nextY
}

func (drone *Drone) CalculateBestPathToAvoidMonsters(state *GameState, targetX, targetY int) (int, int) {
	// Define angles to check for alternative paths
	angles := []int{-90, -45, 0, 45, 90}
	bestX, bestY := drone.X, drone.Y
	maxDistanceFromMonsters := 0

	for _, angle := range angles {
		// Calculate new direction with the given angle
		newVx, newVy := rotateVectorTowards(targetX-drone.X, targetY-drone.Y, angle)
		newX, newY := drone.X+newVx, drone.Y+newVy

		// Ensure the drone stays within bounds
		newX = clamp(newX, 0, 10000)
		newY = clamp(newY, 0, 10000)

		// Calculate the minimum distance to any monster from this new position
		minDist := drone.MinDistanceToAnyMonster(state, newX, newY)

		// Choose the direction that maximizes the distance to the nearest monster
		if minDist > maxDistanceFromMonsters {
			bestX, bestY = newX, newY
			maxDistanceFromMonsters = minDist
		}
	}

	return bestX, bestY
}

func (drone *Drone) GetMonstersInPath(state *GameState, targetX, targetY int) []*Creature {
	var monstersInPath []*Creature
	monsters := state.GetMonsters()

	// Calculate path increments
	diffX := float64(targetX - drone.X)
	diffY := float64(targetY - drone.Y)
	distanceToTarget := math.Sqrt(diffX*diffX + diffY*diffY)
	steps := int(distanceToTarget / DroneMovement)
	if steps == 0 {
		steps = 1
	}

	// Check each step along the path
	for i := 0; i <= steps; i++ {
		pointX := drone.X + int(float64(i)*diffX/float64(steps))
		pointY := drone.Y + int(float64(i)*diffY/float64(steps))

		for _, monster := range monsters {
			if distance(pointX, pointY, monster.X, monster.Y) <= 600 {
				monstersInPath = appendUniqueMonster(monstersInPath, monster)
			}
		}
	}

	return monstersInPath
}

// Helper function to add unique monsters to the slice
func appendUniqueMonster(monsters []*Creature, monster *Creature) []*Creature {
	for _, m := range monsters {
		if m.Id == monster.Id {
			return monsters
		}
	}
	return append(monsters, monster)
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
