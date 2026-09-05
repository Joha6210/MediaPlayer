package sources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"mediaplayer/backend/internal/source"
)

type DRRadioObject struct {
	Name     string   `json:"name"`
	Stations []string `json:"stations"`
}

type drAdapter struct {
	httpClient *http.Client
	testMode   bool
	drStations map[string]source.Station
}

func NewDRAdapter(testMode bool) source.Adapter {
	return &drAdapter{
		httpClient: &http.Client{Timeout: 8 * time.Second},
		testMode:   testMode,
		drStations: mapStations(getDRStations()),
	}
}

func mapStations(stations []source.Station) map[string]source.Station {
	m := make(map[string]source.Station)
	for _, s := range stations {
		m[s.Name] = s
	}
	return m
}

func getDRStations() []source.Station {
	httpReq, err := http.Get(drStations)
	if err != nil {
		return []source.Station{}
	}
	defer httpReq.Body.Close()
	body, err := io.ReadAll(httpReq.Body)
	//Serialize the response into a list of stations
	var rawStations []source.Station
	json.Unmarshal(body, &rawStations)

	sort.Slice(rawStations, func(i, j int) bool {
		return rawStations[i].Votes > rawStations[j].Votes
	})

	var filteredStations []source.Station
	for _, station := range rawStations {
		if station.Codec == "MP3" {
			filteredStations = append(filteredStations, station)
		}
	}

	return filteredStations
}

func (a *drAdapter) GetStations() []source.Station {
	return getDRStations()
}

func (a *drAdapter) Resolve(ctx context.Context, req source.SelectRequest) (source.PlayRequest, error) {

	station := req.Station.Name

	if _, exists := a.drStations[station]; !exists {
		return source.PlayRequest{}, errors.New("dr-radio station not found")
	}

	resolvedURL := req.Station.URL
	if resolvedURL == "" {
		resolvedURL = a.drStations[station].URL
	}

	parsed, err := url.Parse(resolvedURL)
	if err != nil {
		return source.PlayRequest{}, fmt.Errorf("invalid dr url: %w", err)
	}

	if !a.testMode {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodHead, parsed.String(), nil)
		if err != nil {
			return source.PlayRequest{}, err
		}

		resp, err := a.httpClient.Do(httpReq)
		if err != nil {
			return source.PlayRequest{}, fmt.Errorf("checking dr stream availability: %w", err)
		}
		resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode > 399 {
			return source.PlayRequest{}, fmt.Errorf("dr stream unavailable: status %d", resp.StatusCode)
		}
	}

	return source.PlayRequest{
		URL:       parsed.String(),
		Title:     "DR " + strings.ToUpper(station),
		UsePlayer: true,
	}, nil
}
