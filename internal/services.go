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
	GetUnitQueueItem(ctx context.Context, id, cityID, playerID string) (*dbUnitQueueItem, error)
	ListUnitQueueItems(ctx context.Context, cityID, playerID, lastID string, pageSize int) ([]*dbUnitQueueItem, error)
	GetBuildingQueueItem(ctx context.Context, id, cityID, playerID string) (*dbBuildingQueueItem, error)
	ListBuildingQueueItems(ctx context.Context, cityID, playerID, lastID string, pageSize int) ([]*dbBuildingQueueItem, error)
}

type viewerService struct {
	repository viewerRepository
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

func (s *viewerService) GetUnitQueueItem(ctx context.Context, id, cityID, playerID string) (*unitQueueItem, error) {
	item, err := s.repository.GetUnitQueueItem(ctx, id, cityID, playerID)
	if err != nil {
		return nil, err
	}

	return unitQueueItemFromDBModel(item), nil
}

func (s *viewerService) ListUnitQueueItems(ctx context.Context, cityID, playerID, lastID string, pageSize int) ([]*unitQueueItem, error) {
	dbItems, err := s.repository.ListUnitQueueItems(ctx, cityID, playerID, lastID, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*unitQueueItem, len(dbItems))
	for i := 0; i < len(dbItems); i++ {
		items[i] = unitQueueItemFromDBModel(dbItems[i])
	}
	return items, nil
}

func (s *viewerService) GetBuildingQueueItem(ctx context.Context, id, cityID, playerID string) (*buildingQueueItem, error) {
	item, err := s.repository.GetBuildingQueueItem(ctx, id, cityID, playerID)
	if err != nil {
		return nil, err
	}

	return buildingQueueItemFromDBModel(item), nil
}

func (s *viewerService) ListBuildingQueueItems(ctx context.Context, cityID, playerID, lastID string, pageSize int) ([]*buildingQueueItem, error) {
	dbItems, err := s.repository.ListBuildingQueueItems(ctx, cityID, playerID, lastID, pageSize)
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
	serverSideEpoch := tSec(time.Now().Unix())

	// important: these values cannot be trusted from the API
	// set them on the server side based on token / internal clock
	startMovement := &startMovementEvent{
		MovementID:     tMovementID(m.id),
		PlayerID:       tPlayerID(playerID),
		OriginID:       tCityID(m.originID),
		DestinationID:  tCityID(m.destinationID),
		DestinationX:   m.destinationX,
		DestinationY:   m.destinationY,
		DepartureEpoch: serverSideEpoch,
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
	serverSideEpoch := tSec(time.Now().Unix())

	queueItem := queueUnitEvent{
		UnitQueueItemID: tUnitQueueItemID(item.id),
		CityID:          tCityID(item.cityID),
		PlayerID:        tPlayerID(playerID),
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
	serverSideEpoch := tSec(time.Now().Unix())

	queueItem := queueBuildingEvent{
		BuildingQueueItemID: item.id,
		CityID:              tCityID(item.cityID),
		PlayerID:            tPlayerID(playerID),
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

func (s *inserterService) CreateCity(ctx context.Context, playerID string, c *city) error {
	serverSideEpoch := tSec(time.Now().Unix())

	createCity := createCityEvent{
		CityID:    tCityID(c.id),
		Name:      c.name,
		PlayerID:  tPlayerID(playerID),
		LocationX: c.locationX,
		LocationY: c.locationY,
		// NOTE: when inserted from outside means it was first city creation
		// this comes with no resources and units - you build from zero. Only
		// founded cities later in the game generate events with these two
		// fields populated.
		ResourceCount: make(map[tResourceName]tResourceCount),
		UnitCount:     make(map[tUnitName]tUnitCount),
	}
	payload, err := json.Marshal(createCity)
	if err != nil {
		return err
	}

	eventID := uuid.NewString()
	e := &event{
		id:      eventID,
		name:    createCityEventName,
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

func (s *inserterService) DeleteCity(ctx context.Context, playerID, cityID string) error {
	serverSideEpoch := tSec(time.Now().Unix())

	deleteCity := deleteCityEvent{
		PlayerID: tPlayerID(playerID),
		CityID:   tCityID(cityID),
	}
	payload, err := json.Marshal(deleteCity)
	if err != nil {
		return err
	}

	eventID := uuid.NewString()
	e := &event{
		id:      eventID,
		name:    createCityEventName,
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
