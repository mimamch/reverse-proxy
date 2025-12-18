package certificate

import (
	"crypto/tls"
	"sync"
	"sync/atomic"
	"time"
)

const maxCache = 1000

type cachedCert struct {
	Certificate *tls.Certificate
	LastAccess  int64 // unix nano (atomic)
}

type CertCache struct {
	mu    sync.RWMutex
	cache map[string]*cachedCert
}

func NewCertCache() *CertCache {
	return &CertCache{
		cache: make(map[string]*cachedCert),
	}
}

func (c *CertCache) Get(domain string) (*tls.Certificate, bool) {
	now := time.Now().UnixNano()
	c.mu.RLock()
	defer c.mu.RUnlock()
	cache, ok := c.cache[domain]

	if !ok {
		return nil, false
	}
	atomic.StoreInt64(&cache.LastAccess, now)

	return cache.Certificate, ok
}

func (c *CertCache) Set(domain string, cert *tls.Certificate) {
	now := time.Now().UnixNano()

	c.mu.Lock()
	defer c.mu.Unlock()

	// evict if cache size exceeded
	if len(c.cache) >= maxCache {
		var oldestDomain string
		var oldestAccess int64 = now

		for d, cached := range c.cache {
			lastAccess := atomic.LoadInt64(&cached.LastAccess)
			if lastAccess < oldestAccess {
				oldestAccess = lastAccess
				oldestDomain = d
			}
		}

		delete(c.cache, oldestDomain)
	}

	c.cache[domain] = &cachedCert{
		Certificate: cert,
		LastAccess:  time.Now().UnixNano(),
	}
}

func (c *CertCache) Invalidate(domain string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, domain)
}
