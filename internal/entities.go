package internal

import (
	"encoding/json"

	api "github.com/luisferreira32/stickerio/models"
)

type (
	tSec int64

	tMovementID          string
	tPlayerID            string
	tCityID              string
	tUnitQueueItemID     string
	tBuildingQueueItemID string

	tBuildingName string
	tUnitName     string
	tUnitStatName string
	tResourceName string

	tBuildingLevel int64
	tUnitCount     int64
	tUnitStatPower int64
	tResourceCount int64
	tCoordinate    int32
	tSpeed         float64

	tResourcesCount map[tResourceName]tResourceCount
	tBuildingsLevel map[tBuildingName]tBuildingLevel
	tUnitsCount     map[tUnitName]tUnitCount
)

func fromUntypedMap[K ~string, V ~int64](m map[string]int64) map[K]V {
	typedMap := make(map[K]V, len(m))
	for k, v := range m {
		typedMap[K(k)] = V(v)
	}
	return typedMap
}

func (t tResourcesCount) toUntypedMap() map[string]int64 {
	ut := make(map[string]int64, len(t))
	for k, v := range t {
		ut[string(k)] = int64(v)
	}
	return ut
}

func (t tBuildingsLevel) toUntypedMap() map[string]int64 {
	ut := make(map[string]int64, len(t))
	for k, v := range t {
		ut[string(k)] = int64(v)
	}
	return ut
}

func (t tUnitsCount) toUntypedMap() map[string]int64 {
	ut := make(map[string]int64, len(t))
	for k, v := range t {
		ut[string(k)] = int64(v)
	}
	return ut
}

type dbCity struct {
	id             tCityID
	name           string
	playerID       tPlayerID
	locationX      tCoordinate
	locationY      tCoordinate
	buildingsLevel string
	resourceBase   string
	resourceEpoch  tSec
	unitCount      string
}

type city struct {
	id             tCityID
	name           string
	playerID       tPlayerID
	locationX      tCoordinate
	locationY      tCoordinate
	buildingsLevel tBuildingsLevel
	resourceBase   tResourcesCount
	resourceEpoch  tSec
	unitCount      tUnitsCount
}

func cityToAPIModel(c *city) api.V1City {
	return api.V1City{
		CityInfo: api.V1CityInfo{
			Id:        string(c.id),
			Name:      string(c.name),
			PlayerID:  string(c.playerID),
			LocationX: int32(c.locationX),
			LocationY: int32(c.locationY),
		},
		Buildings: c.buildingsLevel.toUntypedMap(),
		CityResources: api.V1CityResources{
			Epoch:     int64(c.resourceEpoch),
			BaseCount: c.resourceBase.toUntypedMap(),
		},
		UnitCount: c.unitCount.toUntypedMap(),
	}
}

func cityToCityInfoAPIModel(c *city) api.V1CityInfo {
	return api.V1CityInfo{
		Id:        string(c.id),
		Name:      c.name,
		PlayerID:  string(c.playerID),
		LocationX: int32(c.locationX),
		LocationY: int32(c.locationY),
	}
}

func cityFromDBModel(dbCity *dbCity) (*city, error) {
	resourceBase := make(tResourcesCount)
	err := json.Unmarshal([]byte(dbCity.resourceBase), &resourceBase)
	if err != nil {
		return nil, err
	}
	unitCount := make(tUnitsCount)
	err = json.Unmarshal([]byte(dbCity.unitCount), &unitCount)
	if err != nil {
		return nil, err
	}
	buildingsLevel := make(tBuildingsLevel)
	err = json.Unmarshal([]byte(dbCity.buildingsLevel), &buildingsLevel)
	if err != nil {
		return nil, err
	}
	return &city{
		id:             dbCity.id,
		name:           dbCity.name,
		playerID:       dbCity.playerID,
		locationX:      dbCity.locationX,
		locationY:      dbCity.locationY,
		buildingsLevel: buildingsLevel,
		resourceBase:   resourceBase,
		resourceEpoch:  dbCity.resourceEpoch,
		unitCount:      unitCount,
	}, nil
}

func cityToDBModel(c *city) (*dbCity, error) {
	resourceBase, err := json.Marshal(c.resourceBase)
	if err != nil {
		return nil, err
	}
	unitCount, err := json.Marshal(c.unitCount)
	if err != nil {
		return nil, err
	}
	buildingsLevel, err := json.Marshal(c.buildingsLevel)
	if err != nil {
		return nil, err
	}
	return &dbCity{
		id:             c.id,
		name:           c.name,
		playerID:       c.playerID,
		locationX:      c.locationX,
		locationY:      c.locationY,
		buildingsLevel: string(buildingsLevel),
		resourceBase:   string(resourceBase),
		resourceEpoch:  c.resourceEpoch,
		unitCount:      string(unitCount),
	}, nil
}

type dbMovement struct {
	id             tMovementID
	playerID       tPlayerID
	originID       tCityID
	destinationID  tCityID
	destinationX   tCoordinate
	destinationY   tCoordinate
	departureEpoch tSec
	speed          tSpeed
	resourceCount  string
	unitCount      string
}

