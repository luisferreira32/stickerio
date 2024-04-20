package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	startMovementEventName   = "startmovement"
	arrivalMovementEventName = "arrival"
	returnMovementEventName  = "returnmovement"
	queueUnitEventName       = "queueunit"
	createUnitEventName      = "createunit"
	queueBuildingEventName   = "queuebuilding"
	upgradeBuildingEventName = "upgradebuilding"
	createCityEventName      = "createcity"
	deleteCityEventName      = "deletecity"
)

var (
	// thrown when the event is correctly built but it is invalid to process it
	// e.g., try to move more troops than you own
	// if you're doing a batch processing of events just skip over this one
	errPreConditionFailed = fmt.Errorf("pre-condition failed")
)

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
	DeleteUnitQueueItemsFromCity(ctx context.Context, cityID string) error
	DeleteBuildingQueueItem(ctx context.Context, id string) error
	DeleteBuildingQueueItemsFromCity(ctx context.Context, cityID string) error
}

type upsertIDs struct {
	cities    map[tCityID]struct{}
	movements map[tMovementID]struct{}
	unitQ     map[tCityID]map[tUnitQueueItemID]struct{}
	buildingQ map[tCityID]map[tBuildingQueueItemID]struct{}
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

	inMemoryStateLock *sync.Mutex
	inMemoryState     *inMemoryStorage
	// Goes through a phase of population and deletion while inMemoryStateLock
	// is locked. Utilized to seletively upsert changes to the views.
	toUpsert upsertIDs

	internalEventQueue chan *event
}

