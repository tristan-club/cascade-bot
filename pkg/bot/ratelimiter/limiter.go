package ratelimiter

import (
	"context"
	"golang.org/x/time/rate"
	"sync"
)

type IdLimiter struct {
	keys map[int64]*rate.Limiter
	mu   *sync.RWMutex
	r    rate.Limit
	b    int
}

func NewIdRateLimiter(r rate.Limit, b int) *IdLimiter {
	i := &IdLimiter{
		keys: make(map[int64]*rate.Limiter),
		mu:   &sync.RWMutex{},
		r:    r,
		b:    b,
	}

	return i
}

// Add creates a new rate limiter and adds it to the keys map,
// using the key
func (i *IdLimiter) Add(key int64) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)

	i.keys[key] = limiter

	return limiter
}

// GetLimiter returns the rate limiter for the provided key if it exists.
// Otherwise, calls Add to add key address to the map
func (i *IdLimiter) GetLimiter(key int64) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.keys[key]

	if !exists {
		i.mu.Unlock()
		return i.Add(key)
	}

	i.mu.Unlock()

	return limiter
}

func (i *IdLimiter) Wait(chatId int64) error {
	return i.GetLimiter(chatId).Wait(context.Background())
}