type movement struct {
	id             tMovementID
	playerID       tPlayerID
	originID       tCityID
	destinationID  tCityID
	destinationX   tCoordinate
	destinationY   tCoordinate
	departureEpoch tSec
	speed          tSpeed
	resourceCount  tResourcesCount
	unitCount      tUnitsCount
}

func movementToAPIModel(m *movement) api.V1Movement {
	resources := make(map[string]int64, len(m.resourceCount))
	for k, v := range m.resourceCount {
		resources[string(k)] = int64(v)
	}
	units := make(map[string]int64, len(m.unitCount))
	for k, v := range m.unitCount {
		units[string(k)] = int64(v)
	}
	return api.V1Movement{
		Id:             string(m.id),
		PlayerID:       string(m.playerID),
		OriginID:       string(m.originID),
		DestinationID:  string(m.destinationID),
		DestinationX:   int32(m.destinationX),
		DestinationY:   int32(m.destinationY),
		DepartureEpoch: int64(m.departureEpoch),
		Speed:          float64(m.speed),
		UnitCount:      units,
		ResourceCount:  resources,
	}
}

func movementFromDBModel(dbMovement *dbMovement) (*movement, error) {
	resourceCount := make(tResourcesCount)
	err := json.Unmarshal([]byte(dbMovement.resourceCount), &resourceCount)
	if err != nil {
		return nil, err
	}
	unitCount := make(tUnitsCount)
	err = json.Unmarshal([]byte(dbMovement.unitCount), &unitCount)
	if err != nil {
		return nil, err
	}
	return &movement{
		id:             dbMovement.id,
		playerID:       dbMovement.playerID,
		originID:       dbMovement.originID,
		destinationID:  dbMovement.destinationID,
		destinationX:   dbMovement.destinationX,
		destinationY:   dbMovement.destinationY,
		departureEpoch: dbMovement.departureEpoch,
		speed:          dbMovement.speed,
		resourceCount:  resourceCount,
		unitCount:      unitCount,
	}, nil
}

func movementToDBModel(m *movement) (*dbMovement, error) {
	resourceCount, err := json.Marshal(m.resourceCount)
	if err != nil {
		return nil, err
	}
	unitCount, err := json.Marshal(m.unitCount)
	if err != nil {
		return nil, err
	}
	return &dbMovement{
		id:             m.id,
		playerID:       m.playerID,
		originID:       m.originID,
		destinationID:  m.destinationID,
		destinationX:   m.destinationX,
		destinationY:   m.destinationY,
		departureEpoch: m.departureEpoch,
		speed:          m.speed,
		resourceCount:  string(resourceCount),
		unitCount:      string(unitCount),
	}, nil
}

type dbUnitQueueItem struct {
	id          tUnitQueueItemID
	cityID      tCityID
	playerID    tPlayerID
	queuedEpoch tSec
	durationSec tSec
	unitCount   tUnitCount
	unitType    tUnitName
}

type unitQueueItem struct {
	id          tUnitQueueItemID
	cityID      tCityID
	playerID    tPlayerID
	queuedEpoch tSec
	durationSec tSec
	unitCount   tUnitCount
	unitType    tUnitName
}

func unitQueueItemToAPIModel(item *unitQueueItem) api.V1UnitQueueItem {
	return api.V1UnitQueueItem{
		Id:          string(item.id),
		QueuedEpoch: int64(item.queuedEpoch),
		DurationSec: int64(item.durationSec),
		UnitCount:   int64(item.unitCount),
		UnitType:    string(item.unitType),
	}
}

func unitQueueItemFromDBModel(dbItem *dbUnitQueueItem) *unitQueueItem {
	return &unitQueueItem{
		id:          dbItem.id,
		cityID:      dbItem.cityID,
		playerID:    dbItem.playerID,
		queuedEpoch: dbItem.queuedEpoch,
		durationSec: dbItem.durationSec,
		unitCount:   dbItem.unitCount,
		unitType:    dbItem.unitType,
	}
}

func unitQueueItemToDBModel(item *unitQueueItem) *dbUnitQueueItem {
	return &dbUnitQueueItem{
		id:          item.id,
		cityID:      item.cityID,
		playerID:    item.playerID,
		queuedEpoch: item.queuedEpoch,
		durationSec: item.durationSec,
		unitCount:   item.unitCount,
		unitType:    item.unitType,
	}
}

type dbBuildingQueueItem struct {
	id             tBuildingQueueItemID
	cityID         tCityID
	playerID       tPlayerID
	queuedEpoch    tSec
	durationSec    tSec
	targetLevel    tBuildingLevel
	targetBuilding tBuildingName
}

type buildingQueueItem struct {
	id             tBuildingQueueItemID
	cityID         tCityID
	playerID       tPlayerID
	queuedEpoch    tSec
	durationSec    tSec
	targetLevel    tBuildingLevel
	targetBuilding tBuildingName
}

