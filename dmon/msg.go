package dmon

import (
	"encoding/json"
	"time"
)

type Msg struct {
	ID        int64     `json:"id"`
	Stamp     time.Time `json:"stamp"`
	Level     string    `json:"level"`
	System    string    `json:"system"`
	Component string    `json:"component"`
	Message   string    `json:"message"`
}

func (m *Msg) MarshalJSON() ([]byte, error) {
	type Alias Msg
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	return json.Marshal(&aux)
}

func (m *Msg) UnmarshalJSON(data []byte) (err error) {
	type Alias Msg
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err = json.Unmarshal(data, &aux); err != nil {
		return
	}
	return
}
