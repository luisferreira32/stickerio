package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	api "github.com/luisferreira32/stickerio/models"
)

type stickerioRepository interface {
	InsertEvent(ctx context.Context, e *event) error
	GetCity(ctx context.Context, id, playerID string) (*city, error)
	GetCityInfo(ctx context.Context, id string) (*city, error)
	ListCityInfo(ctx context.Context, lastID string, pageSize int, filters ...listCityInfoFilterOpt) ([]*city, error)
	GetMovement(ctx context.Context, id, playerID string) (*movement, error)
	ListMovements(ctx context.Context, playerID, lastID string, pageSize int, filters ...listMovementsFilterOpt) ([]*movement, error)
	GetUnitQueueItem(ctx context.Context, id, cityID string) (*unitQueueItem, error)
	ListUnitQueueItems(ctx context.Context, cityID, lastID string, pageSize int) ([]*unitQueueItem, error)
	GetBuildingQueueItem(ctx context.Context, id, cityID string) (*buildingQueueItem, error)
	ListBuildingQueueItems(ctx context.Context, cityID, lastID string, pageSize int) ([]*buildingQueueItem, error)
}

type eventSourcer interface {
	queueEventHandling(e *event)
}

func errHandle(w http.ResponseWriter, fmtStr string, args ...any) {
	http.Error(w, fmt.Sprintf(fmtStr, args...), http.StatusInternalServerError)
	// TODO better error handling instead of panic :)
	panic(fmt.Sprintf(fmtStr, args...))
}

func NewServerHandler(repository stickerioRepository, eventSourcer eventSourcer) *ServerHandler {
	return &ServerHandler{
		repository:   repository,
		eventSourcer: eventSourcer,
	}
}

type ServerHandler struct {
	repository   stickerioRepository
	eventSourcer eventSourcer
}

func (s *ServerHandler) GetWelcome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`Welcome to Stickerio API.`))
}

