package main

import (
	"fmt"
	"math"
	"os"
)

type GameState struct {
	Turn           int
	MapSize        int
	MyScore        int
	FoeScore       int
	Creatures      []Creature
	MyScans        []Scan
	FoeScans       []Scan
	MyDrones       []Drone
	FoeDrones      []Drone
	DroneScanCount map[int]int
}

type Creature struct {
	Id         int
	Color      int
	Type       int
	CreatureX  int
	CreatureY  int
	CreatureVx int
	CreatureVy int
	Visible    bool
}

type Drone struct {
	Id         int
	X          int
	Y          int
	LightPower bool
	Battery    int
}

type Scan struct {
	CreatureId int
}

type Action struct {
	Type       string
	TargetX    int
	TargetY    int
	LightPower bool
}

type Cluster struct {
	CentroidX, CentroidY int
	Fishes               []Creature
}

/**
 * Score points by scanning valuable fish faster than your opponent.
 **/

func main() {

	state := GameState{}
	var creatureCount int
	fmt.Scan(&creatureCount)
	state.Creatures = make([]Creature, creatureCount)

	for i := 0; i < creatureCount; i++ {
		var creatureId, color, _type int
		fmt.Scan(&creatureId, &color, &_type)
		state.Creatures[i] = Creature{creatureId, color, _type, 0, 0, 0, 0, false}
	}
	for {
		var myScore int
		fmt.Scan(&myScore)
		state.MyScore = myScore

		var foeScore int
		fmt.Scan(&foeScore)
		state.FoeScore = foeScore

		var myScanCount int
		fmt.Scan(&myScanCount)
		state.MyScans = make([]Scan, myScanCount)

		for i := 0; i < myScanCount; i++ {
			var creatureId int
			fmt.Scan(&creatureId)
			state.MyScans[i] = Scan{creatureId}
		}
		var foeScanCount int
		fmt.Scan(&foeScanCount)
		state.FoeScans = make([]Scan, foeScanCount)

		for i := 0; i < foeScanCount; i++ {
			var creatureId int
			fmt.Scan(&creatureId)
			state.FoeScans[i] = Scan{creatureId}
		}
		var myDroneCount int
		fmt.Scan(&myDroneCount)
		state.MyDrones = make([]Drone, myDroneCount)

		for i := 0; i < myDroneCount; i++ {
			var droneId, droneX, droneY, emergency, battery int
			fmt.Scan(&droneId, &droneX, &droneY, &emergency, &battery)
			state.MyDrones[i] = Drone{droneId, droneX, droneY, emergency == 1, battery}
		}
		var foeDroneCount int
		fmt.Scan(&foeDroneCount)
		state.FoeDrones = make([]Drone, foeDroneCount)

		for i := 0; i < foeDroneCount; i++ {
			var droneId, droneX, droneY, emergency, battery int
			fmt.Scan(&droneId, &droneX, &droneY, &emergency, &battery)
			state.FoeDrones[i] = Drone{droneId, droneX, droneY, emergency == 1, battery}
		}
		var droneScanCount int
		fmt.Scan(&droneScanCount)
		state.DroneScanCount = make(map[int]int)

		for i := 0; i < droneScanCount; i++ {
			var droneId, creatureId int
			fmt.Scan(&droneId, &creatureId)
			state.DroneScanCount[droneId] = creatureId
		}
		var visibleCreatureCount int
		fmt.Scan(&visibleCreatureCount)

		for i := 0; i < visibleCreatureCount; i++ {
			var creatureId, creatureX, creatureY, creatureVx, creatureVy int
			fmt.Scan(&creatureId, &creatureX, &creatureY, &creatureVx, &creatureVy)
			for j := 0; j < len(state.Creatures); j++ {
				if state.Creatures[j].Id == creatureId {
					state.Creatures[j].CreatureX = creatureX
					state.Creatures[j].CreatureY = creatureY
					state.Creatures[j].CreatureVx = creatureVx
					state.Creatures[j].CreatureVy = creatureVy
					state.Creatures[j].Visible = true
				}
			}
		}
		var radarBlipCount int
		fmt.Scan(&radarBlipCount)

		for i := 0; i < radarBlipCount; i++ {
			var droneId, creatureId int
			var radar string
			fmt.Scan(&droneId, &creatureId, &radar)
		}
		for i := 0; i < myDroneCount; i++ {
			drone := state.MyDrones[i]
			log("Drone", drone.Id, "is at", drone.X, drone.Y)

			// Find clusters with unscanned fish
			unscanned := unscannedFish(state.Creatures, append(state.MyScans, state.FoeScans...))
			clusters := findClusters(unscanned, 1000) // Example threshold
			targetCluster := selectTargetCluster(clusters, drone)

			// Count fish within a radius of 2000 units
			fishCount := countFishWithinRadius(unscanned, drone, 2000)

			// Determine light power
			lightPower := 0
			if fishCount > 2 && drone.Battery > 6 {
				lightPower = 1
			}

			if len(targetCluster.Fishes) > 0 {
				fmt.Println("MOVE", targetCluster.CentroidX, targetCluster.CentroidY, lightPower)
			} else {
				fmt.Println("WAIT", lightPower)
			}
		}
	}

}

func countFishWithinRadius(fishes []Creature, drone Drone, radius float64) int {
	count := 0
	for _, fish := range fishes {
		if distance(drone.X, drone.Y, fish.CreatureX, fish.CreatureY) <= radius {
			count++
		}
	}
	return count
}

// Function to calculate the centroid of a cluster
func centroid(fishes []Creature) (int, int) {
	var sumX, sumY int
	for _, fish := range fishes {
		sumX += fish.CreatureX
		sumY += fish.CreatureY
	}
	count := len(fishes)
	return sumX / count, sumY / count
}

// Function to find clusters of creatures
func findClusters(creatures []Creature, threshold float64) []Cluster {
	var clusters []Cluster
	visited := make(map[int]bool)

	for _, creature := range creatures {
		if visited[creature.Id] {
			continue
		}

		cluster := Cluster{Fishes: []Creature{creature}}
		visited[creature.Id] = true

		for _, otherCreature := range creatures {
			if !visited[otherCreature.Id] {
				dist := distance(creature.CreatureX, creature.CreatureY, otherCreature.CreatureX, otherCreature.CreatureY)
				if dist <= threshold {
					cluster.Fishes = append(cluster.Fishes, otherCreature)
					visited[otherCreature.Id] = true
				}
			}
		}

		cx, cy := centroid(cluster.Fishes)
		cluster.CentroidX = cx
		cluster.CentroidY = cy
		clusters = append(clusters, cluster)
	}

	return clusters
}
func selectTargetCluster(clusters []Cluster, drone Drone) Cluster {
	minDist := math.MaxFloat64
	var targetCluster Cluster

	for _, cluster := range clusters {
		dist := distance(drone.X, drone.Y, cluster.CentroidX, cluster.CentroidY)
		if dist < minDist {
			minDist = dist
			targetCluster = cluster
		}
	}

	return targetCluster
}

func unscannedFish(creatures []Creature, scans []Scan) []Creature {
	scanned := make(map[int]bool)
	for _, scan := range scans {
		scanned[scan.CreatureId] = true
	}

	var unscanned []Creature
	for _, creature := range creatures {
		if !scanned[creature.Id] {
			unscanned = append(unscanned, creature)
		}
	}
	return unscanned
}
func log(messages ...any) {
	fmt.Fprintln(os.Stderr, messages)
}

func distance(x1, y1, x2, y2 int) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(float64(dx*dx + dy*dy))
}
