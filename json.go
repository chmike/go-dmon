package main

import (
	"encoding/json"
	"time"
)

type monEntry struct {
	mID       int64
	Stamp     time.Time `json:"stamp"`
	Level     string    `json:"level"`
	System    string    `json:"system"`
	Component string    `json:"component"`
	Message   string    `json:"message"`
}

func (m *monEntry) MarshalJSON() ([]byte, error) {
	type Alias monEntry
	aux := &struct {
		Stamp string `json:"stamp"`
		*Alias
	}{
		Stamp: m.Stamp.Format("2006-01-02 15:04:05.000000"),
		Alias: (*Alias)(m),
	}
	return json.Marshal(&aux)
}

func (m *monEntry) UnmarshalJSON(data []byte) (err error) {
	type Alias monEntry
	aux := struct {
		Stamp string `json:"stamp"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err = json.Unmarshal(data, &aux); err != nil {
		return
	}
	m.Stamp, err = time.Parse("2006-01-02 15:04:05.000000", aux.Stamp)
	return
}
