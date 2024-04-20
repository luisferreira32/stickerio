package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	api "github.com/luisferreira32/stickerio/models"
)

func errHandle(w http.ResponseWriter, fmtStr string, args ...any) {
	http.Error(w, fmt.Sprintf(fmtStr, args...), http.StatusInternalServerError)
	// TODO better error handling instead of panic :)
	panic(fmt.Sprintf(fmtStr, args...))
}

func NewServerHandler(repository *StickerioRepository, eventSourcer eventSourcer) *ServerHandler {
	return &ServerHandler{
		viewer:   viewerService{repository: repository},
		inserter: inserterService{repository: repository, eventSourcer: eventSourcer},
	}
}

type ServerHandler struct {
	viewer   viewerService
	inserter inserterService
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

	resp := cityToAPIModel(city)
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

	resp := cityToCityInfoAPIModel(city)
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
		resp[i] = cityToCityInfoAPIModel(cities[i])
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

	resp := movementToAPIModel(movement)
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
	destinationIDFilter := r.URL.Query().Get(DestinationID.String())

	movements, err := s.viewer.ListMovements(r.Context(), playerID, lastID, pageSize, originIDFilter, destinationIDFilter)
	if err != nil {
		errHandle(w, err.Error())
	}

	movementsList := make([]api.V1Movement, len(movements))
	for i := 0; i < len(movements); i++ {
		movementsList[i] = movementToAPIModel(movements[i])
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

	resp := unitQueueItemToAPIModel(item)
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
		unitQueueItemsList[i] = unitQueueItemToAPIModel(items[i])
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

	resp := buildingQueueItemToAPIModel(item)
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
		buildingQueueItemsList[i] = buildingQueueItemToAPIModel(items[i])
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

	err = s.inserter.StartMovement(r.Context(), playerID, &movement{
		id:            tMovementID(m.Id),
		originID:      tCityID(m.OriginID),
		destinationID: tCityID(m.DestinationID),
		destinationX:  tCoordinate(m.DestinationX),
		destinationY:  tCoordinate(m.DestinationY),
		resourceCount: m.ResourceCount,
		unitCount:     m.UnitCount,
	})
	if err != nil {
		errHandle(w, err.Error())
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *ServerHandler) QueueUnit(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value(PlayerIDKey).(string)
	cityID := r.Context().Value(CityIDKey).(string)

	decoder := json.NewDecoder(r.Body)
	item := api.V1UnitQueueItem{}
	err := decoder.Decode(&item)
	if err != nil {
		errHandle(w, err.Error())
	}

	err = s.inserter.QueueUnit(r.Context(), playerID, &unitQueueItem{
		id:        tUnitQueueItemID(item.Id),
		cityID:    tCityID(cityID),
		unitCount: tUnitCount(item.UnitCount),
		unitType:  tUnitName(item.UnitType),
	})
	if err != nil {
		errHandle(w, err.Error())
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *ServerHandler) QueueBuilding(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value(PlayerIDKey).(string)
	cityID := r.Context().Value(CityIDKey).(string)

	decoder := json.NewDecoder(r.Body)
	item := api.V1BuildingQueueItem{}
	err := decoder.Decode(&item)
	if err != nil {
		errHandle(w, err.Error())
	}

	err = s.inserter.QueueBuilding(r.Context(), playerID, &buildingQueueItem{
		id:             tBuildingQueueItemID(item.Id),
		cityID:         tCityID(cityID),
		targetLevel:    tBuildingLevel(item.Level),
		targetBuilding: tBuildingName(item.Building),
	})
	if err != nil {
		errHandle(w, err.Error())
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *ServerHandler) CreateCity(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value(PlayerIDKey).(string)

	decoder := json.NewDecoder(r.Body)
	m := api.V1CityInfo{}
	err := decoder.Decode(&m)
	if err != nil {
		errHandle(w, err.Error())
	}

	err = s.inserter.CreateCity(r.Context(), playerID, &city{
		id:        tCityID(m.Id),
		name:      m.Name,
		playerID:  tPlayerID(playerID),
		locationX: tCoordinate(m.LocationX),
		locationY: tCoordinate(m.LocationY),
	})
	if err != nil {
		errHandle(w, err.Error())
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *ServerHandler) DeleteCity(w http.ResponseWriter, r *http.Request) {
	playerID := r.Context().Value(PlayerIDKey).(string)
	cityID := r.Context().Value(CityIDKey).(string)

	err := s.inserter.DeleteCity(r.Context(), playerID, cityID)
	if err != nil {
		errHandle(w, err.Error())
	}

	w.WriteHeader(http.StatusAccepted)
}
