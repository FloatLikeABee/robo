package mcp

import (
	"sync"
	"time"
)

var (
	rateMu    sync.Mutex
	rateHits  = map[string][]time.Time{}
	rateLimit = map[string]struct {
		N int
		W time.Duration
	}{
		"share_note":   {20, 24 * time.Hour},
		"pool_search":  {60, time.Minute},
		"pool_get":     {300, time.Hour},
		"pool_report":  {20, 24 * time.Hour},
	}
)

func AllowRate(user, action string) bool {
	rateMu.Lock()
	defer rateMu.Unlock()
	lim, ok := rateLimit[action]
	if !ok {
		return true
	}
	key := user + "\x00" + action
	now := time.Now()
	cut := now.Add(-lim.W)
	kept := rateHits[key][:0]
	for _, t := range rateHits[key] {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= lim.N {
		rateHits[key] = kept
		return false
	}
	rateHits[key] = append(kept, now)
	return true
}

func resetRates() {
	rateMu.Lock()
	rateHits = map[string][]time.Time{}
	rateMu.Unlock()
}
