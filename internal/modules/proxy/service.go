package proxy

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Service interface {
	GetTarget(domain string) (*SelectedTarget, error)
}

type service struct {
	repository Repository
	proxyCache *ProxyCache
}

func NewService(repository Repository, proxyCache *ProxyCache) Service {
	return &service{
		repository: repository,
		proxyCache: proxyCache,
	}
}

func (s *service) GetTarget(domain string) (*SelectedTarget, error) {
	currentTime := time.Now().UnixNano()

	config, cacheFound := s.proxyCache.Get(domain)
	var route *TargetConfig
	var nextIdx uint64
	if cacheFound {
		route = config.Route
		nextIdx = config.NextIdx

		if config.Route == nil {
			return nil, ErrNoRouteFound // cached but route is nil
		}
	}

	if !cacheFound || currentTime > config.ExpiryTime {
		configFromDB, err := s.repository.GetTargetConfig(domain)
		if err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.proxyCache.Set(domain, nil)
				return nil, ErrNoRouteFound
			}

			return &SelectedTarget{}, err
		}
		s.proxyCache.Set(domain, configFromDB)
		route = configFromDB
	}

	backends := route.Backends
	numBackends := len(backends)
	if numBackends == 0 {
		return nil, ErrNoRouteFound
	}

	backendIndex := (nextIdx) % uint64(numBackends)

	chosenBackend := backends[backendIndex]
	if len(route.Headers) == 0 {
		route.Headers = make(map[string]string)
	}

	return &SelectedTarget{
		Backend:    chosenBackend,
		Headers:    route.Headers,
		ForceHTTPS: route.ForceHTTPS,
	}, nil
}
