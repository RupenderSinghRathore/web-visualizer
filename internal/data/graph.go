package data

type Edge struct {
	Visited int
	Status  int
	Links   []string
}

type Graph map[string]*Edge
