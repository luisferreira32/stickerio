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
)

var (
	// thrown when the event is correctly built but it is invalid to process it
	// e.g., try to move more troops than you own
	// if you're doing a batch processing of events just skip over this one
	errPreConditionFailed = fmt.Errorf("pre-condition failed")
)

// NOTE: this types exist to make sure event sourcing compares are against valid types

type tMovementID string
type tPlayerID string
type tCityID string
type tEpoch int64
type tUnitCount int32
type tResourceCount int64
type tItemID string

// Inserted when a player starts a movement.
// Its processing generates an arrivalMovementEvent at a later epcoh (calculated based on distance between cities / speed).
type startMovementEvent struct {
	MovementID     tMovementID    `json:"movementID"`
	PlayerID       tPlayerID      `json:"playerID"`
	OriginID       tCityID        `json:"originID"`
	DestinationID  tCityID        `json:"destinationID"`
	DepartureEpoch tEpoch         `json:"departureEpoch"`
	StickmenCount  tUnitCount     `json:"stickmenCount"`
	SwordsmenCount tUnitCount     `json:"swordsmenCount"`
	SticksCount    tResourceCount `json:"sticksCount"`
	CirclesCount   tResourceCount `json:"circlesCount"`
}

// Inserted when a startMovementEvent is processed.
// The arrival processing does the options:
// * reinforce if it's from the same player;
// * battle if it's from separate player
// * forage if it's abandoned
// Then, based on survival (if any), if the current city is not owned by the player (i.e., it was not a reinforce), re-calculate a startMovement event and schedule it.
type arrivalMovementEvent struct {
	MovementID     tMovementID    `json:"movementID"`
	PlayerID       tPlayerID      `json:"playerID"`
	OriginID       tCityID        `json:"originID"`
	DestinationID  tCityID        `json:"destinationID"`
	SwordsmenCount tUnitCount     `json:"swordsmenCount"`
	StickmenCount  tUnitCount     `json:"stickmenCount"`
	SticksCount    tResourceCount `json:"sticksCount"`
	CirclesCount   tResourceCount `json:"circlesCount"`
}

type eventsRepository interface {
	InsertEvent(ctx context.Context, e *event) error
	ListEvents(ctx context.Context, untilEpoch int64) ([]*event, error)
	UpsertMovement(ctx context.Context, m *movement) error
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

	internalEventQueue chan *event
}

func NewEventSourcer(repository eventsRepository) *EventSourcer {
	return &EventSourcer{
		repository:         repository,
		inMemoryStateLock:  &sync.Mutex{},
		internalEventQueue: make(chan *event, 100),
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
	}

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
		}
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

	// validation
	if startMovement.OriginID == startMovement.DestinationID {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "cannot move to the same city")
	}
	originCity := s.cityList[startMovement.OriginID]
	currentSticks := currentResources(originCity.sticksCountBase, originCity.sticksCountEpoch, sticks)
	currentCircles := currentResources(originCity.circlesCountBase, originCity.circlesCountEpoch, circles)
	if tUnitCount(originCity.stickmenCount) < startMovement.StickmenCount ||
		tUnitCount(originCity.swordsmenCount) < startMovement.SwordsmenCount ||
		currentSticks < startMovement.SticksCount ||
		currentCircles < startMovement.CirclesCount {
		return fmt.Errorf("%w, event %s, reason: %s", errPreConditionFailed, e.id, "insuficient resources/units")
	}

	// insert new event
	arrival := &arrivalMovementEvent{
		MovementID:     startMovement.MovementID,
		PlayerID:       startMovement.PlayerID,
		OriginID:       startMovement.OriginID,
		DestinationID:  startMovement.DestinationID,
		StickmenCount:  startMovement.StickmenCount,
		SwordsmenCount: startMovement.SwordsmenCount,
		SticksCount:    startMovement.SticksCount,
		CirclesCount:   startMovement.CirclesCount,
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

	// upsert view table
	m := &movement{
		id:             string(startMovement.MovementID),
		playerID:       string(startMovement.PlayerID),
		originID:       string(startMovement.OriginID),
		destinationID:  string(startMovement.DestinationID),
		departureEpoch: int64(startMovement.DepartureEpoch),
		circlesCount:   int32(startMovement.StickmenCount),
		stickCount:     int32(startMovement.SwordsmenCount),
		stickmenCount:  int32(startMovement.SticksCount),
		swordmenCount:  int32(startMovement.CirclesCount),
		speed:          getMovementSpeed(startMovement.StickmenCount, startMovement.SwordsmenCount),
	}
	err = s.repository.UpsertMovement(ctx, m)
	if err != nil {
		return err
	}

	return nil
}

func getMovementSpeed(stickmenCount, swordsmenCount tUnitCount) float32 {
	var movementSpeed float32
	for _, slowUnit := range slowestUnits {
		if slowUnit == stickmen && stickmenCount > 0 {
			movementSpeed = config.Units[stickmen].UnitSpeed
			break
		}
		if slowUnit == stickmen && swordsmenCount > 0 {
			movementSpeed = config.Units[stickmen].UnitSpeed
			break
		}
	}
	return movementSpeed
}

func currentResources(base, epoch int64, n resourceName) tResourceCount {
	return tResourceCount(base + epoch*int64(config.ResourceTrickles[n]))
}
