package main

import (
	"encoding/json"
	"time"
)

const (
	// tfmt = "2006-01-02T15:04:05.000000"
	tfmt = time.RFC3339Nano
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
		Stamp: m.Stamp.Format(tfmt),
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
	m.Stamp, err = time.Parse(tfmt, aux.Stamp)
	return
}
