package consensus

func mod(x int, y int) int {
	if x < y {
		return x
	}
	dif := x - y
	if dif < y {
		return dif
	} else {
		return mod(dif, y)
	}
}

func next(view int, id int, n int) int {
	round := view/n
	return (round+1)*n + id
}
