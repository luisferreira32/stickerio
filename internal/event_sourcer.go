package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	startMovementEventName   = "startmovement"
	arrivalMovementEventName = "arrival"
	queueUnitEventName       = "queueunit"
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
	UnitCount     tUnitCount     `json:"unitCount"`
	ResourceCount tResourceCount `json:"resourceCount"`
}

// Inserted when a request to queue a unit is done.
type queueUnitEvent struct {
	UnitQueueItemID string `json:"unitQueueItemID"`
	CityID          string `json:"cityID"`
	PlayerID        string `json:"playerID"`
	QueuedEpoch     int64  `json:"queuedEpoch"`
	UnitCount       int32  `json:"unitCount"`
	UnitType        string `json:"unitType"`
}

type eventsRepository interface {
	InsertEvent(ctx context.Context, e *event) error
	ListEvents(ctx context.Context, untilEpoch int64) ([]*event, error)
	UpsertMovement(ctx context.Context, m *dbMovement) error
	UpsertCity(ctx context.Context, m *dbCity) error
	UpsertUnitQueueItem(ctx context.Context, m *dbUnitQueueItem) error
	UpsertBuildingQueueItem(ctx context.Context, m *dbBuildingQueueItem) error
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
//   - event content validation
//   - chain event creation
//   - upsert on view tables
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
		repository:         repository,
		inMemoryStateLock:  &sync.Mutex{},
		internalEventQueue: make(chan *event, 100),
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

	switch e.name {
	case startMovementEventName:
		err := s.processStartMovementEvent(ctx, e)
		if err != nil {
			return err
		}
	case arrivalMovementEventName:
	case queueUnitEventName:
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

	var (
		e *event
	)
	for i := 0; i < len(events); i++ {
		e = events[i]
		switch e.name {
		case startMovementEventName:
			err := s.processStartMovementEvent(ctx, e)
			if err != nil {
				if errors.Is(err, errPreConditionFailed) {
					continue
				}
				return err
			}
		case arrivalMovementEventName:
		case queueUnitEventName:
		}
	}

	// upsert view tables to upsert and clear the maps
	s.upsertViews(ctx)

	return nil
}

func (s *EventSourcer) upsertViews(ctx context.Context) error {
	for cityID := range s.toUpsert.cities {
		dbc, err := cityToDBModel(s.cityList[cityID])
		if err != nil {
			return err
		}
		err = s.repository.UpsertCity(ctx, dbc)
		if err != nil {
			return err
		}
	}
	for movementID := range s.toUpsert.movements {
		dbm, err := movementToDBModel(s.movementList[movementID])
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
			dbitem := unitQueueItemToDBModel(s.unitQueuesPerCity[cityID][itemID])
			err := s.repository.UpsertUnitQueueItem(ctx, dbitem)
			if err != nil {
				return err
			}
		}
	}
	for cityID, unitQ := range s.toUpsert.buildingQ {
		for itemID := range unitQ {
			dbitem := buildingQueueItemToDBModel(s.buildingQueuesPerCity[cityID][itemID])
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
	epoch := time.Now().Unix()

	// validation
	if startMovement.OriginID == startMovement.DestinationID {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "cannot move to the same city")
	}
	err = recalculateResources(epoch, startMovement.ResourceCount, s.cityList[startMovement.OriginID])
	if err != nil {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, err.Error())
	}
	err = recalculateUnits(startMovement.UnitCount, s.cityList[startMovement.OriginID])
	if err != nil {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, err.Error())
	}

	// insert new event
	arrival := &arrivalMovementEvent{
		MovementID:    startMovement.MovementID,
		PlayerID:      startMovement.PlayerID,
		OriginID:      startMovement.OriginID,
		DestinationID: startMovement.DestinationID,
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
		epoch:   time.Now().Unix(),
		payload: string(payload),
	}
	s.queueEventHandling(chainEvent)

	// upsert cached table and signal future view table upsert
	m := &movement{
		id:             string(startMovement.MovementID),
		playerID:       string(startMovement.PlayerID),
		originID:       string(startMovement.OriginID),
		destinationID:  string(startMovement.DestinationID),
		departureEpoch: epoch,
		speed:          getMovementSpeed(startMovement.UnitCount),
		resourceCount:  startMovement.ResourceCount,
		unitCount:      startMovement.UnitCount,
	}
	s.movementList[startMovement.MovementID] = m
	s.toUpsert.movements[startMovement.MovementID] = struct{}{}
	s.toUpsert.cities[startMovement.OriginID] = struct{}{}

	return nil
}

func (s *EventSourcer) processArrivalMovementEvent(ctx context.Context, e *event) error {
	// parsing
	// validation
	// insert chain events
	// upsert cached table and signal future view table upsert
	return nil
}

func (s *EventSourcer) processQueueUnitEventName(ctx context.Context, e *event) error {
	// parsing
	// validation
	// insert chain events
	// upsert cached table and signal future view table upsert
	return nil
}

func recalculateResources(epoch int64, cost tResourceCount, c *city) error {
	missingResources := ""
	for resourceName, resourceCost := range cost {
		if c.resourceBase[resourceName] > resourceCost {
			continue
		}
		// TODO: more efficient than this
		currentResources := int64(float32(c.resourceBase[resourceName]) +
			float32(epoch-c.resourceEpoch)*
				float32(config.ResourceTrickles[tResourceName(resourceName)])*
				config.EconomicBuildings[tBuildingName("mines")].Multiplier[c.mineLevel])
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
		currentResources := int64(float32(c.resourceBase[resourceName]) +
			float32(epoch-c.resourceEpoch)*
				float32(config.ResourceTrickles[tResourceName(resourceName)])*
				config.EconomicBuildings[tBuildingName("mines")].Multiplier[c.mineLevel])
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

func getMovementSpeed(unitCount tUnitCount) float32 {
	for _, unitName := range slowestUnits {
		if unitCount[string(unitName)] > 0 {
			return config.Units[unitName].UnitSpeed
		}
	}
	// NOTE: this should not happen, as movements would only happen with pre-existing units
	return 1.0
}