func (s *ServerHandler) GetCity(w http.ResponseWriter, r *http.Request) {
	cityID := r.Context().Value(CityIDKey).(string)
	playerID := r.Context().Value(PlayerIDKey).(string)
	city, err := s.repository.GetCity(r.Context(), cityID, playerID)
	if err != nil {
		errHandle(w, err.Error())
	}

	resp := api.V1City{
		CityInfo: api.V1CityInfo{
			Id:        city.id,
			Name:      city.name,
			PlayerID:  city.playerID,
			LocationX: city.locationX,
			LocationY: city.locationY,
		},
		CityBuildings: api.V1CityBuildings{
			BarracksLevel: city.barracksLevel,
			MinesLevel:    city.mineLevel,
		},
		CityResources: api.V1CityResources{
			SticksCountBase:   city.sticksCountBase,
			SticksCountEpoch:  city.sticksCountEpoch,
			CirclesCountBase:  city.circlesCountBase,
			CirclesCountEpoch: city.circlesCountEpoch,
		},
		UnitCount: api.V1UnitCount{
			StickmenCount:  city.stickmenCount,
			SwordsmenCount: city.swordsmenCount,
		},
	}

	respBytes, err := resp.MarshalJSON()
	if err != nil {
		errHandle(w, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(respBytes)
	if err != nil {
		errHandle(w, err.Error())
	}
}

func (s *ServerHandler) GetCityInfo(w http.ResponseWriter, r *http.Request) {
	cityID := r.Context().Value(CityIDKey).(string)
	city, err := s.repository.GetCityInfo(r.Context(), cityID)
	if err != nil {
		errHandle(w, err.Error())
	}

	resp := api.V1CityInfo{
		Id:        city.id,
		Name:      city.name,
		PlayerID:  city.playerID,
		LocationX: city.locationX,
		LocationY: city.locationY,
	}

	respBytes, err := resp.MarshalJSON()
	if err != nil {
		errHandle(w, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(respBytes)
	if err != nil {
		errHandle(w, err.Error())
	}
}

func (s *ServerHandler) ListCityInfo(w http.ResponseWriter, r *http.Request) {
	playerIDFilter := r.URL.Query().Get(PlayerID.String())
	locationBoundsFilter := r.URL.Query().Get(LocationBounds.String())
	additionalFilters := make([]listCityInfoFilterOpt, 0)
	if playerIDFilter != "" {
		additionalFilters = append(additionalFilters, withPlayerID(playerIDFilter))
	}
	if locationBoundsFilter != "" {
		// TODO: fix this
	}
	lastID := r.Context().Value(LastIDKey).(string)
	pageSize, err := strconv.Atoi(r.Context().Value(PageSize).(string))
	if err != nil {
		errHandle(w, err.Error())
	}

	cities, err := s.repository.ListCityInfo(r.Context(), lastID, pageSize, additionalFilters...)
	if err != nil {
		errHandle(w, err.Error())
	}

	resp := make([]api.V1CityInfo, len(cities))
	for i := 0; i < len(cities); i++ {
		resp[i].Id = cities[i].id
		resp[i].Name = cities[i].name
		resp[i].PlayerID = cities[i].playerID
		resp[i].LocationX = cities[i].locationX
		resp[i].LocationY = cities[i].locationY
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		errHandle(w, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(respBytes)
	if err != nil {
		errHandle(w, err.Error())
	}
}

func (s *ServerHandler) GetMovement(w http.ResponseWriter, r *http.Request) {
	movementID := r.Context().Value(MovementIDKey).(string)
	playerID := r.Context().Value(PlayerIDKey).(string)
	movement, err := s.repository.GetMovement(r.Context(), movementID, playerID)
	if err != nil {
		errHandle(w, err.Error())
	}

	resp := &api.V1Movement{
		Id:             movement.id,
		PlayerID:       movement.playerID,
		OriginID:       movement.originID,
		DestinationID:  movement.destinationID,
		DepartureEpoch: movement.departureEpoch,
		Speed:          movement.speed,
		UnitCount: api.V1UnitCount{
			StickmenCount:  movement.stickmenCount,
			SwordsmenCount: movement.swordmenCount,
		},
		ResourceCount: api.V1ResourceCount{
			SticksCount:  movement.stickCount,
			CirclesCount: movement.circlesCount,
		},
	}

	respBytes, err := resp.MarshalJSON()
	if err != nil {
		errHandle(w, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(respBytes)
	if err != nil {
		errHandle(w, err.Error())
	}
}

func (s *ServerHandler) ListMovements(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value(PlayerIDKey).(string)
	lastID := r.Context().Value(LastIDKey).(string)
	pageSize, err := strconv.Atoi(r.Context().Value(PageSize).(string))
	if err != nil {
		errHandle(w, err.Error())
	}

	additionalFilters := make([]listMovementsFilterOpt, 0)
	originIDFilter := r.URL.Query().Get(OriginID.String())
	if originIDFilter != "" {
		additionalFilters = append(additionalFilters, withOriginCityID(originIDFilter))
	}

	movements, err := s.repository.ListMovements(r.Context(), playerID, lastID, pageSize, additionalFilters...)
	if err != nil {
		errHandle(w, err.Error())
	}

	movementsList := make([]api.V1Movement, len(movements))
	for i := 0; i < len(movements); i++ {
		movement := movements[i]
		movementsList[i].Id = movement.id
		movementsList[i].PlayerID = movement.playerID
		movementsList[i].OriginID = movement.originID
		movementsList[i].DestinationID = movement.destinationID
		movementsList[i].DepartureEpoch = movement.departureEpoch
		movementsList[i].Speed = movement.speed
		movementsList[i].UnitCount.StickmenCount = movement.stickmenCount
		movementsList[i].UnitCount.SwordsmenCount = movement.swordmenCount
		movementsList[i].ResourceCount.SticksCount = movement.stickCount
		movementsList[i].ResourceCount.CirclesCount = movement.circlesCount
	}

	resp, err := json.Marshal(movementsList)
	if err != nil {
		errHandle(w, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		errHandle(w, err.Error())
	}
}

func (s *ServerHandler) GetUnitQueueItem(w http.ResponseWriter, r *http.Request) {
	unitQueueItemID := r.Context().Value(UnitQueueItemIDKey).(string)
	cityID := r.Context().Value(CityIDKey).(string)
	playerID := r.Context().Value(PlayerIDKey).(string)

	city, err := s.repository.GetCity(r.Context(), cityID, playerID)
	if err != nil {
		errHandle(w, err.Error())
	}

	item, err := s.repository.GetUnitQueueItem(r.Context(), unitQueueItemID, city.id)
	if err != nil {
		errHandle(w, err.Error())
	}

	resp := &api.V1UnitQueueItem{
		Id:          item.id,
		QueuedEpoch: item.queuedEpoch,
		DurationSec: item.durationSec,
		UnitCount:   item.unitCount,
		UnitType:    item.unitType,
	}

	respBytes, err := resp.MarshalJSON()
	if err != nil {
		errHandle(w, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(respBytes)
	if err != nil {
		errHandle(w, err.Error())
	}
}

func (s *ServerHandler) ListUnitQueueItem(w http.ResponseWriter, r *http.Request) {
	cityID := r.Context().Value(CityIDKey).(string)
	playerID := r.Context().Value(PlayerIDKey).(string)
	lastID := r.Context().Value(LastIDKey).(string)
	pageSize, err := strconv.Atoi(r.Context().Value(PageSize).(string))
	if err != nil {
		errHandle(w, err.Error())
	}

	city, err := s.repository.GetCity(r.Context(), cityID, playerID)
	if err != nil {
		errHandle(w, err.Error())
	}

	items, err := s.repository.ListUnitQueueItems(r.Context(), city.id, lastID, pageSize)
	if err != nil {
		errHandle(w, err.Error())
	}

	unitQueueItemsList := make([]api.V1UnitQueueItem, len(items))
	for i := 0; i < len(items); i++ {
		item := items[i]
		unitQueueItemsList[i].Id = item.id
		unitQueueItemsList[i].QueuedEpoch = item.queuedEpoch
		unitQueueItemsList[i].DurationSec = item.durationSec
		unitQueueItemsList[i].UnitCount = item.unitCount
		unitQueueItemsList[i].UnitType = item.unitType
	}

	respBytes, err := json.Marshal(unitQueueItemsList)
	if err != nil {
		errHandle(w, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(respBytes)
	if err != nil {
		errHandle(w, err.Error())
	}
}

func (s *ServerHandler) GetBuildingQueueItem(w http.ResponseWriter, r *http.Request) {
	buildingQueueItemID := r.Context().Value(BuildingQueueItemIDKey).(string)
	cityID := r.Context().Value(CityIDKey).(string)
	playerID := r.Context().Value(PlayerIDKey).(string)

	city, err := s.repository.GetCity(r.Context(), cityID, playerID)
	if err != nil {
		errHandle(w, err.Error())
	}

	item, err := s.repository.GetBuildingQueueItem(r.Context(), buildingQueueItemID, city.id)
	if err != nil {
		errHandle(w, err.Error())
	}

	resp := &api.V1BuildingQueueItem{
		Id:          item.id,
		QueuedEpoch: item.queuedEpoch,
		DurationSec: item.durationSec,
		Level:       item.targetLevel,
		Building:    item.targetBuilding,
	}

	respBytes, err := resp.MarshalJSON()
	if err != nil {
		errHandle(w, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(respBytes)
	if err != nil {
		errHandle(w, err.Error())
	}
}

func (s *ServerHandler) ListBuildingQueueItems(w http.ResponseWriter, r *http.Request) {
	cityID := r.Context().Value(CityIDKey).(string)
	playerID := r.Context().Value(PlayerIDKey).(string)
	lastID := r.Context().Value(LastIDKey).(string)
	pageSize, err := strconv.Atoi(r.Context().Value(PageSize).(string))
	if err != nil {
		errHandle(w, err.Error())
	}

	city, err := s.repository.GetCity(r.Context(), cityID, playerID)
	if err != nil {
		errHandle(w, err.Error())
	}

	items, err := s.repository.ListBuildingQueueItems(r.Context(), city.id, lastID, pageSize)
	if err != nil {
		errHandle(w, err.Error())
	}

	buildingQueueItemsList := make([]api.V1BuildingQueueItem, len(items))
	for i := 0; i < len(items); i++ {
		item := items[i]
		buildingQueueItemsList[i].Id = item.id
		buildingQueueItemsList[i].QueuedEpoch = item.queuedEpoch
		buildingQueueItemsList[i].DurationSec = item.durationSec
		buildingQueueItemsList[i].Level = item.targetLevel
		buildingQueueItemsList[i].Building = item.targetBuilding
	}

	resp, err := json.Marshal(buildingQueueItemsList)
	if err != nil {
		errHandle(w, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		errHandle(w, err.Error())
	}
}

func (s *ServerHandler) StartMovement(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value(PlayerIDKey).(string)

	decoder := json.NewDecoder(r.Body)
	startMovement := api.V1Movement{}
	err := decoder.Decode(&startMovement)
	if err != nil {
		errHandle(w, err.Error())
	}

	serversideEpoch := time.Now().Unix()

	// important: these values cannot be trusted from the API
	// set them on the server side based on token / internal clock
	startMovement.DepartureEpoch = serversideEpoch
	startMovement.PlayerID = playerID

	payload, err := startMovement.MarshalJSON()
	if err != nil {
		errHandle(w, err.Error())
	}

	eventID := uuid.NewString()
	s.repository.InsertEvent(r.Context(), &event{
		id:      eventID,
		epoch:   serversideEpoch,
		payload: string(payload), // json marshalled bytes are valid UTF-8 strings
	})

	w.WriteHeader(http.StatusCreated)
}
