package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"
)

type GameState struct {
	MyScore      int
	FoeScore     int
	MyScans      []int
	FoeScans     []int
	MyDrones     []*Drone
	FoeDrones    []*Drone
	Scans        map[int][]Scan
	Creatures    []*Creature
	MySavedScans []int
	FoeSavedScan []int
	Initialized  bool
}

type CreatureType int

const (
	Monster     CreatureType = -1
	ShallowFish CreatureType = 0
	MediumFish  CreatureType = 1
	DeepFish    CreatureType = 2
)

const MaxDepth = 9000
const FieldWidth = 10000

const (
	ScoreShallowFish = 1
	ScoreMediumFish  = 2
	ScoreDeepFish    = 3
)

var fishHbitat = map[CreatureType][]int{
	Monster:     []int{2500, 10000},
	ShallowFish: []int{2500, 5000},
	MediumFish:  []int{5000, 7500},
	DeepFish:    []int{7500, 10000},
}

const FirstToSaveModifier = 2

type Creature struct {
	Id         int
	X          int
	Y          int
	Vx         int
	Vy         int
	Type       CreatureType
	RadarBlips map[int]RadarBlip
	Dead       bool
	Visible    bool
	Color      int
}

type Drone struct {
	Id               int
	Index            int
	X                int
	Y                int
	Emergency        int
	Battery          int
	TargetCreature   *Creature
	RadarBlips       map[int]RadarBlip
	DirectionChanged bool
	LastMoveX        int
	LastMoveY        int
	Ascending        bool
	LastLight        int
	ZoneMinX         int
	ZoneMaxX         int
}

type RadarBlip string

const (
	TopLeft     RadarBlip = "TL"
	TopRight    RadarBlip = "TR"
	BottomLeft  RadarBlip = "BL"
	BottomRight RadarBlip = "BR"
)

type Scan struct {
	DroneId    int
	CreatureId int
}

type Point struct {
	X, Y float64
}

