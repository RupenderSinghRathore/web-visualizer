package data

type Edge struct {
	Visited int      `json:"visited"`
	Status  int      `json:"Status"`
	Links   []string `json:"links"`
}

type Graph map[string]*Edge
