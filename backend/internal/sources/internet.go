package sources

import (
	"context"
	"encoding/json"
	"io"
	"mediaplayer/backend/internal/source"
	"net/http"
)

type internetAdapter struct {
	defStation source.Station
}

func NewInternetAdapter() source.Adapter {
	return &internetAdapter{}
}

func (a *internetAdapter) Resolve(_ context.Context, req source.SelectRequest) (source.PlayRequest, error) {
	/*if req.Station.URL == "" {
		return source.PlayRequest{}, errors.New("internet radio requires url")
	}
	return source.PlayRequest{
		URL:       req.Station.URL,
		Title:     req.Station.Name,
		UsePlayer: true,
	}, nil*/
	return source.PlayRequest{}, nil
}

func (a *internetAdapter) IsRadioAdapter() bool {
	return true
}

func (a *internetAdapter) getStationsByCountry(country string) []string {
	httpReq, err := http.Get(stationsByCountry + country)
	if err != nil {
		return []string{}
	}
	defer httpReq.Body.Close()
	body, err := io.ReadAll(httpReq.Body)
	//Serialize the response into a list of stations
	var s []source.Station
	json.Unmarshal(body, &s)

	return []string{}
}

func (a *internetAdapter) GetStations() map[string]source.Station {
	a.getStationsByCountry("denmark")
	return map[string]source.Station{}
}

func (a *internetAdapter) SetFavoriteStation(stationUUID string) error {
	return nil
}

func (a *internetAdapter) DefaultStation() source.Station {
	return a.defStation
}
