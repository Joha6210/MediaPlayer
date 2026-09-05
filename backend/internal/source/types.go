package source

import (
	"context"
)

type Station struct {
	Changeuuid                 string      `json:"changeuuid"`
	Stationuuid                string      `json:"stationuuid"`
	Serveruuid                 string      `json:"serveruuid"`
	Name                       string      `json:"name"`
	URL                        string      `json:"url"`
	URL_resolved               string      `json:"url_resolved"`
	Homepage                   string      `json:"homepage"`
	Favicon                    string      `json:"favicon"`
	Tags                       string      `json:"tags"`
	Country                    string      `json:"country"`
	Countrycode                string      `json:"countrycode"`
	Iso_3166_2                 string      `json:"iso_3166_2"`
	State                      string      `json:"state"`
	Language                   string      `json:"language"`
	Languagecodes              string      `json:"languagecodes"`
	Votes                      int         `json:"votes"`
	Lastchangetime             string      `json:"lastchangetime"`
	Lastchangetime_iso8601     string      `json:"lastchangetime_iso8601"`
	Codec                      string      `json:"codec"`
	Bitrate                    int         `json:"bitrate"`
	Hls                        int         `json:"hls"`
	Lastcheckok                int         `json:"lastcheckok"`
	Lastchecktime              string      `json:"lastchecktime"`
	Lastchecktime_iso8601      string      `json:"lastchecktime_iso8601"`
	Lastcheckoktime            string      `json:"lastcheckoktime"`
	Lastcheckoktime_iso8601    string      `json:"lastcheckoktime_iso8601"`
	Lastlocalchecktime         string      `json:"lastlocalchecktime"`
	Lastlocalchecktime_iso8601 string      `json:"lastlocalchecktime_iso8601"`
	Clicktimestamp             string      `json:"clicktimestamp"`
	Clicktimestamp_iso8601     string      `json:"clicktimestamp_iso8601"`
	Clickcount                 int         `json:"clickcount"`
	Clicktrend                 int         `json:"clicktrend"`
	SSL_error                  int         `json:"ssl_error"`
	Geo_lat                    interface{} `json:"geo_lat"`      // Can be null
	Geo_long                   interface{} `json:"geo_long"`     // Can be null
	Geo_distance               interface{} `json:"geo_distance"` // Can be null
	Has_extended_info          bool        `json:"has_extended_info"`
}

type Adapter interface {
	Resolve(ctx context.Context, req SelectRequest) (PlayRequest, error)
	GetStations() []Station
}

type Player interface {
	Play(url string, volume int, headers map[string]string) error
	Stop() error
	Pause(paused bool) error
	SetVolume(volume int) error
	ListenEvents() (<-chan struct {
		Title  string
		Paused bool
	}, error)
}

type SourceState struct {
	ActiveSource  string  `json:"activeSource"`
	Label         string  `json:"label"`
	Playing       bool    `json:"playing"`
	Volume        int     `json:"volume"`
	StreamTitle   string  `json:"stream_title"`
	Paused        bool    `json:"paused"`
	ActiveStation Station `json:"station,omitempty"`
}

type SelectRequest struct {
	Source  string            `json:"source"`
	Station Station           `json:"station,omitempty"`
	Title   string            `json:"title,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

type PlayRequest struct {
	URL       string
	Title     string
	Headers   map[string]string
	UsePlayer bool
}

type SourceDefaultsConfig struct {
	DefaultVolume int `yaml:"defaultVolume"`
}
