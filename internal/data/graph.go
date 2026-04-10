package data

import (
	"encoding/json"
)

type Edge struct {
	Visited int      `json:"visited"`
	Status  int      `json:"status"`
	Depth   int      `json:"depth"`
	Links   []string `json:"links"`
}

type Graph map[string]*Edge

func (e Graph) String() string {
	data, _ := json.MarshalIndent(e, "", "\t")
	return string(data)
}
