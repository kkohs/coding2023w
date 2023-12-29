package main

type GameState struct {
	MyScore      int
	FoeScore     int
	MyScanCount  int
	FoeScanCount int
	MyDrones     []*Drone
	FoeDrones    []*Drone
	Creatures    []*Creature
	MyScans      []*Creature
	FoeScans     []*Creature
	Turn         int
}

// NewGameState returns a new GameState.
func NewGameState() *GameState {
	return &GameState{}
}

// UpdateMyDrone updates the drone with the given ID in the GameState's MyDrones or adds new if not present.
func (state *GameState) UpdateMyDrone(id, x, y, emergency, battery int) {
	for _, drone := range state.MyDrones {
		if drone.Id == id {
			drone.X = x
			drone.Y = y
			drone.Emergency = emergency
			drone.Battery = battery
			return
		}
	}
	state.MyDrones = append(state.MyDrones, &Drone{
		Id:        id,
		X:         x,
		Y:         y,
		Emergency: emergency,
		Battery:   battery,
	})
}

// UpdateFoeDrone updates the drone with the given ID in the GameState's FoeDrones or adds new if not present.
func (state *GameState) UpdateFoeDrone(id, x, y, emergency, battery int) {
	for _, drone := range state.FoeDrones {
		if drone.Id == id {
			drone.X = x
			drone.Y = y
			drone.Emergency = emergency
			drone.Battery = battery
			return
		}
	}
	state.FoeDrones = append(state.FoeDrones, &Drone{
		Id:        id,
		X:         x,
		Y:         y,
		Emergency: emergency,
		Battery:   battery,
	})
}

// AddCreature adds a creature to the GameState's Creatures slice.
func (state *GameState) AddCreature(creature *Creature) {
	if state.Creatures == nil {
		state.Creatures = make([]*Creature, 0)
	}
	state.Creatures = append(state.Creatures, creature)
}

// UpdateCreature updates the creature with the given ID in the GameState's
// Creatures slice.
func (state *GameState) UpdateCreature(id, x, y, vx, vy int) {
	for _, creature := range state.Creatures {
		if creature.Id == id {
			creature.X = x
			creature.Y = y
			creature.Vx = vx
			creature.Vy = vy
			creature.LastVisibleTurn = state.Turn
			return
		}
	}
}

// UpdateRadarBlip updates the radar blip with the given ID in the GameState's
// MyDrones or FoeDrones slice.
func (state *GameState) UpdateRadarBlip(droneId, creatureId int, radar string) {
	drone := state.GetDrone(droneId)
	if drone == nil {
		return
	}
	drone.AddRadarBlip(creatureId, RadarBlip(radar))
}

// AddMyScan adds a creature to the GameState's MyScans slice if not present.
func (state *GameState) AddMyScan(creatureId int) {
	for _, creature := range state.MyScans {
		if creature.Id == creatureId {
			return
		}
	}
	for _, creature := range state.Creatures {
		if creature.Id == creatureId {
			state.MyScans = append(state.MyScans, creature)
			return
		}
	}
}

// AddFoeScan adds a creature to the GameState's FoeScans slice if not present.
func (state *GameState) AddFoeScan(creatureId int) {
	for _, creature := range state.FoeScans {
		if creature.Id == creatureId {
			return
		}
	}
	for _, creature := range state.Creatures {
		if creature.Id == creatureId {
			state.FoeScans = append(state.FoeScans, creature)
			return
		}
	}
}

// GetDrone returns the drone with the given ID from the GameState's MyDrones or
// FoeDrones slice.
func (state *GameState) GetDrone(id int) *Drone {
	for _, drone := range state.MyDrones {
		if drone.Id == id {
			return drone
		}
	}
	for _, drone := range state.FoeDrones {
		if drone.Id == id {
			return drone
		}
	}
	return nil
}

// GetCreature returns the creature with the given ID from the GameState's
// Creatures slice.
func (state *GameState) GetCreature(id int) *Creature {
	for _, creature := range state.Creatures {
		if creature.Id == id {
			return creature
		}
	}
	return nil
}

// Print all state information each in new line
func (state *GameState) Print() {
	for _, creature := range state.Creatures {
		Log(creature.String())
	}

	for _, drone := range state.MyDrones {
		Log(drone.String())
	}
}

// NextTurn increments the turn counter
func (state *GameState) NextTurn() {
	state.Turn++

	// Clear scans
	state.MyScans = []*Creature{}
	state.FoeScans = []*Creature{}
	// Clear radar blips
	for _, drone := range state.MyDrones {
		drone.ClearRadarBlips()
	}
}
