package certificate

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/mimamch/reverse-proxy/internal/modules/proxy"
	"github.com/nrednav/cuid2"
	"gorm.io/gorm"
)

type Repository interface {
	Get(host string) (*tls.Certificate, error)
	Save(host string, cert *tls.Certificate) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Get(host string) (*tls.Certificate, error) {
	row := r.db.Table("certificates").Select("cert, key, expires_at").
		Joins("JOIN hosts ON certificates.host_id = hosts.id").
		Where("hosts.host = ?", host).
		Row()

	var certPEM, keyPEM []byte
	var expires time.Time

	if err := row.Scan(&certPEM, &keyPEM, &expires); err != nil {
		return nil, err
	}

	if time.Now().After(expires) {
		return nil, errors.New("certificate expired")
	}

	tls, error := tls.X509KeyPair(certPEM, keyPEM)
	if error != nil {
		return nil, error
	}

	return &tls, nil
}

func (r *repository) Save(host string, cert *tls.Certificate) error {
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Certificate[0],
	})
	keyBytes, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		return err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})

	var hostDb proxy.HostModel
	if err := r.db.Where("host = ?", host).First(&hostDb).Error; err != nil {
		return err
	}

	leaf := cert.Leaf

	if leaf == nil {
		var err error
		leaf, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return err
		}
	}

	expires := leaf.NotAfter

	err = r.db.Create(&Certificate{
		ID:        cuid2.Generate(),
		HostID:    hostDb.ID,
		Key:       string(keyPEM),
		Cert:      string(certPEM),
		ExpiresAt: expires,
	}).Error

	return err
}
