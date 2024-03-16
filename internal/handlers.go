package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	api "github.com/luisferreira32/stickerio/models"
)

type eventSourcer interface {
	queueEventHandling(e *event)
}

func errHandle(w http.ResponseWriter, fmtStr string, args ...any) {
	http.Error(w, fmt.Sprintf(fmtStr, args...), http.StatusInternalServerError)
	// TODO better error handling instead of panic :)
	panic(fmt.Sprintf(fmtStr, args...))
}

func NewServerHandler(repository *StickerioRepository, eventSourcer eventSourcer) *ServerHandler {
	return &ServerHandler{
		eventSourcer: eventSourcer,
		viewer:       viewerService{repository: repository},
		inserter:     inserterService{repository: repository},
	}
}

type ServerHandler struct {
	eventSourcer eventSourcer
	viewer       viewerService
	inserter     inserterService
}

func (s *ServerHandler) GetWelcome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`Welcome to Stickerio API.`))
}

func (s *ServerHandler) GetCity(w http.ResponseWriter, r *http.Request) {
	cityID := r.Context().Value(CityIDKey).(string)
	playerID := r.Context().Value(PlayerIDKey).(string)
	city, err := s.viewer.GetCity(r.Context(), cityID, playerID)
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
			Epoch:     city.resourceEpoch,
			BaseCount: city.resourceBase,
		},
		UnitCount: city.unitCount,
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
	city, err := s.viewer.GetCityInfo(r.Context(), cityID)
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
	lastID := r.Context().Value(LastIDKey).(string)
	pageSize, err := strconv.Atoi(r.Context().Value(PageSize).(string))
	if err != nil {
		errHandle(w, err.Error())
	}

	cities, err := s.viewer.ListCityInfo(r.Context(), lastID, pageSize, playerIDFilter, locationBoundsFilter)
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
	movement, err := s.viewer.GetMovement(r.Context(), movementID, playerID)
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
		UnitCount:      movement.unitCount,
		ResourceCount:  movement.resourceCount,
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
	originIDFilter := r.URL.Query().Get(OriginID.String())

	movements, err := s.viewer.ListMovements(r.Context(), playerID, lastID, pageSize, originIDFilter)
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
		movementsList[i].UnitCount = movement.unitCount
		movementsList[i].ResourceCount = movement.unitCount
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

	item, err := s.viewer.GetUnitQueueItem(r.Context(), unitQueueItemID, cityID, playerID)
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

	items, err := s.viewer.ListUnitQueueItems(r.Context(), cityID, playerID, lastID, pageSize)
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

	item, err := s.viewer.GetBuildingQueueItem(r.Context(), buildingQueueItemID, cityID, playerID)
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

	items, err := s.viewer.ListBuildingQueueItems(r.Context(), cityID, playerID, lastID, pageSize)
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
	m := api.V1Movement{}
	err := decoder.Decode(&m)
	if err != nil {
		errHandle(w, err.Error())
	}

	err = s.inserter.StartMovement(r.Context(), &movement{
		id:            m.Id,
		playerID:      playerID,
		originID:      m.OriginID,
		destinationID: m.DestinationID,
		resourceCount: m.ResourceCount,
		unitCount:     m.UnitCount,
	})

	w.WriteHeader(http.StatusCreated)
}
