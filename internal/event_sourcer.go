package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	startMovementEventName   = "startmovement"
	arrivalMovementEventName = "arrival"
	queueUnitEventName       = "queueunit"
	createUnitEventName      = "createunit"
	queueBuildingEventName   = "queuebuilding"
	upgradeBuildingEventName = "upgradebuilding"
	createCityEventName      = "createcity"
)

var (
	// thrown when the event is correctly built but it is invalid to process it
	// e.g., try to move more troops than you own
	// if you're doing a batch processing of events just skip over this one
	errPreConditionFailed = fmt.Errorf("pre-condition failed")
)

// Inserted when a player starts a movement.
// Its processing generates an arrivalMovementEvent at a later epoch (calculated based on distance between cities / speed).
type startMovementEvent struct {
	MovementID     tMovementID    `json:"movementID"`
	PlayerID       tPlayerID      `json:"playerID"`
	OriginID       tCityID        `json:"originID"`
	DestinationID  tCityID        `json:"destinationID"`
	DestinationX   int32          `json:"destinationX"`
	DestinationY   int32          `json:"destinationY"`
	DepartureEpoch tEpoch         `json:"departureEpoch"`
	UnitCount      tUnitCount     `json:"unitCount"`
	ResourceCount  tResourceCount `json:"resourceCount"`
}

// Inserted when a startMovementEvent is processed.
// The arrival processing does the options:
// * reinforce if it's from the same player;
// * battle if it's from separate player
// * forage if it's abandoned
// Then, based on survival (if any), if the current city is not owned by the player (i.e., it was not a reinforce), re-calculate a startMovement event and schedule it.
type arrivalMovementEvent struct {
	MovementID    tMovementID    `json:"movementID"`
	PlayerID      tPlayerID      `json:"playerID"`
	OriginID      tCityID        `json:"originID"`
	DestinationID tCityID        `json:"destinationID"`
	DestinationX  int32          `json:"destinationX"`
	DestinationY  int32          `json:"destinationY"`
	UnitCount     tUnitCount     `json:"unitCount"`
	ResourceCount tResourceCount `json:"resourceCount"`
}

// Inserted when a request to queue a unit is done.
// If the resources of the city are enough an event is set for the future
// to create the units.
type queueUnitEvent struct {
	UnitQueueItemID tItemID   `json:"unitQueueItemID"`
	CityID          tCityID   `json:"cityID"`
	PlayerID        tPlayerID `json:"playerID"`
	UnitCount       int64     `json:"unitCount"`
	UnitType        tUnitName `json:"unitType"`
}

// Inserted when a queueUnitEvent is processed.
// If the cityID still belongs to the original player, new units will be added to the city.
// If the city was conquered by a different player, nothing will happen.
type createUnitEvent struct {
	UnitQueueItemID tItemID   `json:"unitQueueItemID"`
	CityID          tCityID   `json:"cityID"`
	PlayerID        tPlayerID `json:"playerID"`
	UnitCount       int64     `json:"unitCount"`
	UnitType        tUnitName `json:"unitType"`
}

// Inserted when a request to queue a building upgrade is done.
type queueBuildingEvent struct {
	BuildingQueueItemID tItemID       `json:"buildingQueueItemID"`
	CityID              tCityID       `json:"cityID"`
	PlayerID            tPlayerID     `json:"playerID"`
	TargetLevel         int64         `json:"targetLevel"`
	TargetBuilding      tBuildingName `json:"targetBuilding"`
}

// Inserted when a queueBuildingEvent is processed.
// If the cityID still belongs to the original player, the building is upgraded.
// If the city was conquered by a different player, nothing will happen.
type upgradeBuildingEvent struct {
	BuildingQueueItemID tItemID       `json:"buildingQueueItemID"`
	CityID              tCityID       `json:"cityID"`
	PlayerID            tPlayerID     `json:"playerID"`
	TargetLevel         int64         `json:"targetLevel"`
	TargetBuilding      tBuildingName `json:"targetBuilding"`
}

