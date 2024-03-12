package envelope_algo

import "math/rand"

func AfterShuffle(count, amount int64) []int64 {
	inds := make([]int64, 0)
	max := amount - min*count
	remain := max
	for i := int64(0); i < count; i++ {
		x := SimpleRand(count-i, remain)
		remain -= x
		inds = append(inds, x)
	}

	rand.Shuffle(len(inds), func(i, j int) {
		inds[i], inds[j] = inds[j], inds[i]
	})
	return inds

}