func NewEventSourcer(repository eventsRepository) *EventSourcer {
	inMemoryState := &inMemoryStorage{}
	inMemoryState.clear()
	return &EventSourcer{
		repository:         repository,
		inMemoryStateLock:  &sync.Mutex{},
		internalEventQueue: make(chan *event, 100),
		inMemoryState:      inMemoryState,
		toUpsert: upsertIDs{
			cities:    make(map[tCityID]struct{}),
			movements: make(map[tMovementID]struct{}),
			unitQ:     make(map[tCityID]map[tUnitQueueItemID]struct{}),
			buildingQ: make(map[tCityID]map[tBuildingQueueItemID]struct{}),
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
	case returnMovementEventName:
		err = s.processReturnMovementEvent(ctx, e)
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
	case deleteCityEventName:
		err = s.processDeleteCityEvent(ctx, e)
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
	s.inMemoryState.clear()

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
		case returnMovementEventName:
			err = s.processReturnMovementEvent(ctx, e)
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
		case deleteCityEventName:
			err = s.processDeleteCityEvent(ctx, e)
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
		c, ok := s.inMemoryState.cityList[cityID]
		if !ok {
			err := s.repository.DeleteCity(ctx, string(cityID))
			if err != nil {
				return err
			}
			err = s.repository.DeleteBuildingQueueItemsFromCity(ctx, string(cityID))
			if err != nil {
				return err
			}
			err = s.repository.DeleteUnitQueueItemsFromCity(ctx, string(cityID))
			if err != nil {
				return err
			}
			continue
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
		m, ok := s.inMemoryState.movementList[movementID]
		if !ok {
			err := s.repository.DeleteMovement(ctx, string(movementID))
			if err != nil {
				return err
			}
			continue
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
			item, ok := s.inMemoryState.unitQueuesPerCity[cityID][itemID]
			if !ok {
				err := s.repository.DeleteUnitQueueItem(ctx, string(itemID))
				if err != nil {
					return err
				}
				continue
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
			item, ok := s.inMemoryState.buildingQueuesPerCity[cityID][itemID]
			if !ok {
				err := s.repository.DeleteBuildingQueueItem(ctx, string(itemID))
				if err != nil {
					return err
				}
				continue
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
		unitQ:     make(map[tCityID]map[tUnitQueueItemID]struct{}),
		buildingQ: make(map[tCityID]map[tBuildingQueueItemID]struct{}),
	}
	return nil
}

// Processing generates an arrivalMovementEvent at a later epoch (calculated based on distance between cities / speed).
func (s *EventSourcer) processStartMovementEvent(ctx context.Context, e *event) error {
	// parsing
	startMovement := startMovementEvent{}
	err := json.Unmarshal([]byte(e.payload), &startMovement)
	if err != nil {
		return err
	}

	// validation and event calculations
	if s.inMemoryState.cityList[startMovement.OriginID].playerID != startMovement.PlayerID {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "cannot alter cities of other players")
	}
	if startMovement.OriginID == startMovement.DestinationID {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "cannot move to the same city")
	}
	err = reCityCalculateResources(startMovement.DepartureEpoch, startMovement.ResourceCount, s.inMemoryState.cityList[startMovement.OriginID])
	if err != nil {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, err.Error())
	}

	insufficientUnits := ""
	for unitName, unitCount := range startMovement.UnitCount {
		if s.inMemoryState.cityList[startMovement.OriginID].unitCount[unitName] > unitCount {
			continue
		}
		insufficientUnits += fmt.Sprintf("missing %s units", unitName)
	}
	if insufficientUnits != "" {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, insufficientUnits)
	}
	for unitName, unitCount := range startMovement.UnitCount {
		s.inMemoryState.cityList[startMovement.OriginID].unitCount[unitName] -= unitCount
	}
	speed := getGroupMovementSpeed(startMovement.UnitCount)
	travelDurationSec := travelTime(
		s.inMemoryState.cityList[startMovement.OriginID].locationX,
		s.inMemoryState.cityList[startMovement.OriginID].locationY,
		startMovement.DestinationX,
		startMovement.DestinationY,
		speed,
	)

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
		epoch:   startMovement.DepartureEpoch + travelDurationSec,
		payload: string(payload),
	}
	err = s.repository.InsertEvent(ctx, chainEvent)
	if err != nil {
		return err
	}

	// upsert cached table and signal future view table upsert
	m := &movement{
		id:             startMovement.MovementID,
		playerID:       startMovement.PlayerID,
		originID:       startMovement.OriginID,
		destinationID:  startMovement.DestinationID,
		destinationX:   startMovement.DestinationX,
		destinationY:   startMovement.DestinationY,
		departureEpoch: startMovement.DepartureEpoch,
		speed:          speed,
		resourceCount:  startMovement.ResourceCount,
		unitCount:      startMovement.UnitCount,
	}
	s.inMemoryState.movementList[startMovement.MovementID] = m
	s.toUpsert.movements[startMovement.MovementID] = struct{}{}
	s.toUpsert.cities[startMovement.OriginID] = struct{}{}

	return nil
}

