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

type militaryBuildingSpecs struct {
	Multiplier []float32          `json:"multiplier"`
	Units      map[tUnitName]bool `json:"units"`
}

type economicBuildingSpecs struct {
	Multiplier []float32              `json:"multiplier"`
	Resources  map[tResourceName]bool `json:"resources"`
}

type unitSpecs struct {
	UnitSpeed              float32        `json:"speed"`
	UnitProductionSpeedSec int32          `json:"productionSpeed"`
	UnitCost               tResourceCount `json:"cost"`
}

type gameConfig struct {
	MilitaryBuildings map[tBuildingName]militaryBuildingSpecs `json:"militaryBuildings"`
	EconomicBuildings map[tBuildingName]economicBuildingSpecs `json:"economicBuildings"`
	Units             map[tUnitName]unitSpecs                 `json:"units"`
	ResourceTrickles  map[tResourceName]tResourceTrickle      `json:"resources"`
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
