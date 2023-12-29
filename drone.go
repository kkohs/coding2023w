package main

import "fmt"

type Drone struct {
	Id         int
	X          int
	Y          int
	Emergency  int
	Battery    int
	Scans      []*Creature
	RadarBlips map[int]RadarBlip
	Target     *Creature
}

type RadarBlip string

const (
	TopRight    RadarBlip = "TR"
	TopLeft     RadarBlip = "TL"
	BottomRight RadarBlip = "BR"
	BottomLeft  RadarBlip = "BL"
)

// AddScan to the drone's Scans slice if not present already.
func (drone *Drone) AddScan(creature *Creature) {
	if drone.Scans == nil {
		drone.Scans = []*Creature{}
	}
	if creature == nil {
		return
	}
	for _, scan := range drone.Scans {
		if scan.Id == creature.Id {
			return
		}
	}
	drone.Scans = append(drone.Scans, creature)
}

// ClearScans clears the drone's Scans slice.
func (drone *Drone) ClearScans() {
	drone.Scans = []*Creature{}
}

// AddRadarBlip to the drone's RadarBlips map if not present already.
func (drone *Drone) AddRadarBlip(creatureId int, radar RadarBlip) {
	if drone.RadarBlips == nil {
		drone.RadarBlips = make(map[int]RadarBlip)
	}
	if _, ok := drone.RadarBlips[creatureId]; !ok {
		drone.RadarBlips[creatureId] = radar
	}
}

// String returns a string representation of the Drone with field names.
func (drone *Drone) String() string {
	return fmt.Sprintf("Drone{Id: %d, X: %d, Y: %d, Emergency: %d, Battery: %d, Scans: %v, RadarBlips: %v}", drone.Id, drone.X, drone.Y, drone.Emergency, drone.Battery, drone.Scans, drone.RadarBlips)
}

// ClearRadarBlips clears the drone's RadarBlips map.
func (drone *Drone) ClearRadarBlips() {
	drone.RadarBlips = make(map[int]RadarBlip)
}
