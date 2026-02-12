package data

type Edge struct {
	Visited int
	Status  int
	Links   map[string]struct{}
}

type Graph map[string]*Edge
