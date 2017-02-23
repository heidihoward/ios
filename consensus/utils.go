package consensus

func Majority(n int) int {
	return (n / 2) + 1
}

func mod(x int, y int) int {
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