// The arrival processing does the options:
// * reinforce if it's from the same player;
// * battle if it's from separate player and insert returnMovementEvent if troops survive;
// * forage if it's abandoned and insert returnMovementEvent;
// * create a new city if the unit type sent has the capability for settling;
func (s *EventSourcer) processArrivalMovementEvent(ctx context.Context, e *event) error {
	// parsing
	arrivalMovement := arrivalMovementEvent{}
	err := json.Unmarshal([]byte(e.payload), &arrivalMovement)
	if err != nil {
		return err
	}

	// validation and event calculations
	destinationID := arrivalMovement.DestinationID
	if c := s.inMemoryState.getCityByLocation(arrivalMovement.DestinationX, arrivalMovement.DestinationY); c != nil {
		destinationID = tCityID(c.id)
	}

	switch {
	case destinationID == "":
		// TODO ensure this does not overflow or avoid int64 for resource calculations (re-type it)
		var (
			freeCarryCapacity tResourceCount
		)
		for unitType, unitCount := range arrivalMovement.UnitCount {
			freeCarryCapacity += tResourceCount(unitCount) * cfg.Units[tUnitName(unitType)].CarryCapacity
		}
		for _, resourceCount := range arrivalMovement.ResourceCount {
			freeCarryCapacity -= resourceCount
		}
		if freeCarryCapacity > 1 {
			foragableResources := tResourceCount(rand.Float64() * cfg.ForagingCoefficient * float64(freeCarryCapacity))
			// FIXME: maps are not ordered, but is it random enough?
			for resourceName, resourceCount := range arrivalMovement.ResourceCount {
				foragedResource := tResourceCount(rand.Int63n(int64(foragableResources)))
				arrivalMovement.ResourceCount[resourceName] = resourceCount + foragedResource
				foragableResources -= foragedResource
			}
		}

		originCity := s.inMemoryState.cityList[arrivalMovement.OriginID]

		speed := getGroupMovementSpeed(arrivalMovement.UnitCount)
		travelDurationSec := travelTime(
			arrivalMovement.DestinationY,
			arrivalMovement.DestinationY,
			originCity.locationX,
			originCity.locationY,
			speed,
		)

		s.inMemoryState.movementList[arrivalMovement.MovementID].originID = destinationID
		s.inMemoryState.movementList[arrivalMovement.MovementID].destinationID = arrivalMovement.OriginID
		s.inMemoryState.movementList[arrivalMovement.MovementID].destinationX = originCity.locationX
		s.inMemoryState.movementList[arrivalMovement.MovementID].destinationY = originCity.locationY
		s.inMemoryState.movementList[arrivalMovement.MovementID].resourceCount = arrivalMovement.ResourceCount

		// insert chain events
		returnMovement := &returnMovementEvent{
			MovementID: arrivalMovement.MovementID,
			PlayerID:   arrivalMovement.PlayerID,
			// switch origin and destination
			OriginID:      arrivalMovement.DestinationID,
			DestinationID: arrivalMovement.OriginID,
			DestinationX:  originCity.locationX,
			DestinationY:  originCity.locationY,
			UnitCount:     arrivalMovement.UnitCount,
			ResourceCount: arrivalMovement.ResourceCount,
		}
		payload, err := json.Marshal(returnMovement)
		if err != nil {
			return err
		}
		chainEvent := &event{
			id:      uuid.NewString(),
			name:    returnMovementEventName,
			epoch:   e.epoch + travelDurationSec,
			payload: string(payload),
		}
		err = s.repository.InsertEvent(ctx, chainEvent)
		if err != nil {
			return err
		}

	case arrivalMovement.PlayerID == tPlayerID(s.inMemoryState.cityList[destinationID].playerID):
		for resourceName, resourceTransported := range arrivalMovement.ResourceCount {
			s.inMemoryState.cityList[destinationID].resourceBase[resourceName] += resourceTransported
		}
		for unitName, reinforcementCount := range arrivalMovement.UnitCount {
			s.inMemoryState.cityList[destinationID].unitCount[unitName] += reinforcementCount
		}
		// upsert cached table and signal future view table upsert
		delete(s.inMemoryState.movementList, arrivalMovement.MovementID)
		s.toUpsert.cities[destinationID] = struct{}{}

	case arrivalMovement.PlayerID != tPlayerID(s.inMemoryState.cityList[destinationID].playerID):
		// TODO: future aliances possibility and treat this as a permanent reinforcement instead of an attack (?)

		// TODO: a more balanced battle system (research how it is usually done)
		// and use GPU for matrix calculations maybe?
		var (
			swing                       = 0.0
			swingMax     tUnitStatPower = 0
			swingMin     tUnitStatPower = 0
			epoch                       = e.epoch
			attackers                   = arrivalMovement.UnitCount
			initialLoad                 = arrivalMovement.ResourceCount
			defenderCity                = s.inMemoryState.cityList[destinationID]
		)

		attackerStats := make(map[tUnitStatName]tUnitStatPower)
		for unitName, unitCount := range attackers {
			for statName, statValue := range cfg.Units[tUnitName(unitName)].CombatStats {
				attackerStats[statName] += statValue * tUnitStatPower(unitCount)
				swingMax += statValue * tUnitStatPower(unitCount)
			}
		}
		defendersStats := make(map[tUnitStatName]tUnitStatPower)
		for unitName, unitCount := range defenderCity.unitCount {
			for statName, statValue := range cfg.Units[tUnitName(unitName)].CombatStats {
				defendersStats[statName] += statValue * tUnitStatPower(unitCount)
				swingMin -= statValue * tUnitStatPower(unitCount)
			}
		}

		for statName := range attackerStats {
			swing += float64(attackerStats[statName])*(cfg.CombatEfficiency+(1-cfg.CombatEfficiency)*rand.Float64()) - float64(defendersStats[statName])*(cfg.CombatEfficiency+(1-cfg.CombatEfficiency)*rand.Float64())
		}

		normalizedSwing := 0.5 * swing / float64(swingMax-swingMin)
		switch {
		case swingMin == 0:
			// no units lost, due to the fact that there were no defenders
		case swingMax == 0 && swingMin < 0:
			// no combatant units went attacking... so defenders just kill them
			for unitName := range attackers {
				attackers[unitName] = 0
			}
		default:
			for unitName, unitCount := range defenderCity.unitCount {
				defenderCity.unitCount[unitName] = tUnitCount(float64(unitCount) * (.5 - normalizedSwing))
			}
			for unitName, unitCount := range attackers {
				attackers[unitName] = tUnitCount(float64(unitCount) * (.5 + normalizedSwing))
			}
		}
		var (
			attackersFreeCapacity tResourceCount
			liveAttackers         bool
		)
		for unitType, unitCount := range attackers {
			attackersFreeCapacity += tResourceCount(unitCount) * cfg.Units[tUnitName(unitType)].CarryCapacity
			if unitCount == 0 {
				continue
			}
			liveAttackers = true
		}
		s.toUpsert.cities[destinationID] = struct{}{} // upsert attacked city regardless
		if !liveAttackers {
			return nil
		}

		for _, resourceCount := range initialLoad {
			attackersFreeCapacity -= resourceCount
		}

		// TODO: do not utilize this hack to make it re-calculate the epoch and current base
		reCityCalculateResources(epoch, make(map[tResourceName]tResourceCount), defenderCity)

		if attackersFreeCapacity < 0 {
			// edge case: attackers bring resources to the defenders! inverted plunder
			resourcesToLeave := -attackersFreeCapacity / tResourceCount(len(initialLoad))
			negativeCost := make(map[tResourceName]tResourceCount)
			for resourceName, resourceCount := range initialLoad {
				initialLoad[resourceName] = resourceCount - resourcesToLeave
				negativeCost[resourceName] = -resourcesToLeave
			}
			reCityCalculateResources(epoch, negativeCost, defenderCity)
		} else if attackersFreeCapacity > 0 {
			resourcesToPlunderPerType := attackersFreeCapacity / tResourceCount(len(defenderCity.resourceBase))
			for resourceName, resourceCount := range defenderCity.resourceBase {
				if resourceCount >= resourcesToPlunderPerType {
					defenderCity.resourceBase[resourceName] -= resourcesToPlunderPerType
					initialLoad[resourceName] += resourcesToPlunderPerType
					attackersFreeCapacity -= resourcesToPlunderPerType
				} else {
					defenderCity.resourceBase[resourceName] -= resourceCount
					initialLoad[resourceName] += resourceCount
					attackersFreeCapacity -= resourceCount
				}
			}
			// NOTE: do not rely on randomness of map key access for this second round ?
			if attackersFreeCapacity > 0 {
				for resourceName, resourceCount := range defenderCity.resourceBase {
					if resourceCount == 0 || attackersFreeCapacity == 0 {
						continue
					}
					if resourceCount >= attackersFreeCapacity {
						defenderCity.resourceBase[resourceName] -= attackersFreeCapacity
						initialLoad[resourceName] += attackersFreeCapacity
						attackersFreeCapacity -= attackersFreeCapacity
					} else {
						defenderCity.resourceBase[resourceName] -= resourceCount
						initialLoad[resourceName] += resourceCount
						attackersFreeCapacity -= resourceCount
					}
				}
			}
		}

		originCity := s.inMemoryState.cityList[arrivalMovement.OriginID]
		speed := getGroupMovementSpeed(arrivalMovement.UnitCount)
		travelDurationSec := travelTime(
			arrivalMovement.DestinationY,
			arrivalMovement.DestinationY,
			originCity.locationX,
			originCity.locationY,
			speed,
		)

		s.inMemoryState.movementList[arrivalMovement.MovementID].originID = destinationID
		s.inMemoryState.movementList[arrivalMovement.MovementID].destinationID = arrivalMovement.OriginID
		s.inMemoryState.movementList[arrivalMovement.MovementID].destinationX = originCity.locationX
		s.inMemoryState.movementList[arrivalMovement.MovementID].destinationY = originCity.locationY
		s.inMemoryState.movementList[arrivalMovement.MovementID].resourceCount = arrivalMovement.ResourceCount
		s.inMemoryState.movementList[arrivalMovement.MovementID].unitCount = arrivalMovement.UnitCount
		s.inMemoryState.movementList[arrivalMovement.MovementID].speed = speed

		// insert chain events
		returnMovement := &returnMovementEvent{
			MovementID: arrivalMovement.MovementID,
			PlayerID:   arrivalMovement.PlayerID,
			// switch origin and destination
			OriginID:      arrivalMovement.DestinationID,
			DestinationID: arrivalMovement.OriginID,
			DestinationX:  originCity.locationX,
			DestinationY:  originCity.locationY,
			UnitCount:     arrivalMovement.UnitCount,
			ResourceCount: arrivalMovement.ResourceCount,
		}
		payload, err := json.Marshal(returnMovement)
		if err != nil {
			return err
		}
		chainEvent := &event{
			id:      uuid.NewString(),
			name:    returnMovementEventName,
			epoch:   e.epoch + travelDurationSec,
			payload: string(payload),
		}
		err = s.repository.InsertEvent(ctx, chainEvent)
		if err != nil {
			return err
		}
	}

	// no matter if it is deleted or it's a returning movement, the movement will
	// be updated (e.g., switch destinations)
	s.toUpsert.movements[arrivalMovement.MovementID] = struct{}{}

	return nil
}

