package consensus

import (
	"github.com/golang/glog"
	"math/rand"
)

func mod(x int, y int) int {
	if x < y {
		return x
	}
	dif := x - y
	if dif < y {
		return dif
	}
	return mod(dif, y)
}

func next(view int, id int, n int) int {
	round := view / n
	return (round+1)*n + id
}

func randPeer(n int, id int) int {
	if n == 1 {
		glog.Fatal("No peers present")
	}
	c := rand.Intn(n)
	for c == id {
		c = rand.Intn(n)
	}
	return c
}
