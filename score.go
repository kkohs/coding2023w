package main

// CalculatePotentialPoints calculates the potential score current turn if all my drones ascend
func (state *GameState) CalculatePotentialPoints() int {
	potentialPoints := state.MyScore // Start with current score

	// Temporary map to hold counts of each type and color of fish scanned
	typeCounts := make(map[CreatureType]int)
	colorCounts := make(map[int]int)
	firstScanBonus := make(map[int]bool) // Track if you're the first to scan a particular fish

	// Check scans in both of your drones
	for _, drone := range state.MyDrones {
		for _, scan := range drone.Scans {
			if !scan.IsDelivered(state) {
				potentialPoints += getScanPoints(scan.Type, firstScanBonus[scan.Id])
				typeCounts[scan.Type]++
				colorCounts[scan.Color]++
				firstScanBonus[scan.Id] = true // Assume you're the first for this calculation
			}
		}
	}

	// Add bonus points for scanning all fish of one type or color
	for _, count := range typeCounts {
		if count == len(state.Creatures) {
			potentialPoints += 4 // Bonus for all fish of one type
		}
	}

	for _, count := range colorCounts {
		if count == len(state.Creatures) {
			potentialPoints += 3 // Bonus for all fish of one color
		}
	}

	return potentialPoints
}

func getScanPoints(fishType CreatureType, first bool) int {
	points := 0
	switch fishType {
	case ShallowFish:
		points = 1
	case MediumFish:
		points = 2
	case DeepFish:
		points = 3
	}

	if first {
		points *= 2 // Double points for being the first to save the scan
	}

	return points
}
