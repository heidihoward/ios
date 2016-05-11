package test

import (
	"fmt"
	"github.com/golang/glog"
	"math/rand"
	"strconv"
	"time"
)

// Generator generates workloads for the store
// Store has 10 keys
type Generator struct {
	Ratio    int // percentage of read requests
	Conflict int // 1 to 5, degree of requests which target particular area
	Requests int // terminate after this number of requests
	Interval int // milliseconand delay between client resquest and response
}

func Generate(conf ConfigAuto) *Generator {
	return &Generator{conf.Commands.Reads, conf.Commands.Conflicts, conf.Termination.Requests, conf.Commands.Interval}
}

func (g *Generator) Next() (string, bool) {

	//handle termination after n requests
	if g.Requests == 0 {
		return "", false
	}
	g.Requests--

	delay := rand.Intn(g.Interval)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// generate key
	key := "A" // default just in case
	glog.Info("Starting to generate command")

	// determine which key to operate on
	// range 0-9
	if rand.Intn(5) < g.Conflict-1 {
		// non-conflicted region
		// range conflict to 9
		key = strconv.Itoa(9 - rand.Intn(10-g.Conflict))
	} else {
		// conflicted region
		// range 0 to (conflict-1)
		key = strconv.Itoa(rand.Intn(g.Conflict))
	}
	glog.Info("Key is", key)

	if rand.Intn(100) < g.Ratio {
		return fmt.Sprintf("get %s", key), true
	} else {
		return fmt.Sprintf("update %s 7", key), true
	}
}

func (_ *Generator) Return(_ string) {
	//STUB
}
