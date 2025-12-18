package certificate

import "crypto/tls"

type Service interface {
	Get(host string) (*tls.Certificate, error)
	Save(host string, cert *tls.Certificate) error
}

type service struct {
	repo  Repository
	cache *CertCache
}

func NewService(repo Repository, cache *CertCache) Service {
	return &service{
		repo:  repo,
		cache: cache,
	}
}

func (s *service) Get(host string) (*tls.Certificate, error) {
	if cert, ok := s.cache.Get(host); ok {
		return cert, nil
	}

	cert, err := s.repo.Get(host)
	if err != nil {
		return nil, err
	}

	s.cache.Set(host, cert)
	return cert, nil
}

func (s *service) Save(host string, cert *tls.Certificate) error {
	err := s.repo.Save(host, cert)
	if err != nil {
		return err
	}

	s.cache.Set(host, cert)
	return nil
}
