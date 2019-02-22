package dmon

import "time"

// Msg is a monitoring log meessage.
type Msg struct {
	Stamp     time.Time `json:"stamp"`
	Level     string    `json:"level"`
	System    string    `json:"system"`
	Component string    `json:"component"`
	Message   string    `json:"message"`
}
