package sources

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mediaplayer/backend/internal/source"
	"net/http"
)

type internetAdapter struct{}

func NewInternetAdapter() source.Adapter {
	return &internetAdapter{}
}

func (a *internetAdapter) Resolve(_ context.Context, req source.SelectRequest) (source.PlayRequest, error) {
	if req.Station.URL == "" {
		return source.PlayRequest{}, errors.New("internet radio requires url")
	}
	return source.PlayRequest{
		URL:       req.Station.URL,
		Title:     req.Station.Name,
		UsePlayer: true,
	}, nil
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

func (a *internetAdapter) GetStations() []source.Station {
	a.getStationsByCountry("denmark")
	return []source.Station{}
}
