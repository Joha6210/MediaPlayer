package sources

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"mediaplayer/backend/internal/config"
	"mediaplayer/backend/internal/source"
)

type plexampAdapter struct {
	baseURL    string
	token      string
	httpClient *http.Client
	testMode   bool
}

func NewPlexampAdapter(cfg config.PlexampConfig, testMode bool) source.Adapter {
	return &plexampAdapter{
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		token:      cfg.Token,
		httpClient: &http.Client{Timeout: 8 * time.Second},
		testMode:   testMode,
	}
}

func (a *plexampAdapter) Resolve(ctx context.Context, req source.SelectRequest) (source.PlayRequest, error) {
	if a.testMode {
		path := strings.TrimSpace(req.Meta["path"])
		if path == "" {
			path = "/mock/plexamp/stream.mp3"
		}
		return source.PlayRequest{
			URL:       "http://127.0.0.1:65535" + path,
			Title:     "Plexamp (Mock)",
			UsePlayer: true,
		}, nil
	}

	if a.baseURL == "" {
		return source.PlayRequest{}, errors.New("plexamp baseURL is not configured")
	}
	if a.token == "" {
		return source.PlayRequest{}, errors.New("plexamp token is not configured")
	}

	streamPath := strings.TrimSpace(req.Meta["path"])
	if streamPath == "" {
		return source.PlayRequest{}, errors.New("plexamp requires meta.path")
	}

	streamURL := a.baseURL + streamPath

	checkReq, err := http.NewRequestWithContext(ctx, http.MethodHead, streamURL, nil)
	if err != nil {
		return source.PlayRequest{}, err
	}
	checkReq.Header.Set("X-Plex-Token", a.token)

	resp, err := a.httpClient.Do(checkReq)
	if err != nil {
		return source.PlayRequest{}, fmt.Errorf("checking plexamp stream availability: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		return source.PlayRequest{}, fmt.Errorf("plexamp stream unavailable: status %d", resp.StatusCode)
	}

	return source.PlayRequest{
		URL: streamURL,
		Headers: map[string]string{
			"X-Plex-Token": a.token,
		},
		Title:     "Plexamp",
		UsePlayer: true,
	}, nil
}

func (a *plexampAdapter) GetStations() []source.Station {
	return []source.Station{}
}