// Inserted when a city is founded.
// If the current location was already occupied in the mean-time, the handling generates an arrival event.
// If the current location was free, the city is created and all units/resources are added to the newly
// created city.
type createCityEvent struct {
	CityID        tCityID        `json:"cityID"`
	Name          string         `json:"name"`
	PlayerID      tPlayerID      `json:"playerID"`
	LocationX     int32          `json:"locationX"`
	LocationY     int32          `json:"locationY"`
	ResourceCount tResourceCount `json:"resourceCount"`
	UnitCount     tUnitCount     `json:"unitCount"`
}

type eventsRepository interface {
	InsertEvent(ctx context.Context, e *event) error
	ListEvents(ctx context.Context, untilEpoch int64) ([]*event, error)
	UpsertMovement(ctx context.Context, m *dbMovement) error
	UpsertCity(ctx context.Context, m *dbCity) error
	UpsertUnitQueueItem(ctx context.Context, m *dbUnitQueueItem) error
	UpsertBuildingQueueItem(ctx context.Context, m *dbBuildingQueueItem) error
	DeleteMovement(ctx context.Context, id string) error
	DeleteCity(ctx context.Context, id string) error
	DeleteUnitQueueItem(ctx context.Context, id string) error
	DeleteBuildingQueueItem(ctx context.Context, id string) error
}

type upsertIDs struct {
	cities    map[tCityID]struct{}
	movements map[tMovementID]struct{}
	unitQ     map[tCityID]map[tItemID]struct{}
	buildingQ map[tCityID]map[tItemID]struct{}
}

// The EventSourcer is the magic of this game.
// It will hold an in-memory state of the game for quick calculations and ordered event processing,
// but will trigger re-sync periods when the whole event log is re-processed to ensure the consistent
// state.
// The in memory state will hold:
//   - a map of city IDs to full city descriptions
//   - a map of movement IDs to full movement descriptions
//   - a map of city IDs to a map of unit queue itmes
//   - a map of city IDs to a map of building queue itmes
//
// Any of the event processors will do:
//   - event payload parsing
//   - event content validation and re-calculation of current state
//   - chain event creation
//   - upsert on cached view tables
//
// After all events are processed, the actual view tables are updated.
// Chain events are only processed on re-sync: make it so that re-sync happens often enough.
type EventSourcer struct {
	repository eventsRepository

	inMemoryStateLock     *sync.Mutex
	cityList              map[tCityID]*city
	movementList          map[tMovementID]*movement
	unitQueuesPerCity     map[tCityID]map[tItemID]*unitQueueItem
	buildingQueuesPerCity map[tCityID]map[tItemID]*buildingQueueItem
	// Goes through a phase of population and deletion while inMemoryStateLock
	// is locked. Utilized to seletively upsert changes to the views.
	toUpsert upsertIDs

	internalEventQueue chan *event
}

func NewEventSourcer(repository eventsRepository) *EventSourcer {
	return &EventSourcer{
		repository:            repository,
		inMemoryStateLock:     &sync.Mutex{},
		internalEventQueue:    make(chan *event, 100),
		cityList:              make(map[tCityID]*city),
		movementList:          make(map[tMovementID]*movement),
		unitQueuesPerCity:     make(map[tCityID]map[tItemID]*unitQueueItem),
		buildingQueuesPerCity: make(map[tCityID]map[tItemID]*buildingQueueItem),
		toUpsert: upsertIDs{
			cities:    make(map[tCityID]struct{}),
			movements: make(map[tMovementID]struct{}),
			unitQ:     make(map[tCityID]map[tItemID]struct{}),
			buildingQ: make(map[tCityID]map[tItemID]struct{}),
		},
	}
}

func (s *EventSourcer) queueEventHandling(e *event) {
	select {
	case s.internalEventQueue <- e:
	default:
		log.Printf("Could not queue event %s, will wait for re-sync to process it.", e.id)
	}
}

func (s *EventSourcer) StartEventsWorker(ctx context.Context, resyncPeriod time.Duration) {
	resyncTimer := time.NewTicker(resyncPeriod)
	defer resyncTimer.Stop()
	for {
		select {
		case e := <-s.internalEventQueue:
			err := s.processEvent(ctx, e)
			if err != nil {
				log.Printf("Process event %s, got: %v", e.id, err)
			}
		case <-ctx.Done():
			return
		}

		select {
		case <-resyncTimer.C:
			err := s.reSyncEvents(ctx)
			if err != nil {
				log.Printf("Re-sync events on %v, got: %v", time.Now(), err)
			}
		default:
		}
	}
}

