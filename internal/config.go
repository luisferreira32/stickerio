package internal

import (
	"encoding/json"
	"fmt"
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
	Multiplier  []float64          `json:"multiplier"`
	UpgradeCost []tResourceCount   `json:"cost"`
	MaxLevel    int                `json:"maxLevel"`
	Units       map[tUnitName]bool `json:"units"`
}

type economicBuildingSpecs struct {
	Multiplier  []float64              `json:"multiplier"`
	UpgradeCost []tResourceCount       `json:"cost"`
	MaxLevel    int                    `json:"maxLevel"`
	Resources   map[tResourceName]bool `json:"resources"`
}

type unitSpecs struct {
	UnitSpeed              float64        `json:"speed"`
	UnitProductionSpeedSec int64          `json:"productionSpeed"`
	UnitCost               tResourceCount `json:"cost"`
}

type gameConfig struct {
	MilitaryBuildings map[tBuildingName]militaryBuildingSpecs `json:"militaryBuildings"`
	EconomicBuildings map[tBuildingName]economicBuildingSpecs `json:"economicBuildings"`
	Units             map[tUnitName]unitSpecs                 `json:"units"`
	ResourceTrickles  map[tResourceName]tResourceTrickle      `json:"resources"`
}

var (
	readOnlyConfig              gameConfig
	readOnlySlowestUnits        []tUnitName
	readOnlyResourceMultipliers map[tResourceName][]tBuildingName
	readOnlyTrainingMultipliers map[tUnitName][]tBuildingName
)

func init() {
	rawConfig, err := os.ReadFile("../config.json") // TODO: simplify the path for configuration reading
	if err != nil {
		panic(err)
	}

	readOnlyConfig = gameConfig{}

	err = json.Unmarshal(rawConfig, &readOnlyConfig)
	if err != nil {
		panic(err)
	}

	// TODO: basic checks for consistency
	// * levels / costs / multipliers match
	// * costs are done with existing resources

	for n, v := range readOnlyConfig.EconomicBuildings {
		if len(v.UpgradeCost) != len(v.Multiplier)-1 || len(v.UpgradeCost) != v.MaxLevel {
			panic(fmt.Sprintf("building %s has defined %d multipliers with max level %d and %d upgrade costs", n, len(v.UpgradeCost), v.MaxLevel, len(v.Multiplier)-1))
		}
	}

	for n, v := range readOnlyConfig.MilitaryBuildings {
		if len(v.UpgradeCost) != len(v.Multiplier)-1 || len(v.UpgradeCost) != v.MaxLevel {
			panic(fmt.Sprintf("building %s has defined %d multipliers with max level %d and %d upgrade costs", n, len(v.UpgradeCost), v.MaxLevel, len(v.Multiplier)-1))
		}
	}

	// NOTE: setup some pre-calculations for easier game logic

	readOnlySlowestUnits = make([]tUnitName, 0, len(readOnlyConfig.Units))
	for k := range readOnlyConfig.Units {
		readOnlySlowestUnits = append(readOnlySlowestUnits, k)
	}
	sort.Slice(readOnlySlowestUnits, func(i, j int) bool {
		return readOnlyConfig.Units[readOnlySlowestUnits[i]].UnitSpeed < readOnlyConfig.Units[readOnlySlowestUnits[j]].UnitSpeed
	})

	readOnlyResourceMultipliers := make(map[tResourceName][]tBuildingName, len(readOnlyConfig.ResourceTrickles))
	for resourceKey := range readOnlyConfig.ResourceTrickles {
		readOnlyResourceMultipliers[resourceKey] = make([]tBuildingName, 0)
		for buildingKey, building := range readOnlyConfig.EconomicBuildings {
			if !building.Resources[resourceKey] {
				continue
			}
			readOnlyResourceMultipliers[resourceKey] = append(readOnlyResourceMultipliers[resourceKey], buildingKey)
		}
	}
	readOnlyTrainingMultipliers := make(map[tUnitName][]tBuildingName, len(readOnlyConfig.Units))
	for unitKey := range readOnlyConfig.Units {
		readOnlyTrainingMultipliers[unitKey] = make([]tBuildingName, 0)
		for buildingKey, building := range readOnlyConfig.MilitaryBuildings {
			if !building.Units[unitKey] {
				continue
			}
			readOnlyTrainingMultipliers[unitKey] = append(readOnlyTrainingMultipliers[unitKey], buildingKey)
		}
	}
}
