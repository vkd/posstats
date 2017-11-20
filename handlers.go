package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type posJSON struct {
	App struct {
		Bundle string `json:"bundle"`
	} `json:"app"`
	Device struct {
		Geo struct {
			Country string `json:"country"`
		} `json:"geo"`
		OS  string `json:"os"`
		Ifa string `json:"ifa"`
	} `json:"device"`
}

// PosHandler - handler for calc current pos
func PosHandler(counter PosCounter, stats StatsCollector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var j posJSON
		var err error
		if r.Body == nil {
			err = errors.New("'body' is empty")
		} else {
			err = json.NewDecoder(r.Body).Decode(&j)
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("Bad request: %v", err), http.StatusBadRequest)
			return
		}
		if j.Device.Ifa == "" {
			http.Error(w, fmt.Sprintf("Bad request: 'ifa' is empty"), http.StatusBadRequest)
			return
		}

		pos, err := counter.Pos(j.Device.Ifa)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error on get pos: %v", err), http.StatusInternalServerError)
			return
		}

		err = stats.Add(j.App.Bundle, j.Device.OS, j.Device.Geo.Country)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error on add stats: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"pos": ` + strconv.Itoa(pos) + "}\n")) // TODO check error
	}
}

// PosCounter - interface for calc current pos
type PosCounter interface {
	Pos(ifa string) (int, error)
}

// StatsCollector - interface for storage stats about requests
type StatsCollector interface {
	Add(app, platform, country string) error
}

// StatsHandler - handler for return current stats about previous requests
func StatsHandler(ss StatsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := ss.GetAll()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error on get stats: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats) // TODO check error
	}
}

// StatsGetter - interface for return current stats
type StatsGetter interface {
	GetAll() ([]*Stat, error)
}