func (s *EventSourcer) processEvent(ctx context.Context, e *event) error {
	s.inMemoryStateLock.Lock()
	defer s.inMemoryStateLock.Unlock()

	var (
		err error
	)
	switch e.name {
	case startMovementEventName:
		err = s.processStartMovementEvent(ctx, e)
	case arrivalMovementEventName:
		err = s.processArrivalMovementEvent(ctx, e)
	case queueUnitEventName:
		err = s.processQueueUnitEventName(ctx, e)
	case createUnitEventName:
		err = s.processCreateUnitEvent(ctx, e)
	case queueBuildingEventName:
		err = s.processQueueBuildingEvent(ctx, e)
	case upgradeBuildingEventName:
		err = s.processUpgradeBuildingEvent(ctx, e)
	case createCityEventName:
		err = s.processCreateCityEvent(ctx, e)
	}
	if err != nil {
		return err
	}

	// upsert view tables to upsert and clear the maps
	s.upsertViews(ctx)

	return nil
}

func (s *EventSourcer) reSyncEvents(ctx context.Context) error {
	events, err := s.repository.ListEvents(ctx, time.Now().Unix())
	if err != nil {
		return err
	}

	s.inMemoryStateLock.Lock()
	defer s.inMemoryStateLock.Unlock()

	// clear previous state to re-sync to correctly calculated present state
	s.cityList = make(map[tCityID]*city)
	s.movementList = make(map[tMovementID]*movement)
	s.unitQueuesPerCity = make(map[tCityID]map[tItemID]*unitQueueItem)
	s.buildingQueuesPerCity = make(map[tCityID]map[tItemID]*buildingQueueItem)

	var (
		e *event
	)
	for i := 0; i < len(events); i++ {
		e = events[i]
		switch e.name {
		case startMovementEventName:
			err = s.processStartMovementEvent(ctx, e)
		case arrivalMovementEventName:
			err = s.processArrivalMovementEvent(ctx, e)
		case queueUnitEventName:
			err = s.processQueueUnitEventName(ctx, e)
		case createUnitEventName:
			err = s.processCreateUnitEvent(ctx, e)
		case queueBuildingEventName:
			err = s.processQueueBuildingEvent(ctx, e)
		case upgradeBuildingEventName:
			err = s.processUpgradeBuildingEvent(ctx, e)
		case createCityEventName:
			err = s.processCreateCityEvent(ctx, e)
		default:
			err = fmt.Errorf("%w, event %s, reason: %s %s", errPreConditionFailed, e.id, "unkown event name", e.name)
		}
		if err != nil {
			if errors.Is(err, errPreConditionFailed) {
				continue
			}
			return err
		}
	}

	// upsert view tables to upsert and clear the maps
	s.upsertViews(ctx)

	return nil
}