// Processing will simply reinforce the city where they return to with units/resources.
// If the units return to a city that was conquered - since they would not really expect
// that, they are unfortunately massacred and everything is forever lost.
func (s *EventSourcer) processReturnMovementEvent(ctx context.Context, e *event) error {
	// parsing
	returnMovement := returnMovementEvent{}
	err := json.Unmarshal([]byte(e.payload), &returnMovement)
	if err != nil {
		return err
	}

	// validation and event calculations
	destinationCity := s.inMemoryState.getCityByLocation(returnMovement.DestinationX, returnMovement.DestinationY)
	destinationID := tCityID(destinationCity.id)

	switch {
	case destinationCity == nil:
		// city where the units left from no longer exists... everything in movement will disappear
		delete(s.inMemoryState.movementList, returnMovement.MovementID)
	case destinationCity.playerID != returnMovement.PlayerID:
		// the city was captured by another player, everything is lost to avoid an eternal pendulum of a huge
		// army
		// TODO: don't do this when it is possible to conquer a city
		delete(s.inMemoryState.movementList, returnMovement.MovementID)
	default:
		s.toUpsert.cities[destinationID] = struct{}{} // upsert returned city
		originCity := s.inMemoryState.cityList[returnMovement.OriginID]
		speed := getGroupMovementSpeed(returnMovement.UnitCount)
		travelDurationSec := travelTime(
			returnMovement.DestinationY,
			returnMovement.DestinationY,
			originCity.locationX,
			originCity.locationY,
			speed,
		)

		// insert chain events
		arrival := &arrivalMovementEvent{
			MovementID:    returnMovement.MovementID,
			PlayerID:      returnMovement.PlayerID,
			OriginID:      returnMovement.OriginID,
			DestinationID: returnMovement.DestinationID,
			DestinationX:  returnMovement.DestinationX,
			DestinationY:  returnMovement.DestinationY,
			UnitCount:     returnMovement.UnitCount,
			ResourceCount: returnMovement.ResourceCount,
		}
		payload, err := json.Marshal(arrival)
		if err != nil {
			return err
		}
		chainEvent := &event{
			id:      uuid.NewString(),
			name:    arrivalMovementEventName,
			epoch:   e.epoch + travelDurationSec,
			payload: string(payload),
		}
		err = s.repository.InsertEvent(ctx, chainEvent)
		if err != nil {
			return err
		}
	}
	s.toUpsert.movements[returnMovement.MovementID] = struct{}{}

	return nil
}

