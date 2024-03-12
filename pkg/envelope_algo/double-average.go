package envelope_algo

import (
	"math/rand"
	"time"
)

func DoubleAverage(count, amount, min int64) int64 {
	if count == 1 {
		return amount
	}

	max := amount - min*count

	avg := max / count
	avg2 := 2*avg + min
	rand.Seed(time.Now().UnixNano())
	x := rand.Int63n(avg2) + min
	return x
}
