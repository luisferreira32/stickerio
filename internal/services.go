package internal

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type viewerRepository interface {
	GetCity(ctx context.Context, id, playerID string) (*dbCity, error)
	GetCityInfo(ctx context.Context, id string) (*dbCity, error)
	ListCityInfo(ctx context.Context, lastID string, pageSize int, filters ...listCityInfoFilterOpt) ([]*dbCity, error)
	GetMovement(ctx context.Context, id, playerID string) (*dbMovement, error)
	ListMovements(ctx context.Context, playerID, lastID string, pageSize int, filters ...listMovementsFilterOpt) ([]*dbMovement, error)
	GetUnitQueueItem(ctx context.Context, id, cityID string) (*dbUnitQueueItem, error)
	ListUnitQueueItems(ctx context.Context, cityID, lastID string, pageSize int) ([]*dbUnitQueueItem, error)
	GetBuildingQueueItem(ctx context.Context, id, cityID string) (*dbBuildingQueueItem, error)
	ListBuildingQueueItems(ctx context.Context, cityID, lastID string, pageSize int) ([]*dbBuildingQueueItem, error)
}

type viewerService struct {
	repository viewerRepository
}

type city struct {
	id             string
	name           string
	playerID       string
	locationX      int32
	locationY      int32
	buildingsLevel map[string]int64
	resourceBase   map[string]int64
	resourceEpoch  int64
	unitCount      map[string]int64
}

