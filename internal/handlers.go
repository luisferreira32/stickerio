package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/luisferreira32/stickerio"
)

func errHandle(fmtStr string, args ...any) {
	// TODO better error handling instead of panic :)
	panic(fmt.Sprintf(fmtStr, args...))
}

func NewServerHandler(repository StickerioRepository) *ServerHandler {
	return &ServerHandler{
		repository: repository,
	}
}

type ServerHandler struct {
	repository StickerioRepository
}

func (s *ServerHandler) GetWelcome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`Welcome to Stickerio API.`))
}

func (s *ServerHandler) GetCity(w http.ResponseWriter, r *http.Request) {
	cityID, ok1 := r.Context().Value(CityIDKey).(string)
	playerID, ok2 := r.Context().Value(PlayerIDKey).(string)
	if !ok1 || !ok2 {
		errHandle("cityID/playerID not a string: %v, %v", r.Context().Value(CityIDKey), r.Context().Value(PlayerIDKey))
	}
	city, err := s.repository.GetCity(r.Context(), cityID, playerID)
	if err != nil {
		errHandle(err.Error())
	}

	resp, err := json.Marshal(&stickerio.City{
		CityInfo: &stickerio.CityInfo{
			ID:        city.id,
			Name:      city.name,
			PlayerID:  city.playerID,
			LocationX: city.locationX,
			LocationY: city.locationY,
		},
		CityBuildings: &stickerio.CityBuildings{
			BarracksLevel: city.barracksLevel,
			MineLevel:     city.mineLevel,
		},
		CityResources: &stickerio.CityResources{
			SticksCountBase:  city.sticksCountBase,
			SticksCountEpoch: city.sticksCountEpoch,
		},
		UnitCount: &stickerio.UnitCount{
			SwordsmenCount: city.swordsmenCount,
		},
	})
	if err != nil {
		errHandle(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		errHandle(err.Error())
	}
}

func (s *ServerHandler) GetCityInfo(w http.ResponseWriter, r *http.Request) {
	cityID, ok := r.Context().Value(CityIDKey).(string)
	if !ok {
		errHandle("cityID not a string: %v", r.Context().Value(CityIDKey))
	}
	city, err := s.repository.GetCityInfo(r.Context(), cityID)
	if err != nil {
		errHandle(err.Error())
	}

	resp, err := json.Marshal(&stickerio.City{
		CityInfo: &stickerio.CityInfo{
			ID:        city.id,
			Name:      city.name,
			PlayerID:  city.playerID,
			LocationX: city.locationX,
			LocationY: city.locationY,
		},
	})
	if err != nil {
		errHandle(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		errHandle(err.Error())
	}
}

func (s *ServerHandler) ListCityInfo(w http.ResponseWriter, r *http.Request) {
	b64EncodedFilters := r.URL.Query().Get("filters")
	filtersReq := &stickerio.ListCityInfoFilters{
		PageSize: 10, // defaults
	}
	additionalFilters := make([]listCityInfoFilterOpt, 0)
	if b64EncodedFilters != "" {
		jsonEncodedFilters, err := base64.RawURLEncoding.DecodeString(b64EncodedFilters)
		if err != nil {
			errHandle(err.Error())
		}
		err = json.Unmarshal(jsonEncodedFilters, filtersReq)
		if err != nil {
			errHandle(err.Error())
		}
	}
	if filtersReq.PlayerID != "" {
		additionalFilters = append(additionalFilters, withPlayerID(filtersReq.PlayerID))
	}
	if len(filtersReq.LocationBounds) == 4 && filtersReq.LocationBounds[0] <= filtersReq.LocationBounds[2] && filtersReq.LocationBounds[1] <= filtersReq.LocationBounds[3] {
		additionalFilters = append(additionalFilters, withinLocation(
			filtersReq.LocationBounds[0],
			filtersReq.LocationBounds[1],
			filtersReq.LocationBounds[2],
			filtersReq.LocationBounds[3],
		)...)
	}
	cities, err := s.repository.ListCityInfo(r.Context(), filtersReq.LastCityID, filtersReq.PageSize, additionalFilters...)
	if err != nil {
		errHandle(err.Error())
	}

	cityInfoList := make([]*stickerio.CityInfo, len(cities))
	for i := 0; i < len(cities); i++ {
		cityInfoList[i] = &stickerio.CityInfo{
			ID:        cities[i].id,
			Name:      cities[i].name,
			PlayerID:  cities[i].playerID,
			LocationX: cities[i].locationX,
			LocationY: cities[i].locationY,
		}
	}

	resp, err := json.Marshal(cityInfoList)
	if err != nil {
		errHandle(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		errHandle(err.Error())
	}
}

func (s *ServerHandler) GetMovement(w http.ResponseWriter, r *http.Request) {
	movementID, ok1 := r.Context().Value(MovementIDKey).(string)
	playerID, ok2 := r.Context().Value(PlayerIDKey).(string)
	if !ok1 || !ok2 {
		errHandle("movementID/playerID not a string: %v, %v", r.Context().Value(MovementIDKey), r.Context().Value(PlayerIDKey))
	}
	movement, err := s.repository.GetMovement(r.Context(), movementID, playerID)
	if err != nil {
		errHandle(err.Error())
	}

	resp, err := json.Marshal(&stickerio.Movement{
		ID:             movement.id,
		PlayerID:       movement.playerID,
		OriginID:       movement.originID,
		DestinationID:  movement.destinationID,
		DepartureEpoch: movement.departureEpoch,
		Speed:          movement.speed,
		UnitCount: &stickerio.UnitCount{
			StickmenCount:  movement.stickmenCount,
			SwordsmenCount: movement.swordmenCount,
		},
		ResourcesCount: &stickerio.ResourcesCount{
			SticksCount:  movement.stickCount,
			CirclesCount: movement.circlesCount,
		},
	})
	if err != nil {
		errHandle(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		errHandle(err.Error())
	}
}

func (s *ServerHandler) ListMovements(w http.ResponseWriter, r *http.Request) {
	playerID, ok := r.Context().Value(PlayerIDKey).(string)
	if !ok {
		errHandle("playerID not a string: %v", r.Context().Value(PlayerIDKey))
	}
	b64EncodedFilters := r.URL.Query().Get("filters")
	filtersReq := &stickerio.ListMovementsFilters{
		PageSize: 10, // defaults
	}
	additionalFilters := make([]listMovementsFilterOpt, 0)
	if b64EncodedFilters != "" {
		jsonEncodedFilters, err := base64.RawURLEncoding.DecodeString(b64EncodedFilters)
		if err != nil {
			errHandle(err.Error())
		}
		err = json.Unmarshal(jsonEncodedFilters, filtersReq)
		if err != nil {
			errHandle(err.Error())
		}
	}
	if filtersReq.OriginCityID != "" {
		additionalFilters = append(additionalFilters, withOriginCityID(filtersReq.OriginCityID))
	}
	movements, err := s.repository.ListMovements(r.Context(), playerID, filtersReq.LastMovementID, filtersReq.PageSize, additionalFilters...)
	if err != nil {
		errHandle(err.Error())
	}

	movementsList := make([]*stickerio.Movement, len(movements))
	for i := 0; i < len(movements); i++ {
		movement := movements[i]
		movementsList[i] = &stickerio.Movement{
			ID:             movement.id,
			PlayerID:       movement.playerID,
			OriginID:       movement.originID,
			DestinationID:  movement.destinationID,
			DepartureEpoch: movement.departureEpoch,
			Speed:          movement.speed,
			UnitCount: &stickerio.UnitCount{
				StickmenCount:  movement.stickmenCount,
				SwordsmenCount: movement.swordmenCount,
			},
			ResourcesCount: &stickerio.ResourcesCount{
				SticksCount:  movement.stickCount,
				CirclesCount: movement.circlesCount,
			},
		}
	}

	resp, err := json.Marshal(movementsList)
	if err != nil {
		errHandle(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		errHandle(err.Error())
	}
}

func (s *ServerHandler) GetUnitQueueItem(w http.ResponseWriter, r *http.Request) {
	unitQueueItemID, ok1 := r.Context().Value(UnitQueueItemIDKey).(string)
	cityID, ok2 := r.Context().Value(CityIDKey).(string)
	playerID, ok3 := r.Context().Value(PlayerIDKey).(string)
	if !ok1 || !ok2 || !ok3 {
		errHandle("unitQueueItemID/cityID/playerID not a string: %v, %v, %v", r.Context().Value(UnitQueueItemIDKey), r.Context().Value(CityIDKey), r.Context().Value(PlayerIDKey))
	}

	city, err := s.repository.GetCity(r.Context(), cityID, playerID)
	if err != nil {
		errHandle(err.Error())
	}

	item, err := s.repository.GetUnitQueueItem(r.Context(), unitQueueItemID, city.id)
	if err != nil {
		errHandle(err.Error())
	}

	resp, err := json.Marshal(&stickerio.UnitQueueItem{
		ID:          item.id,
		CityID:      item.cityID,
		QueuedEpoch: item.queuedEpoch,
		DurationSec: item.durationSec,
		UnitCount:   item.unitCount,
		UnitType:    stickerio.UnitName(item.unitType),
	})
	if err != nil {
		errHandle(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		errHandle(err.Error())
	}
}

func (s *ServerHandler) ListUnitQueueItem(w http.ResponseWriter, r *http.Request) {
	cityID, ok1 := r.Context().Value(CityIDKey).(string)
	playerID, ok2 := r.Context().Value(PlayerIDKey).(string)
	if !ok1 || !ok2 {
		errHandle("cityID/playerID not a string: %v, %v", r.Context().Value(CityIDKey), r.Context().Value(PlayerIDKey))
	}

	b64EncodedFilters := r.URL.Query().Get("filters")
	filtersReq := &stickerio.ListUnitQueueItemFilters{
		PageSize: 10, // defaults
	}
	if b64EncodedFilters != "" {
		jsonEncodedFilters, err := base64.RawURLEncoding.DecodeString(b64EncodedFilters)
		if err != nil {
			errHandle(err.Error())
		}
		err = json.Unmarshal(jsonEncodedFilters, filtersReq)
		if err != nil {
			errHandle(err.Error())
		}
	}

	city, err := s.repository.GetCity(r.Context(), cityID, playerID)
	if err != nil {
		errHandle(err.Error())
	}

	items, err := s.repository.ListUnitQueueItems(r.Context(), city.id, filtersReq.LastMovementID, filtersReq.PageSize)
	if err != nil {
		errHandle(err.Error())
	}

	unitQueueItemsList := make([]*stickerio.UnitQueueItem, len(items))
	for i := 0; i < len(items); i++ {
		item := items[i]
		unitQueueItemsList[i] = &stickerio.UnitQueueItem{
			ID:          item.id,
			CityID:      item.cityID,
			QueuedEpoch: item.queuedEpoch,
			DurationSec: item.durationSec,
			UnitCount:   item.unitCount,
			UnitType:    stickerio.UnitName(item.unitType),
		}
	}

	resp, err := json.Marshal(unitQueueItemsList)
	if err != nil {
		errHandle(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		errHandle(err.Error())
	}
}

func (s *ServerHandler) GetBuildingQueueItem(w http.ResponseWriter, r *http.Request) {
	buildingQueueItemID, ok1 := r.Context().Value(BuildingQueueItemIDKey).(string)
	cityID, ok2 := r.Context().Value(CityIDKey).(string)
	playerID, ok3 := r.Context().Value(PlayerIDKey).(string)
	if !ok1 || !ok2 || !ok3 {
		errHandle("unitQueueItemID/cityID/playerID not a string: %v, %v, %v", r.Context().Value(BuildingQueueItemIDKey), r.Context().Value(CityIDKey), r.Context().Value(PlayerIDKey))
	}

	city, err := s.repository.GetCity(r.Context(), cityID, playerID)
	if err != nil {
		errHandle(err.Error())
	}

	item, err := s.repository.GetBuildingQueueItem(r.Context(), buildingQueueItemID, city.id)
	if err != nil {
		errHandle(err.Error())
	}

	resp, err := json.Marshal(&stickerio.BuildingQueueItem{
		ID:             item.id,
		CityID:         item.cityID,
		QueuedEpoch:    item.queuedEpoch,
		DurationSec:    item.durationSec,
		TargetLevel:    item.targetLevel,
		TargetBuilding: stickerio.BuildingName(item.targetBuilding),
	})
	if err != nil {
		errHandle(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		errHandle(err.Error())
	}
}

func (s *ServerHandler) ListBuildingQueueItems(w http.ResponseWriter, r *http.Request) {
	cityID, ok1 := r.Context().Value(CityIDKey).(string)
	playerID, ok2 := r.Context().Value(PlayerIDKey).(string)
	if !ok1 || !ok2 {
		errHandle("cityID/playerID not a string: %v, %v", r.Context().Value(CityIDKey), r.Context().Value(PlayerIDKey))
	}

	b64EncodedFilters := r.URL.Query().Get("filters")
	filtersReq := &stickerio.LisBuildingQueueItemFilters{
		PageSize: 10, // defaults
	}
	if b64EncodedFilters != "" {
		jsonEncodedFilters, err := base64.RawURLEncoding.DecodeString(b64EncodedFilters)
		if err != nil {
			errHandle(err.Error())
		}
		err = json.Unmarshal(jsonEncodedFilters, filtersReq)
		if err != nil {
			errHandle(err.Error())
		}
	}

	city, err := s.repository.GetCity(r.Context(), cityID, playerID)
	if err != nil {
		errHandle(err.Error())
	}

	items, err := s.repository.ListBuildingQueueItems(r.Context(), city.id, filtersReq.LastMovementID, filtersReq.PageSize)
	if err != nil {
		errHandle(err.Error())
	}

	buildingQueueItemsList := make([]*stickerio.BuildingQueueItem, len(items))
	for i := 0; i < len(items); i++ {
		item := items[i]
		buildingQueueItemsList[i] = &stickerio.BuildingQueueItem{
			ID:             item.id,
			CityID:         item.cityID,
			QueuedEpoch:    item.queuedEpoch,
			DurationSec:    item.durationSec,
			TargetLevel:    item.targetLevel,
			TargetBuilding: stickerio.BuildingName(item.targetBuilding),
		}
	}

	resp, err := json.Marshal(buildingQueueItemsList)
	if err != nil {
		errHandle(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		errHandle(err.Error())
	}
}