func main() {
	state := GameState{}

	var creatureCount int
	fmt.Scan(&creatureCount)
	state.Creatures = make([]*Creature, creatureCount)

	for i := 0; i < creatureCount; i++ {
		var creatureId, color, _type int
		fmt.Scan(&creatureId, &color, &_type)
		state.Creatures[i] = &Creature{Id: creatureId, Type: CreatureType(_type), Color: color}
	}

	// sort creatures by id descending
	for i := 0; i < len(state.Creatures); i++ {
		for j := 0; j < len(state.Creatures)-1; j++ {
			if state.Creatures[j].Id < state.Creatures[j+1].Id {
				state.Creatures[j], state.Creatures[j+1] = state.Creatures[j+1], state.Creatures[j]
			}
		}
	}
	for {
		state.resetCreatureVisibility()
		var myScore int
		fmt.Scan(&myScore)
		state.MyScore = myScore

		var foeScore int
		fmt.Scan(&foeScore)
		state.FoeScore = foeScore

		var myScanCount int
		fmt.Scan(&myScanCount)
		state.MyScans = make([]int, myScanCount)

		for i := 0; i < myScanCount; i++ {
			var creatureId int
			fmt.Scan(&creatureId)
			state.MyScans[i] = creatureId
		}
		var foeScanCount int
		fmt.Scan(&foeScanCount)
		state.FoeScans = make([]int, foeScanCount)

		for i := 0; i < foeScanCount; i++ {
			var creatureId int
			fmt.Scan(&creatureId)
			state.FoeScans[i] = creatureId
		}
		var myDroneCount int
		index := 0
		fmt.Scan(&myDroneCount)
		if state.MyDrones == nil {
			state.MyDrones = make([]*Drone, myDroneCount)
		}

		for i := 0; i < myDroneCount; i++ {
			var droneId, droneX, droneY, emergency, battery int
			fmt.Scan(&droneId, &droneX, &droneY, &emergency, &battery)

			if state.MyDrones[i] == nil {
				state.MyDrones[i] = &Drone{Index: index}
				index++
			}
			state.MyDrones[i].Id = droneId
			state.MyDrones[i].X = droneX
			state.MyDrones[i].Y = droneY
			state.MyDrones[i].Emergency = emergency
			state.MyDrones[i].Battery = battery
		}

		sort.Slice(state.MyDrones, func(i, j int) bool {
			return state.MyDrones[i].X < state.MyDrones[j].X
		})

		zoneWidth := FieldWidth / myDroneCount
		for i, drone := range state.MyDrones {
			if drone.ZoneMinX != 0 || drone.ZoneMaxX != 0 {
				continue
			}
			drone.ZoneMinX = i * zoneWidth
			drone.ZoneMaxX = (i + 1) * zoneWidth
		}

		sort.Slice(state.MyDrones, func(i, j int) bool {
			return state.MyDrones[i].Id < state.MyDrones[j].Id
		})

		var foeDroneCount int
		fmt.Scan(&foeDroneCount)

		if state.FoeDrones == nil {
			state.FoeDrones = make([]*Drone, foeDroneCount)
		}
		for i := 0; i < foeDroneCount; i++ {
			var droneId, droneX, droneY, emergency, battery int
			fmt.Scan(&droneId, &droneX, &droneY, &emergency, &battery)
			if state.FoeDrones[i] == nil {
				state.FoeDrones[i] = &Drone{}
			}
			state.FoeDrones[i].Id = droneId
			state.FoeDrones[i].X = droneX
			state.FoeDrones[i].Y = droneY
			state.FoeDrones[i].Emergency = emergency
			state.FoeDrones[i].Battery = battery
		}
		var droneScanCount int
		fmt.Scan(&droneScanCount)
		state.Scans = make(map[int][]Scan)
		for i := 0; i < droneScanCount; i++ {
			var droneId, creatureId int
			fmt.Scan(&droneId, &creatureId)
			if state.Scans[droneId] == nil {
				state.Scans[droneId] = make([]Scan, 0)
			}
			state.Scans[droneId] = append(state.Scans[droneId], Scan{DroneId: droneId, CreatureId: creatureId})
		}
		var visibleCreatureCount int
		fmt.Scan(&visibleCreatureCount)

		for i := 0; i < visibleCreatureCount; i++ {
			var creatureId, creatureX, creatureY, creatureVx, creatureVy int
			fmt.Scan(&creatureId, &creatureX, &creatureY, &creatureVx, &creatureVy)
			for j := 0; j < len(state.Creatures); j++ {
				if state.Creatures[j].Id == creatureId {
					state.Creatures[j].X = creatureX
					state.Creatures[j].Y = creatureY
					state.Creatures[j].Vx = creatureVx
					state.Creatures[j].Vy = creatureVy
					state.Creatures[j].Visible = true
				}
			}
		}
		var radarBlipCount int
		fmt.Scan(&radarBlipCount)
		for c := range state.Creatures {
			state.Creatures[c].RadarBlips = make(map[int]RadarBlip)
		}
		for d := range state.MyDrones {
			state.MyDrones[d].RadarBlips = make(map[int]RadarBlip)
		}

		for i := 0; i < radarBlipCount; i++ {
			var droneId, creatureId int
			var radar string
			fmt.Scan(&droneId, &creatureId, &radar)
			for c := range state.Creatures {
				if state.Creatures[c].Id == creatureId {
					state.Creatures[c].RadarBlips[droneId] = RadarBlip(radar)
				}
			}
			for i2 := range state.MyDrones {
				if state.MyDrones[i2].Id == droneId {
					state.MyDrones[i2].RadarBlips[creatureId] = RadarBlip(radar)
				}
			}
		}

		state.updateDeadCreatures()
		state.updateCreaturesPosition()
		state.printMonsters()
		state.setDroneZones()
		state.approximateFishPositions()

		if !state.Initialized {
			state.Initialized = true
		}
		for i := 0; i < myDroneCount; i++ {
			var message string
			drone := state.MyDrones[i]

			light := 0
			if drone.Y > 3000 && drone.Battery > 5 && drone.LastLight+1000 < drone.Y {
				light = 1
				drone.LastLight = drone.Y
				message = fmt.Sprintf("You herer mate? !")
			}

			if drone.Battery < 5 {
				drone.LastLight = 0
			}

			if drone.TargetCreature == nil || !drone.isWithinZone(drone.TargetCreature) {
				drone.TargetCreature = state.getMostValuableCreature(drone)
			}

			if drone.isDroneReadyToAscend(state) {
				log("Drone is ready to ascend")
				drone.Ascending = true
			}

			if drone.Ascending {
				if drone.Y <= 500 {
					log("Drone reached surface")
					drone.Ascending = false
				} else {
					targetX, targetY := drone.ascend(600, state)
					fmt.Println("MOVE", targetX, targetY, light, message)
					continue
				}
			}

			if drone.TargetCreature != nil {

				targetX, targetY := -1, -1

				if state.isCreatureScanned(drone.TargetCreature.Id, drone.Id) {
					//	log("Creature already scanned ", drone.TargetCreature.Id)
					drone.TargetCreature = state.getMostValuableCreature(drone)
					//	log("New target for drone", drone.TargetCreature.Id, drone.Id)
				} else if state.isCreatureSaved(drone.TargetCreature.Id) {
					log("Creature already saved", drone.TargetCreature.Id)
					drone.TargetCreature = state.getMostValuableCreature(drone)
					//log("New target, for drone", drone.TargetCreature.Id, drone.Id)
				}

				if drone.TargetCreature == nil {
					drone.Ascending = true
					targetX, targetY := drone.ascend(600, state)
					fmt.Println("MOVE", targetX, targetY, light, message)
					continue
				}

				targetX, targetY = drone.determineTargetPosition45(drone.RadarBlips[drone.TargetCreature.Id], 600, state)
				message = fmt.Sprintf("I'm coming for you %d!", drone.TargetCreature.Id)

				if targetX == -1 || targetY == -1 {
					log("Target position not found")
					fmt.Println("WAIT", light)
					continue
				}

				fmt.Println("MOVE", targetX, targetY, light, message)
			} else {
				log("No target found")
				fmt.Println("WAIT", 0)
			}
		}
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (drone *Drone) avoidMonsters(offset int, state GameState) (int, int) {
	var totalDx, totalDy int

	// Aggregate the influence of all nearby monsters
	for creatureId, _ := range drone.RadarBlips {
		creature := state.getCreatureById(creatureId)
		if creature != nil && creature.isMonster() && drone.isCreatureWithinRange(creature) {
			dx := drone.X - creature.X
			dy := drone.Y - creature.Y
			totalDx += dx
			totalDy += dy
		}
	}

	newX, newY := drone.X, drone.Y

	if totalDx != 0 || totalDy != 0 {
		// Determine the direction to move away from the aggregated monster influence
		newX += sign(totalDx) * offset
		newY += sign(totalDy) * offset
	} else {
		// Default movement if no monsters are close
		newX += offset
		newY += offset
	}

	// Apply boundary check and adjust if necessary
	newX, newY = adjustForBoundaries(newX, newY, offset)
	return newX, newY
}

// check if drone is ready to ascend, drone has scanned more than 4 creatures or depth is below 8500 or has at least 1 complete set, 3 of color or type
func (drone *Drone) isDroneReadyToAscend(state GameState) bool {
	if len(drone.getUniqueScans(state)) >= 5 {
		log("Drone has scanned more than 4 creatures", drone.getUniqueScans(state))
		return true
	}
	if drone.Y >= MaxDepth {
		log("Drone is below 8500")
		return true
	}

	return false
}

// Get unique scans from drone
func (drone *Drone) getUniqueScans(state GameState) []Scan {
	var uniqueScans []Scan
	for _, scan := range state.Scans[drone.Id] {
		for i := range state.MyDrones {
			if state.MyDrones[i].Id == drone.Id {
				continue
			}

			if !state.isCreatureScanned(scan.CreatureId, state.MyDrones[i].Id) {
				uniqueScans = append(uniqueScans, scan)
			}
		}
	}
	return uniqueScans
}

// check if drone has complete set, 3 of color or type
func (drone *Drone) hasCompleteSet(state GameState) bool {
	colorMap := make(map[int]int)
	typeMap := make(map[CreatureType]int)
	for _, scan := range state.Scans[drone.Id] {
		creature := state.getCreatureById(scan.CreatureId)
		if creature == nil {
			continue
		}
		colorMap[creature.Color]++
		typeMap[creature.Type]++
	}
	for _, count := range colorMap {
		if count >= 3 {
			log("Drone has complete set of color")
			return true
		}
	}
	for _, count := range typeMap {
		if count >= 4 {
			log("Drone has complete set of type")
			return true
		}
	}
	return false
}

// Method to check if a creature is within the drone's zone
func (drone *Drone) isWithinZone(creature *Creature) bool {
	return creature.X >= drone.ZoneMinX && creature.X <= drone.ZoneMaxX
}

// Method to calculate distance to a creature
func (drone *Drone) distanceToCreature(creature *Creature) float64 {
	return distance(drone.X, drone.Y, creature.X, creature.Y)
}

// Check if drone is withing creatures habitat
func (drone *Drone) isDroneWithinCreatureHabitat(creature *Creature) bool {

	habitat := fishHbitat[creature.Type]
	return drone.Y >= habitat[0] && drone.Y <= habitat[1]
}

// adjustForBoundaries ensures the drone's new position is within the map boundaries
func adjustForBoundaries(x, y, offset int) (int, int) {
	if x < 0 {
		x = offset
	} else if x > 10000 { // Assuming 10000 is the map boundary
		x = 10000 - offset
	}

	if y < 0 {
		y = offset
	} else if y > 10000 { // Assuming 10000 is the map boundary
		y = 10000 - offset
	}

	return x, y
}

// determine surface position while avoiding monsters
func (drone *Drone) ascend(offset int, state GameState) (int, int) {
	// Check for visible monsters and adjust direction if needed
	for creatureId, _ := range drone.RadarBlips {
		creature := state.getCreatureById(creatureId)
		if creature != nil && creature.isMonster() && creature.Visible && drone.isCreatureWithinRange(creature) {
			return drone.avoidMonsters(offset, state)
		}
	}

	// Original logic to move towards target fish
	return drone.X, drone.Y - offset
}

func (drone *Drone) determineTargetPosition45(blip RadarBlip, offset int, state GameState) (int, int) {
	// Check for visible monsters and adjust direction if needed
	for creatureId, _ := range drone.RadarBlips {
		creature := state.getCreatureById(creatureId)
		if creature != nil && creature.isMonster() && drone.isCreatureWithinRange(creature) {
			return drone.avoidMonsters(offset, state)
		}
	}

	// Adjust movement based on radar blip
	return drone.moveTowardsTarget(blip, offset)
}

// Calculates the target position based on the radar blip
func (drone *Drone) moveTowardsTarget(blip RadarBlip, offset int) (int, int) {
	var targetX, targetY int

	switch blip {
	case TopLeft:
		targetX, targetY = drone.X-offset, drone.Y-offset/2
	case TopRight:
		targetX, targetY = drone.X+offset, drone.Y-offset/2
	case BottomLeft:
		targetX, targetY = drone.X-offset, drone.Y+offset/2
	case BottomRight:
		targetX, targetY = drone.X+offset, drone.Y+offset/2
	default:
		targetX, targetY = drone.X, drone.Y-offset // Default case to ascend
	}

	// Apply boundary check and adjust if necessary
	targetX, targetY = adjustForBoundaries(targetX, targetY, offset)
	return targetX, targetY
}

// check if creature is within 2000 units of drone
func (drone *Drone) isCreatureWithinRange(creature *Creature) bool {
	return distance(drone.X, drone.Y, creature.X, creature.Y) <= 1600
}

// print scans and saves
func (state *GameState) setDroneZones() {
	if state.Initialized {
		return
	}

	middlePoint := FieldWidth / 2

	for _, drone := range state.MyDrones {
		if drone.X < middlePoint {
			// Drone is closer to the left side
			drone.ZoneMinX = 0
			drone.ZoneMaxX = middlePoint + 1000
		} else {
			// Drone is closer to the right side
			drone.ZoneMinX = middlePoint - 1000
			drone.ZoneMaxX = FieldWidth
		}
		log(fmt.Sprintf("Drone %d: ZoneMinX = %d, ZoneMaxX = %d", drone.Id, drone.ZoneMinX, drone.ZoneMaxX))
	}
}

// Gets target, target is selected 1. Take shallow fish first if all shallow fish scanned or dead move to medium fish and so on
func (state *GameState) getMostValuableCreature(drone *Drone) *Creature {
	log("Getting most valuable creature for drone", drone.Id)
	shallowFishCount, mediumFishCount, deepFishCount := 0, 0, 0
	for c := range state.Creatures {
		if state.Creatures[c].isMonster() || state.isCreatureSaved(state.Creatures[c].Id) || state.Creatures[c].Dead || state.isCreatureTargeted(state.Creatures[c].Id) || state.isCreatureScannedByMyDrones(state.Creatures[c].Id) {
			continue
		}
		switch state.Creatures[c].Type {
		case ShallowFish:
			shallowFishCount++
		case MediumFish:
			mediumFishCount++
		case DeepFish:
			deepFishCount++
		}
	}
	// Check each fish type starting from the highest value
	log("Shallow fish count", shallowFishCount, "Medium fish count", mediumFishCount, "Deep fish count", deepFishCount)
	if deepFishCount > 0 {
		return state.getMostValuableCreatureByTypeAndZone(DeepFish, drone)
	} else if mediumFishCount > 0 {
		return state.getMostValuableCreatureByTypeAndZone(MediumFish, drone)
	} else if shallowFishCount > 0 {
		return state.getMostValuableCreatureByTypeAndZone(ShallowFish, drone)
	}
	log("No fish found")
	return nil
}

// New method to get the most valuable creature by type within a drone's zone
func (state *GameState) getMostValuableCreatureByTypeAndZone(creatureType CreatureType, drone *Drone) *Creature {
	var closestCreature *Creature
	closestDistance := math.MaxFloat64

	for _, creature := range state.Creatures {
		if creature.Type != creatureType || creature.Dead || state.isCreatureSaved(creature.Id) || state.isCreatureTargeted(creature.Id) {
			log("Creature", creature.Id, "is not of type", creatureType, "is dead", creature.Dead, "is saved", state.isCreatureSaved(creature.Id), "is targeted", state.isCreatureTargeted(creature.Id))
			continue
		}
		if drone.isWithinZone(creature) {
			log("Creature", creature.Id, "is within drone zone", drone.Id)
			distance := drone.distanceToCreature(creature)
			log("Distance to creature", creature.Id, "is", distance)
			if distance < closestDistance {
				closestDistance = distance
				closestCreature = creature
				log("Closest creature", closestCreature.Id, "Distance", closestDistance)
			}
		} else {
			log("Creature", creature.Id, "is not within drone zone", drone.Id)
		}
	}
	return closestCreature
}

// check if creature is targeted by any of my drones
func (state *GameState) isCreatureTargeted(creatureId int) bool {
	for _, drone := range state.MyDrones {
		if drone.TargetCreature != nil && drone.TargetCreature.Id == creatureId {
			return true
		}
	}
	return false
}

// Get most valuable creature by type
func (state *GameState) getMostValuableCreatureByType(creatureType CreatureType, drone *Drone) *Creature {
	var creature *Creature
	startIndex := 0
	endIndex := len(state.Creatures) - 1
	step := 1

	// If drone index is odd, start from the end
	if drone.Index%2 != 0 {
		startIndex = endIndex
		endIndex = 0
		step = -1
	}

	log("Drone index", drone.Index, "Start index", startIndex, "End index", endIndex, "Step", step)

	for c := startIndex; c >= 0 && c <= endIndex; c += step {
		if state.Creatures[c].isMonster() ||
			state.isCreatureSaved(state.Creatures[c].Id) ||
			state.Creatures[c].Dead ||
			state.Creatures[c].Type != creatureType ||
			state.isCreatureTargeted(state.Creatures[c].Id) || state.isCreatureScannedByMyDrones(state.Creatures[c].Id) {
			log("Creature", state.Creatures[c].Id, "is monster", state.Creatures[c].isMonster(), "is saved", state.isCreatureSaved(state.Creatures[c].Id), "is dead", state.Creatures[c].Dead, "is targeted", state.isCreatureTargeted(state.Creatures[c].Id), "is scanned by my drones", state.isCreatureScannedByMyDrones(state.Creatures[c].Id))

			continue
		}
		if creature == nil {
			creature = state.Creatures[c]
			continue
		}
		if creatureType == ShallowFish && state.Creatures[c].Y > creature.Y {
			creature = state.Creatures[c]
		} else if creatureType == MediumFish && state.Creatures[c].Y > creature.Y {
			creature = state.Creatures[c]
		} else if creatureType == DeepFish && state.Creatures[c].Y > creature.Y {
			creature = state.Creatures[c]
		}
	}
	return creature
}

// check if creature is scanned by drone
func (state *GameState) isCreatureScanned(creatureId, droneId int) bool {
	for _, scan := range state.Scans[droneId] {
		if scan.CreatureId == creatureId {
			log("Creature", creatureId, "is scanned by drone", droneId)
			return true
		}
	}
	return false
}

// check if creature is scanned by any of my drones
func (state *GameState) isCreatureScannedByMyDrones(creatureId int) bool {
	for _, drone := range state.MyDrones {
		if state.isCreatureScanned(creatureId, drone.Id) {
			return true
		}
	}
	return false
}

// Updates state of creature if no radar blips are found from any of drones means creature is dead, remove it from targets
func (state *GameState) updateDeadCreatures() {
	for c := range state.Creatures {
		if len(state.Creatures[c].RadarBlips) == 0 {
			state.Creatures[c].Dead = true
		}
	}

	for c := range state.Creatures {
		if !state.Creatures[c].Dead {
			continue
		}
		for d := range state.MyDrones {
			if state.MyDrones[d].TargetCreature != nil && state.MyDrones[d].TargetCreature.Id == state.Creatures[c].Id {
				state.MyDrones[d].TargetCreature = nil
			}
		}
	}
}

// Update creature approximate position based last vx and vy
func (state *GameState) updateCreaturesPosition() {
	for c := range state.Creatures {
		if state.Creatures[c].Visible {
			continue
		}
		state.Creatures[c].X += state.Creatures[c].Vx
		state.Creatures[c].Y += state.Creatures[c].Vy
	}
}

// Check if creature is saved already by me or foe
func (state *GameState) isCreatureSaved(creatureId int) bool {
	for _, id := range state.MySavedScans {
		if id == creatureId {
			return true
		}
	}
	for _, id := range state.FoeSavedScan {
		if id == creatureId {
			return true
		}
	}
	return false
}

// Get creature by id
func (state *GameState) getCreatureById(creatureId int) *Creature {
	for c := range state.Creatures {
		if state.Creatures[c].Id == creatureId {
			return state.Creatures[c]
		}
	}
	return nil
}

// print all monsters with field values
func (state *GameState) printMonsters() {
	for c := range state.Creatures {
		if state.Creatures[c].isMonster() {
			log("Monster", state.Creatures[c].String())
		}
	}
}

// Reset creature visibility
func (state *GameState) resetCreatureVisibility() {
	for c := range state.Creatures {
		state.Creatures[c].Visible = false
	}
}

// approximateFishPositions estimates the positions of fishes based on radar blips from multiple drones.
func (state *GameState) approximateFishPositions() {
	if state.Initialized {
		return
	}
	log("Approximating fish positions")
	for _, creature := range state.Creatures {

		// Get habitat depth range for the creature type
		habitat := fishHbitat[creature.Type]

		// Variables to store the sum of estimated positions and count of blips
		var sumX, sumY, blipCount int

		// Loop through each drone's radar blips for this creature
		for _, drone := range state.MyDrones {
			if blip, exists := drone.RadarBlips[creature.Id]; exists {
				// Convert radar blip to approximate position
				approxX, approxY := convertBlipToApproxPosition(blip, drone, habitat)
				sumX += approxX
				sumY += approxY
				blipCount++
			}
		}

		// Calculate average position if the fish is detected by any drone
		if blipCount > 0 {
			creature.X = sumX / blipCount
			creature.Y = sumY / blipCount
			creature.Vy = rand.Intn(50)
			creature.Vx = rand.Intn(50)
			log("Creature", creature.Id, "approximate position", creature.X, creature.Y)
		}
	}
}

// convertBlipToApproxPosition converts a radar blip into an approximate position.
func convertBlipToApproxPosition(blip RadarBlip, drone *Drone, habitat []int) (int, int) {
	var approxX, approxY int

	// Estimate the X position based on the drone's position and radar blip
	switch blip {
	case TopLeft, BottomLeft:
		approxX = drone.X - 500 // Adjust offset as needed
	case TopRight, BottomRight:
		approxX = drone.X + 500 // Adjust offset as needed
	}

	// Estimate the Y position within the fish's habitat range
	habitatRange := habitat[1] - habitat[0]
	switch blip {
	case TopLeft, TopRight:
		approxY = habitat[0] + habitatRange/4 // Adjust as needed
	case BottomLeft, BottomRight:
		approxY = habitat[1] - habitatRange/4 // Adjust as needed
	}

	return approxX, approxY
}

// check if creature is monster
func (creature *Creature) isMonster() bool {
	return creature.Type == Monster
}

// creature to string with field values
func (creature *Creature) String() string {
	return fmt.Sprintf("Id: %d, Type: %d, X: %d, Y: %d, Vx: %d, Vy: %d, Dead: %t", creature.Id, creature.Type, creature.X, creature.Y, creature.Vx, creature.Vy, creature.Dead)
}

func log(messages ...any) {
	fmt.Fprintln(os.Stderr, messages)
}

func distance(x1, y1, x2, y2 int) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(float64(dx*dx + dy*dy))
}

// Helper function to determine the sign of a number
func sign(x int) int {
	if x < 0 {
		return -1
	} else if x > 0 {
		return 1
	}
	return 0
}
