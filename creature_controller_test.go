package main

import "testing"

func TestEstimatePosition_OneDroneBlipBL(t *testing.T) {
	state := NewGameState()
	state.UpdateMyDrone(0, 2500, 500, 0, 0)
	state.AddCreature(NewCreature(0, 0, ShallowFish))

	state.UpdateRadarBlip(0, 0, string(BottomLeft))
	state.EstimateAll()

	c := state.GetCreature(0)

	// Check if y is within the range of the fish type
	if c.Y < ShallowFishMinDepth || c.Y > ShallowFishMaxDepth {
		t.Errorf("Expected y to be within the range of the fish type, got %d", c.Y)
	}

	// check if estimation is withing bounds of radar blip
	if c.X < 0 || c.X > 2500 {
		t.Errorf("Expected x to be within the bounds of the radar blip, got %d", c.X)
	}
}

func TestEstimatePosition_OneDroneBlipBR(t *testing.T) {
	state := NewGameState()
	state.UpdateMyDrone(0, 2500, 500, 0, 0)
	state.AddCreature(NewCreature(0, 0, ShallowFish))

	state.UpdateRadarBlip(0, 0, string(BottomRight))
	state.EstimateAll()

	c := state.GetCreature(0)

	// Check if y is within the range of the fish type
	if c.Y < ShallowFishMinDepth || c.Y > ShallowFishMaxDepth {
		t.Errorf("Expected y to be within the range of the fish type, got %d", c.Y)
	}

	// check if estimation is withing bounds of radar blip
	if c.X < 2500 || c.X > 10000 {
		t.Errorf("Expected x to be within the bounds of the radar blip, got %d", c.X)
	}
}

func TestEstimatePosition_TwoDronesInBetween(t *testing.T) {

	state := NewGameState()
	state.UpdateMyDrone(0, 2500, 500, 0, 0)
	state.UpdateMyDrone(1, 7500, 500, 0, 0)
	state.AddCreature(NewCreature(0, 0, ShallowFish))

	state.UpdateRadarBlip(0, 0, string(BottomLeft))
	state.UpdateRadarBlip(1, 0, string(BottomRight))
	state.EstimateAll()

	c := state.GetCreature(0)

	// Check if y is within the range of the fish type
	if c.Y < ShallowFishMinDepth || c.Y > ShallowFishMaxDepth {
		t.Errorf("Expected y to be within the range of the fish type, got %d", c.Y)
	}

	// check if estimation is withing bounds of radar blip
	if c.X < 2500 || c.X > 7500 {
		t.Errorf("Expected x to be within the bounds of the radar blip, got %d", c.X)
	}
}

func TestEstimatePosition_TwoDronesInBetweenWithOneBlip(t *testing.T) {

	state := NewGameState()
	state.UpdateMyDrone(0, 2500, 500, 0, 0)
	state.UpdateMyDrone(1, 7500, 500, 0, 0)
	state.AddCreature(NewCreature(0, 0, ShallowFish))

	state.UpdateRadarBlip(0, 0, string(BottomLeft))
	state.UpdateRadarBlip(1, 0, string(BottomLeft))
	state.EstimateAll()

	c := state.GetCreature(0)

	// Check if y is within the range of the fish type
	if c.Y < ShallowFishMinDepth || c.Y > ShallowFishMaxDepth {
		t.Errorf("Expected y to be within the range of the fish type, got %d", c.Y)
	}

	// check if estimation is withing bounds of radar blip
	if c.X < 0 || c.X > 2500 {
		t.Errorf("Expected x to be within the bounds of the radar blip, got %d", c.X)
	}
}

func TestEstimatePosition_MultipleDronesMultipleFish(t *testing.T) {

	state := NewGameState()
	state.UpdateMyDrone(0, 2500, 500, 0, 0)
	state.UpdateMyDrone(1, 7500, 500, 0, 0)
	state.AddCreature(NewCreature(0, 0, ShallowFish))
	state.AddCreature(NewCreature(1, 0, ShallowFish))
	state.AddCreature(NewCreature(2, 0, ShallowFish))

	state.UpdateRadarBlip(0, 0, string(BottomLeft))
	state.UpdateRadarBlip(1, 0, string(BottomLeft))
	state.UpdateRadarBlip(0, 1, string(BottomLeft))
	state.UpdateRadarBlip(1, 1, string(BottomLeft))
	state.UpdateRadarBlip(0, 2, string(BottomLeft))
	state.UpdateRadarBlip(1, 2, string(BottomLeft))
	state.EstimateAll()

	for _, creature := range state.Creatures {

		// Check if y is within the range of the fish type
		if creature.Y < ShallowFishMinDepth || creature.Y > ShallowFishMaxDepth {
			t.Errorf("Expected y to be within the range of the fish type, got %d", creature.Y)
		}

		// check if estimation is withing bounds of radar blip
		if creature.X < 0 || creature.X > 2500 {
			t.Errorf("Expected x to be within the bounds of the radar blip, got %d", creature.X)
		}
		// check if no fish withing 600 units of other
		for _, fish := range state.Creatures {
			if fish.Id != creature.Id && distance(creature.X, creature.Y, fish.X, fish.Y) < 500 {
				t.Errorf("Expected no fish to be within 600 units of other fish, got %d", creature.X)
			}
		}
	}

}