func (s *EventSourcer) upsertViews(ctx context.Context) error {
	for cityID := range s.toUpsert.cities {
		c, ok := s.cityList[cityID]
		if !ok {
			err := s.repository.DeleteCity(ctx, string(cityID))
			if err != nil {
				return err
			}
			// TODO: also delete all view tables queue items
		}
		dbc, err := cityToDBModel(c)
		if err != nil {
			return err
		}
		err = s.repository.UpsertCity(ctx, dbc)
		if err != nil {
			return err
		}
	}
	for movementID := range s.toUpsert.movements {
		m, ok := s.movementList[movementID]
		if !ok {
			err := s.repository.DeleteMovement(ctx, string(movementID))
			if err != nil {
				return err
			}
		}
		dbm, err := movementToDBModel(m)
		if err != nil {
			return err
		}
		err = s.repository.UpsertMovement(ctx, dbm)
		if err != nil {
			return err
		}
	}
	for cityID, unitQ := range s.toUpsert.unitQ {
		for itemID := range unitQ {
			item, ok := s.unitQueuesPerCity[cityID][itemID]
			if !ok {
				err := s.repository.DeleteUnitQueueItem(ctx, string(itemID))
				if err != nil {
					return err
				}
			}
			dbitem := unitQueueItemToDBModel(item)
			err := s.repository.UpsertUnitQueueItem(ctx, dbitem)
			if err != nil {
				return err
			}
		}
	}
	for cityID, buildingQ := range s.toUpsert.buildingQ {
		for itemID := range buildingQ {
			item, ok := s.buildingQueuesPerCity[cityID][itemID]
			if !ok {
				err := s.repository.DeleteBuildingQueueItem(ctx, string(itemID))
				if err != nil {
					return err
				}
			}
			dbitem := buildingQueueItemToDBModel(item)
			err := s.repository.UpsertBuildingQueueItem(ctx, dbitem)
			if err != nil {
				return err
			}
		}
	}
	s.toUpsert = upsertIDs{
		cities:    make(map[tCityID]struct{}),
		movements: make(map[tMovementID]struct{}),
		unitQ:     make(map[tCityID]map[tItemID]struct{}),
		buildingQ: make(map[tCityID]map[tItemID]struct{}),
	}
	return nil
}

func (s *EventSourcer) processStartMovementEvent(ctx context.Context, e *event) error {
	// parsing
	startMovement := startMovementEvent{}
	err := json.Unmarshal([]byte(e.payload), &startMovement)
	if err != nil {
		return err
	}

	// validation and event calculations
	if startMovement.OriginID == startMovement.DestinationID {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "cannot move to the same city")
	}
	err = recalculateResources(int64(startMovement.DepartureEpoch), startMovement.ResourceCount, s.cityList[startMovement.OriginID])
	if err != nil {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, err.Error())
	}
	err = recalculateUnits(startMovement.UnitCount, s.cityList[startMovement.OriginID])
	if err != nil {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, err.Error())
	}
	speed := getMovementSpeed(startMovement.UnitCount)
	distance := dist(
		s.cityList[startMovement.OriginID].locationX,
		s.cityList[startMovement.OriginID].locationY,
		startMovement.DestinationX,
		startMovement.DestinationY,
	)
	travelDurationSec := int64(distance / speed)

	// insert chain events
	arrival := &arrivalMovementEvent{
		MovementID:    startMovement.MovementID,
		PlayerID:      startMovement.PlayerID,
		OriginID:      startMovement.OriginID,
		DestinationID: startMovement.DestinationID,
		DestinationX:  startMovement.DestinationX,
		DestinationY:  startMovement.DestinationY,
		UnitCount:     startMovement.UnitCount,
		ResourceCount: startMovement.ResourceCount,
	}
	payload, err := json.Marshal(arrival)
	if err != nil {
		return err
	}
	chainEvent := &event{
		id:      uuid.NewString(),
		name:    arrivalMovementEventName,
		epoch:   int64(startMovement.DepartureEpoch) + travelDurationSec,
		payload: string(payload),
	}
	err = s.repository.InsertEvent(ctx, chainEvent)
	if err != nil {
		return err
	}

	// upsert cached table and signal future view table upsert
	m := &movement{
		id:             string(startMovement.MovementID),
		playerID:       string(startMovement.PlayerID),
		originID:       string(startMovement.OriginID),
		destinationID:  string(startMovement.DestinationID),
		destinationX:   startMovement.DestinationX,
		destinationY:   startMovement.DestinationY,
		departureEpoch: int64(startMovement.DepartureEpoch),
		speed:          speed,
		resourceCount:  startMovement.ResourceCount,
		unitCount:      startMovement.UnitCount,
	}
	s.movementList[startMovement.MovementID] = m
	s.toUpsert.movements[startMovement.MovementID] = struct{}{}
	s.toUpsert.cities[startMovement.OriginID] = struct{}{}

	return nil
}

