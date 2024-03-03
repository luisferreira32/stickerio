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
		Units: &stickerio.Units{
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
