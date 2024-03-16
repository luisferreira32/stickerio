package internal

import (
	"context"
	"log"
	"sync"
	"time"
)

const (
	startMovementEventName   = "startmovement"
	arrivalMovementEventName = "arrival"
)

// Inserted when a player starts a movement.
// Its processing generates an arrivalMovementEvent at a later epcoh (calculated based on distance between cities / speed).
type startMovementEvent struct {
	MovementID     string  `json:"movementID"`
	PlayerID       string  `json:"playerID"`
	OriginID       string  `json:"originID"`
	DestinationID  string  `json:"destinationID"`
	DepartureEpoch int64   `json:"departureEpoch"`
	Speed          float64 `json:"speed"`
	StickmenCount  int     `json:"stickmenCount"`
	SwordsmenCount int     `json:"swordsmenCount"`
	SticksCount    int     `json:"sticksCount"`
	CirclesCount   int     `json:"circlesCount"`
}

// Inserted when a startMovementEvent is processed.
// The arrival processing does the options:
// * reinforce if it's from the same player;
// * battle if it's from separate player
// * forage if it's abandoned
// Then, based on survival (if any), if the current city is not owned by the player (i.e., it was not a reinforce), re-calculate a startMovement event and schedule it.
type arrivalMovementEvent struct {
	MovementID     string `json:"movementID"`
	PlayerID       string `json:"playerID"`
	OriginID       string `json:"originID"`
	DestinationID  string `json:"destinationID"`
	SwordsmenCount int    `json:"swordsmenCount"`
	StickmenCount  int    `json:"stickmenCount"`
	SticksCount    int    `json:"sticksCount"`
	CirclesCount   int    `json:"circlesCount"`
	CityID         string `json:"cityID"`
}

type eventsRepository interface {
	InsertEvent(ctx context.Context, e *event) error
	ListEvents(ctx context.Context, untilEpoch int64) ([]*event, error)
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
type EventSourcer struct {
	repository eventsRepository

	inMemoryStateLock     *sync.Mutex
	cityList              map[string]*city
	movementList          map[string]*movement
	unitQueuesPerCity     map[string]map[string]*unitQueueItem
	buildingQueuesPerCity map[string]map[string]*buildingQueueItem

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
		case arrivalMovementEventName:
		}
	}

	return nil
}