// If the resources of the city are enough an event is set for the future
// to create the units.
func (s *EventSourcer) processQueueUnitEventName(ctx context.Context, e *event) error {
	// parsing
	queueUnit := queueUnitEvent{}
	err := json.Unmarshal([]byte(e.payload), &queueUnit)
	if err != nil {
		return err
	}

	// validation and event calculations
	if s.inMemoryState.cityList[queueUnit.CityID].playerID != queueUnit.PlayerID {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "cannot alter cities of other players")
	}
	err = reCityCalculateResources(e.epoch, cfg.Units[queueUnit.UnitType].UnitCost, s.inMemoryState.cityList[queueUnit.CityID])
	if err != nil {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, err.Error())
	}

	// TODO: measure and optimize
	multiplier := 1.0
	for _, buildingKey := range cumulativeTrainingMultipliers[queueUnit.UnitType] {
		// TODO: formalize these equations to calculate game time ++ pre-compute most of this
		multiplier *= cfg.Buildings[buildingKey].TrainingMultiplier[s.inMemoryState.cityList[queueUnit.CityID].buildingsLevel[buildingKey]]
	}
	trainingDurationSec := tSec(float64(cfg.Units[queueUnit.UnitType].UnitProductionSpeedSec*tSec(queueUnit.UnitCount)) * multiplier)

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
		id:          queueUnit.UnitQueueItemID,
		cityID:      queueUnit.CityID,
		queuedEpoch: e.epoch,
		durationSec: trainingDurationSec,
		unitCount:   queueUnit.UnitCount,
		unitType:    queueUnit.UnitType,
	}
	s.inMemoryState.unitQueuesPerCity[tCityID(queueItem.cityID)][tUnitQueueItemID(queueItem.id)] = queueItem
	s.toUpsert.unitQ[tCityID(queueItem.cityID)][tUnitQueueItemID(queueItem.id)] = struct{}{}

	return nil
}

