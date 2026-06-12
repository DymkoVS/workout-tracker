// Package ratelimit — простой in-memory ограничитель попыток (fixed window).
// Используется для защиты логина от перебора. Состояние живёт в памяти
// процесса: после рестарта счётчики обнуляются — для одного инстанса этого
// достаточно.
package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	mu       sync.Mutex
	max      int
	window   time.Duration
	attempts map[string][]time.Time
}

func New(max int, window time.Duration) *Limiter {
	l := &Limiter{
		max:      max,
		window:   window,
		attempts: make(map[string][]time.Time),
	}
	go l.cleanupLoop()
	return l
}

// Allow регистрирует попытку и сообщает, не превышен ли лимит для ключа.
func (l *Limiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	fresh := prune(l.attempts[key], now.Add(-l.window))
	if len(fresh) >= l.max {
		l.attempts[key] = fresh
		return false
	}
	l.attempts[key] = append(fresh, now)
	return true
}

// Reset сбрасывает счётчик (например, после успешного логина).
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}

func prune(ts []time.Time, cutoff time.Time) []time.Time {
	out := ts[:0]
	for _, t := range ts {
		if t.After(cutoff) {
			out = append(out, t)
		}
	}
	return out
}

// cleanupLoop периодически выкидывает протухшие ключи, чтобы map не рос вечно.
func (l *Limiter) cleanupLoop() {
	for range time.Tick(10 * time.Minute) {
		cutoff := time.Now().Add(-l.window)
		l.mu.Lock()
		for k, ts := range l.attempts {
			if fresh := prune(ts, cutoff); len(fresh) == 0 {
				delete(l.attempts, k)
			} else {
				l.attempts[k] = fresh
			}
		}
		l.mu.Unlock()
	}
}
