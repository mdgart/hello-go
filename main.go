package main

import (
    "encoding/json"
    "net/http"
    "strings"
	"log"
	"time"
)

func main()  {

	mw := multiWeatherProvider{
		openWeatherMap{},
		weatherUnderground{apiKey: "fba4e4d4f75effb3"},
	}
	
	http.HandleFunc("/weathr/", func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		city := strings.SplitN(r.URL.Path, "/", 3)[2]
		
		temp, err := mw.temperature(city)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"city": city,
			"temp": temp,
			"took": time.Since(begin).String(),
		})

	})
	http.ListenAndServe(":8080", nil)
}


type weatherProvider interface {
	temperature(city string) (float64, error) // in Kelvin
}

type openWeatherMap struct{}

func (w openWeatherMap) temperature(city string) (float64, error) {
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=32dba4e407b670866d49b5a99cf56b29&q=" + city)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var d struct {
		Main struct {
			Kelvin float64 `json:"temp"`
		} `json:"main"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}

	log.Printf("openWeatherMap: %s: %.2f", city, d.Main.Kelvin)
	return d.Main.Kelvin, nil
}

type weatherUnderground struct {
	apiKey string
}

func (w weatherUnderground) temperature(city string) (float64, error) {
	resp, err := http.Get("http://api.wunderground.com/api/" + w.apiKey + "/conditions/q/" + city + ".json")
    if err != nil {
        return 0, err
    }

	defer resp.Body.Close()

	var d struct {
        Observation struct {
            Celsius float64 `json:"temp_c"`
        } `json:"current_observation"`
    }

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}

	kelvin := d.Observation.Celsius + 273.15
    log.Printf("weatherUnderground: %s: %.2f", city, kelvin)
    return kelvin, nil
}

func temperature(city string, providers ...weatherProvider) (float64, error) {
	sum := 0.0

	for _, provider := range providers {
		k, err := provider.temperature(city)
		if err != nil {
			return 0, err
		}

		sum += k
	}

	return sum / float64(len(providers)), nil
}

type multiWeatherProvider []weatherProvider

func (w multiWeatherProvider) temperature(city string) (float64, error) {
	sum := 0.0

    for _, provider := range w {
        k, err := provider.temperature(city)
        if err != nil {
            return 0, err
        }

        sum += k
    }

    return sum / float64(len(w)), nil
}