func (s *EventSourcer) processArrivalMovementEvent(_ context.Context, e *event) error {
	// parsing
	arrivalMovement := arrivalMovementEvent{}
	err := json.Unmarshal([]byte(e.payload), &arrivalMovement)
	if err != nil {
		return err
	}

	// validation and event calculations
	switch {
	case arrivalMovement.DestinationID == "":
		// TODO:
		// might be have been an empty spot when the movement started
		// check if location X,Y were colonized or not
		// if not - forage random value of resources and go back to originID
	case arrivalMovement.PlayerID == tPlayerID(s.cityList[arrivalMovement.DestinationID].playerID):
		// TODO:
		// we've arrived at a friendly city: reinforce the units and top-up resource bases
	case arrivalMovement.PlayerID != tPlayerID(s.cityList[arrivalMovement.DestinationID].playerID):
		// TODO:
		// calculate battle. if units survive, calculate plunder and chain a start movement event
		// if units all die, don't chain another event, just subtract units from the destination ID
		// TODO: future aliances possibility and treat this as a permanent reinforcement (?)
	}

	// insert chain events
	// upsert cached table and signal future view table upsert

	// TODO: remove previous movement event
	return nil
}

func (s *EventSourcer) processQueueUnitEventName(ctx context.Context, e *event) error {
	// parsing
	queueUnit := queueUnitEvent{}
	err := json.Unmarshal([]byte(e.payload), &queueUnit)
	if err != nil {
		return err
	}

	// validation and event calculations
	err = recalculateResources(e.epoch, readOnlyConfig.Units[queueUnit.UnitType].UnitCost, s.cityList[queueUnit.CityID])
	if err != nil {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, err.Error())
	}
	trainingDurationSec := getTrainingDuration(queueUnit.UnitCount, queueUnit.UnitType, s.cityList[queueUnit.CityID])

	// insert chain events
	createUnit := &createUnitEvent{
		UnitQueueItemID: queueUnit.UnitQueueItemID,
		CityID:          queueUnit.CityID,
		PlayerID:        queueUnit.PlayerID,
		UnitCount:       queueUnit.UnitCount,
		UnitType:        queueUnit.UnitType,
	}
	payload, err := json.Marshal(createUnit)
	if err != nil {
		return err
	}
	chainEvent := &event{
		id:      uuid.NewString(),
		name:    createUnitEventName,
		epoch:   e.epoch + trainingDurationSec,
		payload: string(payload),
	}
	err = s.repository.InsertEvent(ctx, chainEvent)
	if err != nil {
		return err
	}

	// upsert cached table and signal future view table upsert
	queueItem := &unitQueueItem{
		id:          string(queueUnit.UnitQueueItemID),
		cityID:      string(queueUnit.CityID),
		queuedEpoch: e.epoch,
		durationSec: trainingDurationSec,
		unitCount:   queueUnit.UnitCount,
		unitType:    string(queueUnit.UnitType),
	}
	s.unitQueuesPerCity[tCityID(queueItem.cityID)][tItemID(queueItem.id)] = queueItem
	s.toUpsert.unitQ[tCityID(queueItem.cityID)][tItemID(queueItem.id)] = struct{}{}

	return nil
}

func (s *EventSourcer) processCreateUnitEvent(_ context.Context, e *event) error {
	// parsing
	createUnit := createUnitEvent{}
	err := json.Unmarshal([]byte(e.payload), &createUnit)
	if err != nil {
		return err
	}
	// validation and event calculations
	if createUnit.PlayerID != tPlayerID(s.cityList[createUnit.CityID].playerID) {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "city has changed owner")
	}
	s.cityList[createUnit.CityID].unitCount[string(createUnit.UnitType)] += int64(createUnit.UnitCount)
	delete(s.unitQueuesPerCity[createUnit.CityID], createUnit.UnitQueueItemID)

	// insert chain events

	// upsert cached table and signal future view table upsert
	s.toUpsert.cities[createUnit.CityID] = struct{}{}
	s.toUpsert.unitQ[createUnit.CityID][createUnit.UnitQueueItemID] = struct{}{}

	return nil
}

