package envelope_algo

import (
	"math/rand"
	"time"
)

func DoubleRandom(count, amount int64) int64 {
	if count == 1 {
		return amount
	}
	max := amount - min*count
	rand.Seed(time.Now().UnixNano())
	seed := rand.Int63n(count*2) + 1
	n := max/seed + min
	x := rand.Int63n(n)
	return x + min
}