// If the cityID still belongs to the original player, new units will be added to the city.
// If the city was conquered by a different player, nothing will happen.
func (s *EventSourcer) processCreateUnitEvent(_ context.Context, e *event) error {
	// parsing
	createUnit := createUnitEvent{}
	err := json.Unmarshal([]byte(e.payload), &createUnit)
	if err != nil {
		return err
	}
	// validation and event calculations
	if createUnit.PlayerID != tPlayerID(s.inMemoryState.cityList[createUnit.CityID].playerID) {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "city has changed owner")
	}
	s.inMemoryState.cityList[createUnit.CityID].unitCount[createUnit.UnitType] += tUnitCount(createUnit.UnitCount)
	delete(s.inMemoryState.unitQueuesPerCity[createUnit.CityID], createUnit.UnitQueueItemID)

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
	if s.inMemoryState.cityList[queueBuilding.CityID].playerID != queueBuilding.PlayerID {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "cannot alter cities of other players")
	}
	targetBuildingSpecs := cfg.Buildings[queueBuilding.TargetBuilding]
	if queueBuilding.TargetLevel > targetBuildingSpecs.MaxLevel {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "cannot upgrade past max level")
	}
	currentBuildingLevel := s.inMemoryState.cityList[queueBuilding.CityID].buildingsLevel[queueBuilding.TargetBuilding]
	if queueBuilding.TargetLevel > currentBuildingLevel+1 {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "only upgrade 1 level at a time")
	}
	upgradeCost := targetBuildingSpecs.UpgradeCost[currentBuildingLevel]
	err = reCityCalculateResources(e.epoch, upgradeCost, s.inMemoryState.cityList[queueBuilding.CityID])
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
		id:             queueBuilding.BuildingQueueItemID,
		cityID:         queueBuilding.CityID,
		queuedEpoch:    e.epoch,
		durationSec:    upgradeDurationSec,
		targetLevel:    queueBuilding.TargetLevel,
		targetBuilding: queueBuilding.TargetBuilding,
	}
	s.inMemoryState.buildingQueuesPerCity[tCityID(queueItem.cityID)][queueItem.id] = queueItem
	s.toUpsert.buildingQ[tCityID(queueItem.cityID)][queueItem.id] = struct{}{}

	return nil
}