func (s *EventSourcer) processQueueBuildingEvent(ctx context.Context, e *event) error {
	// parsing
	queueBuilding := queueBuildingEvent{}
	err := json.Unmarshal([]byte(e.payload), &queueBuilding)
	if err != nil {
		return err
	}

	// validation and event calculations
	targetBuildingSpecs := readOnlyConfig.Buildings[queueBuilding.TargetBuilding]
	if queueBuilding.TargetLevel > int64(targetBuildingSpecs.MaxLevel) {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "cannot upgrade past max level")
	}
	currentBuildingLevel := s.cityList[queueBuilding.CityID].buildingsLevel[string(queueBuilding.TargetBuilding)]
	if queueBuilding.TargetLevel > currentBuildingLevel+1 {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "only upgrade 1 level at a time")
	}
	upgradeCost := targetBuildingSpecs.UpgradeCost[currentBuildingLevel]
	err = recalculateResources(e.epoch, upgradeCost, s.cityList[queueBuilding.CityID])
	if err != nil {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, err.Error())
	}
	upgradeDurationSec := targetBuildingSpecs.UpgradeSpeed[currentBuildingLevel]

	// insert chain events
	upgradeBuilding := &upgradeBuildingEvent{
		BuildingQueueItemID: queueBuilding.BuildingQueueItemID,
		CityID:              queueBuilding.CityID,
		PlayerID:            queueBuilding.PlayerID,
		TargetLevel:         queueBuilding.TargetLevel,
		TargetBuilding:      queueBuilding.TargetBuilding,
	}
	payload, err := json.Marshal(upgradeBuilding)
	if err != nil {
		return err
	}
	chainEvent := &event{
		id:      uuid.NewString(),
		name:    upgradeBuildingEventName,
		epoch:   e.epoch + upgradeDurationSec,
		payload: string(payload),
	}
	err = s.repository.InsertEvent(ctx, chainEvent)
	if err != nil {
		return err
	}

	// upsert cached table and signal future view table upsert
	queueItem := &buildingQueueItem{
		id:             string(queueBuilding.BuildingQueueItemID),
		cityID:         string(queueBuilding.CityID),
		queuedEpoch:    e.epoch,
		durationSec:    upgradeDurationSec,
		targetLevel:    queueBuilding.TargetLevel,
		targetBuilding: string(queueBuilding.TargetBuilding),
	}
	s.buildingQueuesPerCity[tCityID(queueItem.cityID)][tItemID(queueItem.id)] = queueItem
	s.toUpsert.buildingQ[tCityID(queueItem.cityID)][tItemID(queueItem.id)] = struct{}{}

	return nil
}

func (s *EventSourcer) processUpgradeBuildingEvent(_ context.Context, e *event) error {
	// parsing
	upgradeBuilding := upgradeBuildingEvent{}
	err := json.Unmarshal([]byte(e.payload), &upgradeBuilding)
	if err != nil {
		return err
	}

	// validation and event calculations
	// HACK: pass a zero cost event to re-calculate the base and increment the epoch
	err = recalculateResources(int64(e.epoch), make(tResourceCount), s.cityList[upgradeBuilding.CityID])
	if err != nil {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, err.Error())
	}
	s.cityList[upgradeBuilding.CityID].buildingsLevel[string(upgradeBuilding.TargetBuilding)] = int64(upgradeBuilding.TargetLevel)
	delete(s.buildingQueuesPerCity[upgradeBuilding.CityID], upgradeBuilding.BuildingQueueItemID)

	// insert chain events

	// upsert cached table and signal future view table upsert
	s.toUpsert.cities[upgradeBuilding.CityID] = struct{}{}
	s.toUpsert.buildingQ[upgradeBuilding.CityID][upgradeBuilding.BuildingQueueItemID] = struct{}{}

	return nil
}

