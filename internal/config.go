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
	tUnitStats       map[string]int64
)

// NOTE: we take advantage of golang's defaults to false for units/resources
// since a "get" in a map that does not have the keys is false
type buildingSpecs struct {
	ResourceMultiplier []float64              `json:"resourceMultiplier"`
	TrainingMultiplier []float64              `json:"trainingMultiplier"`
	UpgradeCost        []tResourceCount       `json:"cost"`
	UpgradeSpeed       []int64                `json:"upgradeSpeed"`
	MaxLevel           int                    `json:"maxLevel"`
	Units              map[tUnitName]bool     `json:"units"`
	Resources          map[tResourceName]bool `json:"resources"`
}

type unitSpecs struct {
	UnitSpeed              float64        `json:"speed"`
	UnitProductionSpeedSec int64          `json:"productionSpeed"`
	UnitCost               tResourceCount `json:"cost"`
	CombatStats            tUnitStats     `json:"stats"`
	CarryCapacity          int64          `json:"carryCapacity"`
}

type gameConfig struct {
	Buildings           map[tBuildingName]buildingSpecs    `json:"buildings"`
	Units               map[tUnitName]unitSpecs            `json:"units"`
	ResourceTrickles    map[tResourceName]tResourceTrickle `json:"resources"`
	ForagingCoefficient float64
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
	// * check if buildings define training multipliers if they train units
	// * check if buildings define trickle multipliers if they produce resources

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
		for buildingKey, building := range readOnlyConfig.Buildings {
			if !building.Resources[resourceKey] {
				continue
			}
			readOnlyResourceMultipliers[resourceKey] = append(readOnlyResourceMultipliers[resourceKey], buildingKey)
		}
	}
	readOnlyTrainingMultipliers := make(map[tUnitName][]tBuildingName, len(readOnlyConfig.Units))
	for unitKey := range readOnlyConfig.Units {
		readOnlyTrainingMultipliers[unitKey] = make([]tBuildingName, 0)
		for buildingKey, building := range readOnlyConfig.Buildings {
			if !building.Units[unitKey] {
				continue
			}
			readOnlyTrainingMultipliers[unitKey] = append(readOnlyTrainingMultipliers[unitKey], buildingKey)
		}
	}
}
