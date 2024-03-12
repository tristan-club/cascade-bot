package envelope_algo

import (
	"math/rand"
	"time"
)

func BeforeShuffle(count, amount int64) int64 {
	if count == 1 {
		return amount
	}
	max := amount - min*count

	seeds := make([]int64, 0)

	size := count / 2
	if size < 3 {
		size = 3
	}
	for i := int64(0); i < size; i++ {
		x := max / (i + 1)
		seeds = append(seeds, x)
	}
	rand.Seed(time.Now().UnixNano())
	idx := rand.Int63n(int64(len(seeds)))
	x := rand.Int63n(seeds[idx])
	return x + min
}
