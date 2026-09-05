package sources

const (
	InternetRadioSource = "internet-radio"
	DRSource            = "dr-radio"
	PlexampSource       = "plexamp"
	BluetoothSource     = "bluetooth"
)

// Endpoints
var apiHealth = "http://all.api.radio-browser.info/json/stats"
var stationsByCountry = "http://all.api.radio-browser.info/json/stations/bycountry/"
var stationsByName = "http://all.api.radio-browser.info/json/stations/byname/"
var drStations = "http://all.api.radio-browser.info/json/stations/byname/dr%20p"
