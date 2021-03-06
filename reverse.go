package pb_fuzz_workshop

func Reverse(s []int) []int {
	if s == nil {
		return nil
	}

	a := make([]int, len(s))
	copy(a, s)

	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}

	return a
}

func NotReverse(s []int) []int {
	return s
}
