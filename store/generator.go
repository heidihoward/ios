package store

import (
	"math/rand"
)

type Generator struct {
	Ratio int // percentage of read requests
}

func Generate(i int) Generator {
	return Generator{i}
}

func (g Generator) Next() string {
	if rand.Intn(100) < g.Ratio {
		return "get A\n"
	} else {
		return "update A B\n"
	}
}