func buildingQueueItemToAPIModel(item *buildingQueueItem) api.V1BuildingQueueItem {
	return api.V1BuildingQueueItem{
		Id:          string(item.id),
		QueuedEpoch: int64(item.queuedEpoch),
		DurationSec: int64(item.durationSec),
		Level:       int64(item.targetLevel),
		Building:    string(item.targetBuilding),
	}
}

func buildingQueueItemFromDBModel(dbItem *dbBuildingQueueItem) *buildingQueueItem {
	return &buildingQueueItem{
		id:             dbItem.id,
		cityID:         dbItem.cityID,
		playerID:       dbItem.playerID,
		queuedEpoch:    dbItem.queuedEpoch,
		durationSec:    dbItem.durationSec,
		targetLevel:    dbItem.targetLevel,
		targetBuilding: dbItem.targetBuilding,
	}
}

func buildingQueueItemToDBModel(item *buildingQueueItem) *dbBuildingQueueItem {
	return &dbBuildingQueueItem{
		id:             item.id,
		cityID:         item.cityID,
		playerID:       item.playerID,
		queuedEpoch:    item.queuedEpoch,
		durationSec:    item.durationSec,
		targetLevel:    item.targetLevel,
		targetBuilding: item.targetBuilding,
	}
}

type event struct {
	id      string
	name    string
	epoch   tSec
	payload string // TODO: json encode might not be the most efficient way - improve this
}

type startMovementEvent struct {
	MovementID     tMovementID     `json:"movementID"`
	PlayerID       tPlayerID       `json:"playerID"`
	OriginID       tCityID         `json:"originID"`
	DestinationID  tCityID         `json:"destinationID"`
	DestinationX   tCoordinate     `json:"destinationX"`
	DestinationY   tCoordinate     `json:"destinationY"`
	DepartureEpoch tSec            `json:"departureEpoch"`
	UnitCount      tUnitsCount     `json:"unitCount"`
	ResourceCount  tResourcesCount `json:"resourceCount"`
}

type arrivalMovementEvent struct {
	MovementID    tMovementID     `json:"movementID"`
	PlayerID      tPlayerID       `json:"playerID"`
	OriginID      tCityID         `json:"originID"`
	DestinationID tCityID         `json:"destinationID"`
	DestinationX  tCoordinate     `json:"destinationX"`
	DestinationY  tCoordinate     `json:"destinationY"`
	UnitCount     tUnitsCount     `json:"unitCount"`
	ResourceCount tResourcesCount `json:"resourceCount"`
}

type returnMovementEvent struct {
	MovementID    tMovementID     `json:"movementID"`
	PlayerID      tPlayerID       `json:"playerID"`
	OriginID      tCityID         `json:"originID"`
	DestinationID tCityID         `json:"destinationID"`
	DestinationX  tCoordinate     `json:"destinationX"`
	DestinationY  tCoordinate     `json:"destinationY"`
	UnitCount     tUnitsCount     `json:"unitCount"`
	ResourceCount tResourcesCount `json:"resourceCount"`
}

type queueUnitEvent struct {
	UnitQueueItemID tUnitQueueItemID `json:"unitQueueItemID"`
	CityID          tCityID          `json:"cityID"`
	PlayerID        tPlayerID        `json:"playerID"`
	UnitCount       tUnitCount       `json:"unitCount"`
	UnitType        tUnitName        `json:"unitType"`
}

type createUnitEvent struct {
	UnitQueueItemID tUnitQueueItemID `json:"unitQueueItemID"`
	CityID          tCityID          `json:"cityID"`
	PlayerID        tPlayerID        `json:"playerID"`
	UnitCount       tUnitCount       `json:"unitCount"`
	UnitType        tUnitName        `json:"unitType"`
}

type queueBuildingEvent struct {
	BuildingQueueItemID tBuildingQueueItemID `json:"buildingQueueItemID"`
	CityID              tCityID              `json:"cityID"`
	PlayerID            tPlayerID            `json:"playerID"`
	TargetLevel         tBuildingLevel       `json:"targetLevel"`
	TargetBuilding      tBuildingName        `json:"targetBuilding"`
}

type upgradeBuildingEvent struct {
	BuildingQueueItemID tBuildingQueueItemID `json:"buildingQueueItemID"`
	CityID              tCityID              `json:"cityID"`
	PlayerID            tPlayerID            `json:"playerID"`
	TargetLevel         tBuildingLevel       `json:"targetLevel"`
	TargetBuilding      tBuildingName        `json:"targetBuilding"`
}

type createCityEvent struct {
	CityID        tCityID         `json:"cityID"`
	Name          string          `json:"name"`
	PlayerID      tPlayerID       `json:"playerID"`
	LocationX     tCoordinate     `json:"locationX"`
	LocationY     tCoordinate     `json:"locationY"`
	ResourceCount tResourcesCount `json:"resourceCount"`
	UnitCount     tUnitsCount     `json:"unitCount"`
}

type deleteCityEvent struct {
	CityID   tCityID   `json:"cityID"`
	PlayerID tPlayerID `json:"playerID"`
}