func (s *EventSourcer) processCreateCityEvent(_ context.Context, e *event) error {
	// parsing
	createCity := createCityEvent{}
	err := json.Unmarshal([]byte(e.payload), &createCity)
	if err != nil {
		return err
	}

	// validation and event calculations
	if _, ok := s.cityList[createCity.CityID]; ok {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "unexpected repeated cityID")
	}
	// TODO: check if a city was present in the location before, if so, create a war!
	s.cityList[createCity.CityID] = &city{
		id:        string(createCity.CityID),
		name:      createCity.Name,
		playerID:  string(createCity.PlayerID),
		locationX: createCity.LocationX,
		locationY: createCity.LocationY,
		// NOTE: all buildings start at level 0 in a city creation.
		// Go default map values make it so that future calculations of
		// level up are included.
		buildingsLevel: make(map[string]int64),
		resourceBase:   createCity.ResourceCount,
		resourceEpoch:  e.epoch,
		unitCount:      createCity.UnitCount,
	}
	s.buildingQueuesPerCity[createCity.CityID] = make(map[tItemID]*buildingQueueItem)
	s.unitQueuesPerCity[createCity.CityID] = make(map[tItemID]*unitQueueItem)

	// insert chain events

	// upsert cached table and signal future view table upsert
	s.toUpsert.cities[tCityID(createCity.CityID)] = struct{}{}

	return nil
}

func recalculateResources(epoch int64, cost tResourceCount, c *city) error {
	missingResources := ""
	for resourceName, resourceCost := range cost {
		if c.resourceBase[resourceName] > resourceCost {
			continue
		}
		// TODO: measure and optimize
		multiplier := 1.0
		for _, buildingKey := range readOnlyResourceMultipliers[tResourceName(resourceName)] {
			// TODO: formalize these equations to calculate game time
			multiplier *= readOnlyConfig.Buildings[buildingKey].ResourceMultiplier[c.buildingsLevel[string(buildingKey)]]
		}
		currentResources := int64(float64(c.resourceBase[resourceName]) +
			float64(epoch-c.resourceEpoch)*
				float64(readOnlyConfig.ResourceTrickles[tResourceName(resourceName)])*multiplier)
		if currentResources > resourceCost {
			continue
		}
		missingResources += fmt.Sprintf("missing %s resources", resourceName)
	}
	if missingResources != "" {
		return fmt.Errorf(missingResources)
	}

	c.resourceEpoch = epoch
	for resourceName := range c.resourceBase {
		multiplier := 1.0
		for _, buildingKey := range readOnlyResourceMultipliers[tResourceName(resourceName)] {
			// TODO: formalize these equations to calculate game time
			multiplier *= readOnlyConfig.Buildings[buildingKey].ResourceMultiplier[c.buildingsLevel[string(buildingKey)]]
		}
		currentResources := int64(float64(c.resourceBase[resourceName]) +
			float64(epoch-c.resourceEpoch)*
				float64(readOnlyConfig.ResourceTrickles[tResourceName(resourceName)])*multiplier)
		c.resourceBase[resourceName] = currentResources - cost[resourceName]
	}
	return nil
}

func recalculateUnits(cost tUnitCount, c *city) error {
	insufficientUnits := ""
	for unitName, unitCount := range cost {
		if c.unitCount[unitName] > unitCount {
			continue
		}
		insufficientUnits += fmt.Sprintf("missing %s units", unitName)
	}
	if insufficientUnits != "" {
		return fmt.Errorf(insufficientUnits)
	}
	for unitName, unitCount := range cost {
		c.unitCount[unitName] -= unitCount
	}
	return nil
}

func getMovementSpeed(unitCount tUnitCount) float64 {
	for _, unitName := range readOnlySlowestUnits {
		if unitCount[string(unitName)] > 0 {
			return readOnlyConfig.Units[unitName].UnitSpeed
		}
	}
	// NOTE: this should not happen, as movements would only happen with pre-existing units
	return 1.0
}

func getTrainingDuration(unitCount int64, unitName tUnitName, c *city) int64 {
	// TODO: measure and optimize
	multiplier := 1.0
	for _, buildingKey := range readOnlyTrainingMultipliers[unitName] {
		// TODO: formalize these equations to calculate game time
		multiplier *= readOnlyConfig.Buildings[buildingKey].TrainingMultiplier[c.buildingsLevel[string(buildingKey)]]
	}
	return int64(float64(readOnlyConfig.Units[unitName].UnitProductionSpeedSec*unitCount) * multiplier)
}

func dist(x1, y1, x2, y2 int32) float64 {
	// TODO: measure and optimize
	squareDist := (x1-x2)*(x1-x2) + (y1-y2)*(y1-y2)
	if squareDist <= 0 {
		return 0.0
	}
	return math.Sqrt(float64(squareDist))
}
