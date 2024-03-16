package internal

import (
	"encoding/json"
	"os"
)

type buildingName string
type buildingSpecs struct {
	Multiplier []float64 `json:"multiplier"`
}

const (
	mine     buildingName = "mines"
	barracks buildingName = "barracks"
)

type unitName string
type unitSpecs struct {
	UnitSpeed              float64 `json:"speed"`
	UnitProductionSpeedSec int     `json:"production_speed"`
}

const (
	stickmen  unitName = "stickmen"
	swordsmen unitName = "swordsmen"
)

type resourceTrickle int
type resourceName string

const (
	sticks  resourceName = "sticks"
	circles resourceName = "circles"
)

type gameConfig struct {
	Buildings        map[buildingName]buildingSpecs
	Units            map[unitName]unitSpecs
	ResourceTrickles map[resourceName]resourceTrickle
}

var (
	config gameConfig
)

func init() {
	rawConfig, err := os.ReadFile("../config.json") // TODO: simplify the path for configuration reading
	if err != nil {
		panic(err)
	}

	config = gameConfig{}

	err = json.Unmarshal(rawConfig, &config)
	if err != nil {
		panic(err)
	}
}
