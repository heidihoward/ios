package store

import (
	"math/rand"
)

type Generator struct {
	Ratio    int // percentage of read requests
	Conflict int // percentage of requests which target particular area
}

func Generate(ratio int, conflict int) Generator {
	return Generator{ratio, conflict}
}

func (g Generator) Next() string {
	if rand.Intn(100) < g.Ratio {
		return "get A\n"
	} else {
		return "update A B\n"
	}
}
