package certificate

import (
	"crypto/tls"
	"log"

	"golang.org/x/crypto/acme/autocert"
)

func NewTLSConfig(
	cache *CertCache,
	store Service,
	acmeManager *autocert.Manager,
) *tls.Config {
	tlsConfig := acmeManager.TLSConfig()

	tlsConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		domain := hello.ServerName
		log.Println("Received TLS handshake for domain:", domain)

		// 1️⃣ memory cache
		if cert, ok := cache.Get(domain); ok {
			return cert, nil
		}

		// 2️⃣ postgres
		if cert, err := store.Get(domain); err == nil {
			cache.Set(domain, cert)
			return cert, nil
		}

		if hello.ServerName == "" {
			log.Printf("Empty SNI, cannot obtain certificate: %s\n", domain)
			return nil, nil
		}

		// 3️⃣ let's encrypt
		log.Println("Obtaining certificate from Let's Encrypt for domain:", domain)
		log.Printf("TLS handshake SNI=%q from %s", hello.ServerName, hello.Conn.RemoteAddr())
		cert, err := acmeManager.GetCertificate(hello)
		if err != nil {
			return nil, err
		}

		cache.Set(domain, cert)
		_ = store.Save(domain, cert)

		return cert, nil
	}

	return tlsConfig
}