func cityFromDBModel(dbCity *dbCity) (*city, error) {
	resourceBase := make(map[string]int64)
	err := json.Unmarshal([]byte(dbCity.resourceBase), &resourceBase)
	if err != nil {
		return nil, err
	}
	unitCount := make(map[string]int64)
	err = json.Unmarshal([]byte(dbCity.unitCount), &unitCount)
	if err != nil {
		return nil, err
	}
	buildingsLevel := make(map[string]int64)
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

func (s *viewerService) GetCity(ctx context.Context, id, playerID string) (*city, error) {
	dbCity, err := s.repository.GetCityInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	return cityFromDBModel(dbCity)
}

func (s *viewerService) GetCityInfo(ctx context.Context, id string) (*city, error) {
	dbCity, err := s.repository.GetCityInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	return cityFromDBModel(dbCity)
}

func (s *viewerService) ListCityInfo(ctx context.Context, lastID string, pageSize int, playerIDFilter string, locationBoundsFilter string) ([]*city, error) {
	additionalFilters := make([]listCityInfoFilterOpt, 0)
	if playerIDFilter != "" {
		additionalFilters = append(additionalFilters, withPlayerID(playerIDFilter))
	}
	if locationBoundsFilter != "" {
		// TODO: fix this
	}

	dbCities, err := s.repository.ListCityInfo(ctx, lastID, pageSize, additionalFilters...)
	if err != nil {
		return nil, err
	}
	cities := make([]*city, len(dbCities))
	for i := 0; i < len(dbCities); i++ {
		city, err := cityFromDBModel(dbCities[i])
		if err != nil {
			return nil, err
		}
		cities[i] = city
	}
	return cities, nil
}

type movement struct {
	id             string
	playerID       string
	originID       string
	destinationID  string
	destinationX   int32
	destinationY   int32
	departureEpoch int64
	speed          float64
	resourceCount  map[string]int64
	unitCount      map[string]int64
}

func movementFromDBModel(dbMovement *dbMovement) (*movement, error) {
	resourceCount := make(map[string]int64)
	err := json.Unmarshal([]byte(dbMovement.resourceCount), &resourceCount)
	if err != nil {
		return nil, err
	}
	unitCount := make(map[string]int64)
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

func (s *viewerService) GetMovement(ctx context.Context, id, playerID string) (*movement, error) {
	movement, err := s.repository.GetMovement(ctx, id, playerID)
	if err != nil {
		return nil, err
	}

	return movementFromDBModel(movement)
}

func (s *viewerService) ListMovements(ctx context.Context, playerID, lastID string, pageSize int, originIDFilter, destinationIDFilter string) ([]*movement, error) {
	additionalFilters := make([]listMovementsFilterOpt, 0)
	if originIDFilter != "" {
		additionalFilters = append(additionalFilters, withOriginCityID(originIDFilter))
	}
	if destinationIDFilter != "" {
		additionalFilters = append(additionalFilters, withDestinationCityID(destinationIDFilter))
	}
	dbMovements, err := s.repository.ListMovements(ctx, playerID, lastID, pageSize, additionalFilters...)
	if err != nil {
		return nil, err
	}

	movements := make([]*movement, len(dbMovements))
	for i := 0; i < len(dbMovements); i++ {
		movement, err := movementFromDBModel(dbMovements[i])
		if err != nil {
			return nil, err
		}
		movements[i] = movement
	}
	return movements, nil
}

type unitQueueItem struct {
	id          string
	cityID      string
	queuedEpoch int64
	durationSec int64
	unitCount   int64
	unitType    string
}

func unitQueueItemFromDBModel(item *dbUnitQueueItem) *unitQueueItem {
	return &unitQueueItem{
		id:          item.id,
		cityID:      item.cityID,
		queuedEpoch: item.queuedEpoch,
		durationSec: item.durationSec,
		unitCount:   item.unitCount,
		unitType:    item.unitType,
	}
}

func unitQueueItemToDBModel(item *unitQueueItem) *dbUnitQueueItem {
	return &dbUnitQueueItem{
		id:          item.id,
		cityID:      item.cityID,
		queuedEpoch: item.queuedEpoch,
		durationSec: item.durationSec,
		unitCount:   item.unitCount,
		unitType:    item.unitType,
	}
}

func (s *viewerService) GetUnitQueueItem(ctx context.Context, id, cityID, playerID string) (*unitQueueItem, error) {
	city, err := s.repository.GetCity(ctx, cityID, playerID)
	if err != nil {
		return nil, err
	}

	item, err := s.repository.GetUnitQueueItem(ctx, id, city.id)
	if err != nil {
		return nil, err
	}

	return unitQueueItemFromDBModel(item), nil
}

func (s *viewerService) ListUnitQueueItems(ctx context.Context, cityID, playerID, lastID string, pageSize int) ([]*unitQueueItem, error) {
	city, err := s.repository.GetCity(ctx, cityID, playerID)
	if err != nil {
		return nil, err
	}

	dbItems, err := s.repository.ListUnitQueueItems(ctx, city.id, lastID, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*unitQueueItem, len(dbItems))
	for i := 0; i < len(dbItems); i++ {
		items[i] = unitQueueItemFromDBModel(dbItems[i])
	}
	return items, nil
}

type buildingQueueItem struct {
	id             string
	cityID         string
	queuedEpoch    int64
	durationSec    int64
	targetLevel    int32
	targetBuilding string
}

func buildingQueueItemFromDBModel(item *dbBuildingQueueItem) *buildingQueueItem {
	return &buildingQueueItem{
		id:             item.id,
		queuedEpoch:    item.queuedEpoch,
		durationSec:    item.durationSec,
		targetLevel:    item.targetLevel,
		targetBuilding: item.targetBuilding,
	}
}

func buildingQueueItemToDBModel(item *buildingQueueItem) *dbBuildingQueueItem {
	return &dbBuildingQueueItem{
		id:             item.id,
		queuedEpoch:    item.queuedEpoch,
		durationSec:    item.durationSec,
		targetLevel:    item.targetLevel,
		targetBuilding: item.targetBuilding,
	}
}

func (s *viewerService) GetBuildingQueueItem(ctx context.Context, id, cityID, playerID string) (*buildingQueueItem, error) {
	city, err := s.repository.GetCity(ctx, cityID, playerID)
	if err != nil {
		return nil, err
	}

	item, err := s.repository.GetBuildingQueueItem(ctx, id, city.id)
	if err != nil {
		return nil, err
	}

	return buildingQueueItemFromDBModel(item), nil
}

func (s *viewerService) ListBuildingQueueItems(ctx context.Context, cityID, playerID, lastID string, pageSize int) ([]*buildingQueueItem, error) {
	city, err := s.repository.GetCity(ctx, cityID, playerID)
	if err != nil {
		return nil, err
	}

	dbItems, err := s.repository.ListBuildingQueueItems(ctx, city.id, lastID, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*buildingQueueItem, len(dbItems))
	for i := 0; i < len(dbItems); i++ {
		items[i] = buildingQueueItemFromDBModel(dbItems[i])
	}
	return items, nil

}

type eventSourcer interface {
	queueEventHandling(e *event)
}

type eventInserter interface {
	InsertEvent(ctx context.Context, e *event) error
}

type inserterService struct {
	repository   eventInserter
	eventSourcer eventSourcer
}

func (s *inserterService) StartMovement(ctx context.Context, playerID string, m *movement) error {
	serverSideEpoch := time.Now().Unix()

	// important: these values cannot be trusted from the API
	// set them on the server side based on token / internal clock
	startMovement := &startMovementEvent{
		MovementID:     tMovementID(m.id),
		PlayerID:       tPlayerID(playerID),
		OriginID:       tCityID(m.originID),
		DestinationID:  tCityID(m.destinationID),
		DestinationX:   m.destinationX,
		DestinationY:   m.destinationY,
		DepartureEpoch: tEpoch(serverSideEpoch),
		UnitCount:      m.unitCount,
		ResourceCount:  m.resourceCount,
	}

	payload, err := json.Marshal(startMovement)
	if err != nil {
		return err
	}

	eventID := uuid.NewString()
	e := &event{
		id:      eventID,
		name:    startMovementEventName,
		epoch:   serverSideEpoch,
		payload: string(payload),
	}
	err = s.repository.InsertEvent(ctx, e)
	if err != nil {
		return err
	}
	s.eventSourcer.queueEventHandling(e)
	return nil
}

func (s *inserterService) QueueUnit(ctx context.Context, playerID string, item *unitQueueItem) error {
	serverSideEpoch := time.Now().Unix()

	queueItem := queueUnitEvent{
		UnitQueueItemID: tItemID(item.id),
		CityID:          tCityID(item.cityID),
		PlayerID:        tPlayerID(playerID),
		QueuedEpoch:     tEpoch(serverSideEpoch),
		UnitCount:       item.unitCount,
		UnitType:        tUnitName(item.unitType),
	}
	payload, err := json.Marshal(queueItem)
	if err != nil {
		return err
	}

	eventID := uuid.NewString()
	e := &event{
		id:      eventID,
		name:    queueUnitEventName,
		epoch:   serverSideEpoch,
		payload: string(payload),
	}

	err = s.repository.InsertEvent(ctx, e)
	if err != nil {
		return err
	}
	s.eventSourcer.queueEventHandling(e)
	return nil
}

func (s *inserterService) QueueBuilding(ctx context.Context, playerID string, item *buildingQueueItem) error {
	serverSideEpoch := time.Now().Unix()

	queueItem := queueBuildingEvent{
		BuildingQueueItemID: tItemID(item.id),
		CityID:              tCityID(item.cityID),
		PlayerID:            tPlayerID(playerID),
		QueuedEpoch:         tEpoch(serverSideEpoch),
		TargetLevel:         item.targetLevel,
		TargetBuilding:      tBuildingName(item.targetBuilding),
	}
	payload, err := json.Marshal(queueItem)
	if err != nil {
		return err
	}

	eventID := uuid.NewString()
	e := &event{
		id:      eventID,
		name:    queueBuildingEventName,
		epoch:   serverSideEpoch,
		payload: string(payload),
	}

	err = s.repository.InsertEvent(ctx, e)
	if err != nil {
		return err
	}
	s.eventSourcer.queueEventHandling(e)
	return nil
}
