package stickerio

import (
	"encoding/json"
	"os"
)

type BuildingName string
type BuildingSpecs struct {
	Multiplier []float64 `json:"multiplier"`
}

const (
	Mine     BuildingName = "mines"
	Barracks BuildingName = "barracks"
)

type UnitName string
type UnitSpecs struct {
	UnitSpeed              float64 `json:"speed"`
	UnitProductionSpeedSec int     `json:"production_speed"`
}

const (
	Stickmen  UnitName = "stickmen"
	Swordsmen UnitName = "swordsmen"
)

type ResourceTrickle int
type ResourceName string

const (
	Sticks  ResourceName = "sticks"
	Circles ResourceName = "circles"
)

type GameConfig struct {
	Buildings        map[BuildingName]BuildingSpecs
	Units            map[UnitName]UnitSpecs
	ResourceTrickles map[ResourceName]ResourceTrickle
}

var (
	Config GameConfig
)

func init() {
	rawConfig, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	Config = GameConfig{}

	err = json.Unmarshal(rawConfig, &Config)
	if err != nil {
		panic(err)
	}
}
