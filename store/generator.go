package store

import (
	"fmt"
	"github.com/heidi-ann/hydra/config"
	"math/rand"
)

// Generator generates workloads for the store
// Store has 10 keys
type Generator struct {
	Ratio    int // percentage of read requests
	Conflict int // 1 to 5, degree of requests which target particular area
}

func Generate(conf config.ConfigAuto) Generator {
	return Generator{conf.Commands.Reads, conf.Commands.Conflicts}
}

func (g Generator) Next() string {

	key := "A"

	// determine which key to operate on
	// range 0-9
	if rand.Intn(5) < g.Conflict-1 {
		// non-conflicted region
		// range conflict to 9
		key = string(9 - rand.Intn(10-g.Conflict))
	} else {
		// conflicted region
		// range 0 to (conflict-1)
		key = string(rand.Intn(g.Conflict))
	}

	if rand.Intn(100) < g.Ratio {
		return fmt.Sprintf("get %s\n", key)
	} else {
		return fmt.Sprintf("update %s 7\n", key)
	}
}