// If the cityID still belongs to the original player, the building is upgraded.
// If the city was conquered by a different player, nothing will happen.
func (s *EventSourcer) processUpgradeBuildingEvent(_ context.Context, e *event) error {
	// parsing
	upgradeBuilding := upgradeBuildingEvent{}
	err := json.Unmarshal([]byte(e.payload), &upgradeBuilding)
	if err != nil {
		return err
	}

	// validation and event calculations
	if s.inMemoryState.cityList[upgradeBuilding.CityID].playerID != upgradeBuilding.PlayerID {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "cannot alter cities of other players")
	}
	// HACK: pass a zero cost event to re-calculate the base and increment the epoch
	err = reCityCalculateResources(e.epoch, make(map[tResourceName]tResourceCount), s.inMemoryState.cityList[upgradeBuilding.CityID])
	if err != nil {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, err.Error())
	}
	s.inMemoryState.cityList[upgradeBuilding.CityID].buildingsLevel[upgradeBuilding.TargetBuilding] = upgradeBuilding.TargetLevel
	delete(s.inMemoryState.buildingQueuesPerCity[upgradeBuilding.CityID], upgradeBuilding.BuildingQueueItemID)

	// insert chain events

	// upsert cached table and signal future view table upsert
	s.toUpsert.cities[upgradeBuilding.CityID] = struct{}{}
	s.toUpsert.buildingQ[upgradeBuilding.CityID][upgradeBuilding.BuildingQueueItemID] = struct{}{}

	return nil
}

