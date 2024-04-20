package internal

import (
	"encoding/json"
	"math/rand"
	"os"
	"sort"
)

// NOTE: we take advantage of golang's defaults to false for units/resources
// since a "get" in a map that does not have the keys is false
type buildingSpecs struct {
	ResourceMultiplier []float64              `json:"resourceMultiplier"`
	TrainingMultiplier []float64              `json:"trainingMultiplier"`
	UpgradeCost        []tResourcesCount      `json:"cost"`
	UpgradeSpeed       []tSec                 `json:"upgradeSpeed"`
	MaxLevel           tBuildingLevel         `json:"maxLevel"`
	Units              map[tUnitName]bool     `json:"units"`
	Resources          map[tResourceName]bool `json:"resources"`
}

type unitSpecs struct {
	UnitSpeed              tSpeed                           `json:"speed"`
	UnitProductionSpeedSec tSec                             `json:"productionSpeed"`
	UnitCost               tResourcesCount                  `json:"cost"`
	CombatStats            map[tUnitStatName]tUnitStatPower `json:"stats"`
	CarryCapacity          tResourceCount                   `json:"carryCapacity"`
}

type gameConfig struct {
	Buildings           map[tBuildingName]buildingSpecs `json:"buildings"`
	Units               map[tUnitName]unitSpecs         `json:"units"`
	ResourceTrickles    tResourcesCount                 `json:"resources"`
	ForagingCoefficient float64
	CombatEfficiency    float64
}

var (
	cfg gameConfig

	// pre-computations
	sortedSlowestUnits            []tUnitName
	cumulativeResourceMultipliers map[tResourceName][]tBuildingName
	cumulativeTrainingMultipliers map[tUnitName][]tBuildingName
)

func init() {
	rawConfig, err := os.ReadFile("../config.json") // TODO: simplify the path for configuration reading
	if err != nil {
		panic(err)
	}

	cfg = gameConfig{}

	err = json.Unmarshal(rawConfig, &cfg)
	if err != nil {
		panic(err)
	}

	// TODO: basic checks for consistency
	// * levels / costs / multipliers match
	// * costs are done with existing resources
	// * check if buildings define training multipliers if they train units
	// * check if buildings define trickle multipliers if they produce resources
	// * ensure all units have a value for each stat type (even if zero)

	// NOTE: setup some pre-calculations for easier game logic

	sortedSlowestUnits = make([]tUnitName, 0, len(cfg.Units))
	for k := range cfg.Units {
		sortedSlowestUnits = append(sortedSlowestUnits, k)
	}
	sort.Slice(sortedSlowestUnits, func(i, j int) bool {
		return cfg.Units[sortedSlowestUnits[i]].UnitSpeed < cfg.Units[sortedSlowestUnits[j]].UnitSpeed
	})

	readOnlyResourceMultipliers := make(map[tResourceName][]tBuildingName, len(cfg.ResourceTrickles))
	for resourceKey := range cfg.ResourceTrickles {
		readOnlyResourceMultipliers[resourceKey] = make([]tBuildingName, 0)
		for buildingKey, building := range cfg.Buildings {
			if !building.Resources[resourceKey] {
				continue
			}
			readOnlyResourceMultipliers[resourceKey] = append(readOnlyResourceMultipliers[resourceKey], buildingKey)
		}
	}
	readOnlyTrainingMultipliers := make(map[tUnitName][]tBuildingName, len(cfg.Units))
	for unitKey := range cfg.Units {
		readOnlyTrainingMultipliers[unitKey] = make([]tBuildingName, 0)
		for buildingKey, building := range cfg.Buildings {
			if !building.Units[unitKey] {
				continue
			}
			readOnlyTrainingMultipliers[unitKey] = append(readOnlyTrainingMultipliers[unitKey], buildingKey)
		}
	}

	// TODO efficiency range can come from config + future bonuses
	cfg.CombatEfficiency = rand.Float64()*0.3 + 0.7
}
