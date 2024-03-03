package internal

import (
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
	w.Write([]byte(`Welcome to Stickerio.\n`))
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

func (s *ServerHandler) GetBuildingQueueItem(w http.ResponseWriter, r *http.Request) {
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

	item, err := s.repository.GetBuildingtQueueItem(r.Context(), unitQueueItemID, city.id)
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