// The city is created and all units/resources are added to the newly created city.
// Note that this event might only be created for available locations, an arrival event
// might also generate a createCityEvent if the conditions are met.
func (s *EventSourcer) processCreateCityEvent(_ context.Context, e *event) error {
	// parsing
	createCity := createCityEvent{}
	err := json.Unmarshal([]byte(e.payload), &createCity)
	if err != nil {
		return err
	}

	// validation and event calculations
	if _, ok := s.inMemoryState.cityList[createCity.CityID]; ok {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "unexpected repeated cityID")
	}
	if s.inMemoryState.getCityByLocation(createCity.LocationX, createCity.LocationY) != nil {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "unexpected occupied location")
	}

	// insert chain events

	// upsert cached table and signal future view table upsert
	s.inMemoryState.createCity(createCity.CityID, &city{
		id:        createCity.CityID,
		name:      createCity.Name,
		playerID:  createCity.PlayerID,
		locationX: createCity.LocationX,
		locationY: createCity.LocationY,
		// NOTE: all buildings start at level 0 in a city creation.
		// Go default map values make it so that future calculations of
		// level up are included.
		buildingsLevel: make(map[tBuildingName]tBuildingLevel),
		resourceBase:   createCity.ResourceCount,
		resourceEpoch:  e.epoch,
		unitCount:      createCity.UnitCount,
	})
	s.inMemoryState.buildingQueuesPerCity[createCity.CityID] = make(map[tBuildingQueueItemID]*buildingQueueItem)
	s.inMemoryState.unitQueuesPerCity[createCity.CityID] = make(map[tUnitQueueItemID]*unitQueueItem)
	s.toUpsert.cities[tCityID(createCity.CityID)] = struct{}{}
	return nil
}

// Its processing will effectively delete the city, vanquish the resources,
// raze the buildings, and annihilate all units residing the city.
func (s *EventSourcer) processDeleteCityEvent(_ context.Context, e *event) error {
	// parsing
	deleteCity := deleteCityEvent{}
	err := json.Unmarshal([]byte(e.payload), &deleteCity)
	if err != nil {
		return err
	}

	// validation and event calculations
	if s.inMemoryState.cityList[deleteCity.CityID].playerID != deleteCity.PlayerID {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "cannot delete cities of other players")
	}

	// insert chain events

	// upsert cached table and signal future view table upsert
	s.inMemoryState.deleteCity(deleteCity.CityID)
	s.toUpsert.cities[tCityID(deleteCity.CityID)] = struct{}{}
	return nil
}

func reCityCalculateResources(epoch tSec, cost map[tResourceName]tResourceCount, c *city) error {
	missingResources := ""
	for resourceName, resourceCost := range cost {
		if c.resourceBase[resourceName] > resourceCost {
			continue
		}
		// TODO: measure and optimize
		multiplier := 1.0
		for _, buildingKey := range cumulativeResourceMultipliers[tResourceName(resourceName)] {
			// TODO: formalize these equations to calculate game time
			multiplier *= cfg.Buildings[buildingKey].ResourceMultiplier[c.buildingsLevel[buildingKey]]
		}
		currentResources := tResourceCount(float64(c.resourceBase[resourceName]) +
			float64(epoch-c.resourceEpoch)*
				float64(cfg.ResourceTrickles[tResourceName(resourceName)])*multiplier)
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
		for _, buildingKey := range cumulativeResourceMultipliers[tResourceName(resourceName)] {
			// TODO: formalize these equations to calculate game time
			multiplier *= cfg.Buildings[buildingKey].ResourceMultiplier[c.buildingsLevel[buildingKey]]
		}
		currentResources := tResourceCount(float64(c.resourceBase[resourceName]) +
			float64(epoch-c.resourceEpoch)*
				float64(cfg.ResourceTrickles[tResourceName(resourceName)])*multiplier)
		c.resourceBase[resourceName] = currentResources - cost[resourceName]
	}
	return nil
}

func getGroupMovementSpeed(unitCount map[tUnitName]tUnitCount) tSpeed {
	for _, unitName := range sortedSlowestUnits {
		if unitCount[unitName] > 0 {
			return cfg.Units[unitName].UnitSpeed
		}
	}
	// NOTE: this should not happen, as movements would only happen with pre-existing units
	return 1.0
}

func travelTime(x1, y1, x2, y2 tCoordinate, speed tSpeed) tSec {
	// TODO: measure and optimize
	squareDist := (x1-x2)*(x1-x2) + (y1-y2)*(y1-y2)
	if squareDist <= 0 {
		return 0.0
	}

	return tSec(math.Sqrt(float64(squareDist)) / float64(speed))
}
