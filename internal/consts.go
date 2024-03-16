package internal

import (
	"encoding/json"
	"os"
	"sort"
)

type (
	tMovementID      string
	tPlayerID        string
	tCityID          string
	tEpoch           int64
	tUnitCount       map[string]int64
	tResourceCount   map[string]int64
	tItemID          string
	tBuildingName    string
	tUnitName        string
	tResourceTrickle int64
	tResourceName    string
)

type buildingSpecs struct {
	Multiplier []float64 `json:"multiplier"`
}

type unitSpecs struct {
	UnitSpeed              float32 `json:"speed"`
	UnitProductionSpeedSec int     `json:"production_speed"`
}

type gameConfig struct {
	Buildings        map[tBuildingName]buildingSpecs
	Units            map[tUnitName]unitSpecs
	ResourceTrickles map[tResourceName]tResourceTrickle
}

var (
	config       gameConfig
	slowestUnits []tUnitName
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

	slowestUnits = make([]tUnitName, 0, len(config.Units))
	for k := range config.Units {
		slowestUnits = append(slowestUnits, k)
	}
	sort.Slice(slowestUnits, func(i, j int) bool {
		return config.Units[slowestUnits[i]].UnitSpeed < config.Units[slowestUnits[j]].UnitSpeed
	})
}
