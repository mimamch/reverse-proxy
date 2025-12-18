package proxy

import (
	"sync"
	"sync/atomic"
	"time"
)

type cachedConfig struct {
	Route      *TargetConfig
	ExpiryTime int64 // unix nano
	NextIdx    uint64
	LastAccess int64 // unix nano (atomic)
}

const MaxCacheSize = 1000

type ProxyCache struct {
	routeCache map[string]*cachedConfig
	mutex      sync.RWMutex
	cacheTTL   time.Duration
	maxSize    int
}

func NewProxyCache() *ProxyCache {
	return &ProxyCache{
		routeCache: make(map[string]*cachedConfig),
		cacheTTL:   1 * time.Hour,
		maxSize:    MaxCacheSize,
	}
}

// ========================
// GET (NO WRITE LOCK)
// ========================
func (c *ProxyCache) Get(domain string) (*cachedConfig, bool) {
	c.mutex.RLock()
	config, found := c.routeCache[domain]
	c.mutex.RUnlock()

	if !found {
		return nil, false
	}

	now := time.Now().UnixNano()

	// cek expiry
	if atomic.LoadInt64(&config.ExpiryTime) < now {
		c.mutex.Lock()
		// double check
		if cfg, ok := c.routeCache[domain]; ok {
			if atomic.LoadInt64(&cfg.ExpiryTime) < now {
				delete(c.routeCache, domain)
			}
		}
		c.mutex.Unlock()
		return nil, false
	}

	// update LastAccess TANPA LOCK
	atomic.StoreInt64(&config.LastAccess, now)

	if config.NextIdx == ^uint64(0) { // overflow check
		atomic.StoreUint64(&config.NextIdx, 0)
	} else {
		atomic.AddUint64(&config.NextIdx, 1)
	}

	return config, true
}

// ========================
// SET
// ========================
func (c *ProxyCache) Set(domain string, config *TargetConfig) {
	now := time.Now().UnixNano()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// evict jika penuh
	if len(c.routeCache) >= c.maxSize {
		c.evictLRU()
	}

	c.routeCache[domain] = &cachedConfig{
		Route:      config,
		ExpiryTime: now + c.cacheTTL.Nanoseconds(),
		NextIdx:    0,
		LastAccess: now,
	}
}

// ========================
// LRU EVICTION (BEST-EFFORT)
// ========================
func (c *ProxyCache) evictLRU() {
	var oldestKey string
	var oldestTime int64 = time.Now().UnixNano()

	for key, cfg := range c.routeCache {
		lastAccess := atomic.LoadInt64(&cfg.LastAccess)
		if lastAccess < oldestTime {
			oldestTime = lastAccess
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(c.routeCache, oldestKey)
	}
}

// ========================
// OPTIONAL: CLEANUP EXPIRED
// ========================
func (c *ProxyCache) CleanupExpired() {
	now := time.Now().UnixNano()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key, cfg := range c.routeCache {
		if atomic.LoadInt64(&cfg.ExpiryTime) < now {
			delete(c.routeCache, key)
		}
	}
}

func (c *ProxyCache) Invalidate(domain string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.routeCache, domain)
}
