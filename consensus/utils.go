package consensus

func Majority(n int) int {
	return (n / 2) + 1
}

func mod(x int, y int) int {
	dif := x - y
	if dif < y {
		return x
	} else {
		return mod(dif, y)
	}
}
