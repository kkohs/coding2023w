package main

import "fmt"

type CreatureType int

const (
	Monster     CreatureType = -1
	ShallowFish CreatureType = 0
	MediumFish  CreatureType = 1
	DeepFish    CreatureType = 2

	ShallowFishMinDepth = 2500
	MediumFishMinDepth  = 5000
	DeepFishMinDepth    = 7500

	ShallowFishMaxDepth = 5000
	MediumFishMaxDepth  = 7500
	DeepFishMaxDepth    = 10000
	FishCollision       = 600

	FishSpeed = 200
)

var (
	fishDepthsByType = map[CreatureType][2]int{
		ShallowFish: {ShallowFishMinDepth, ShallowFishMaxDepth},
		MediumFish:  {MediumFishMinDepth, MediumFishMaxDepth},
		DeepFish:    {DeepFishMinDepth, DeepFishMaxDepth},
	}
)

type Creature struct {
	Id              int
	Color           int
	Type            CreatureType
	X               int
	Y               int
	Vx              int
	Vy              int
	LastVisibleTurn int
}

// NewCreature returns a new Creature with the given ID, color and type.
func NewCreature(id, color int, _type CreatureType) *Creature {
	return &Creature{
		Id:              id,
		Color:           color,
		Type:            _type,
		LastVisibleTurn: -1,
	}
}

// Perform creature movement based on its type and current position vx,vy and nearby creatures and drones
func (creature *Creature) Move(state *GameState) {
	// Get minY and maxY for the creature type
	dimensionBoundaries := fishDepthsByType[creature.Type]
	minY := dimensionBoundaries[0]
	maxY := dimensionBoundaries[1]

	// Add velocity to position
	creature.X += creature.Vx
	creature.Y += creature.Vy

	// if not monster and there is fish withing 600 units, change direction
	if creature.Type != Monster {
		for _, fish := range state.Creatures {
			if fish.Type == Monster {
				continue
			}
			if fish.Id == creature.Id {
				continue
			}
			if fish.X > creature.X-FishCollision && fish.X < creature.X+FishCollision && fish.Y > creature.Y-FishCollision && fish.Y < creature.Y+FishCollision {
				creature.Vx = -creature.Vx
				creature.Vy = -creature.Vy
				break
			}
		}
	}

	// Adjust for boundaries
	if creature.X < 0 || creature.X > 10000 {
		creature.Vx = -creature.Vx // Reverse X velocity
	}

	if creature.Y < minY || creature.Y > maxY {
		creature.Vy = -creature.Vy // Reverse Y velocity
	}
}

// String returns a string representation of the Creature with field names.
func (creature *Creature) String() string {
	return fmt.Sprintf("Creature{Id: %d, Color: %d, Type: %d, X: %d, Y: %d, Vx: %d, Vy: %d, LastVisibleTurn: %d}", creature.Id, creature.Color, creature.Type, creature.X, creature.Y, creature.Vx, creature.Vy, creature.LastVisibleTurn)
}
