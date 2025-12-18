package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/mimamch/reverse-proxy/internal/config"
	"github.com/mimamch/reverse-proxy/internal/modules/proxy"
	"github.com/mimamch/reverse-proxy/pkg/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func generateTestCertificate(t *testing.T) (*tls.Certificate, *rsa.PrivateKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test.example.com"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	assert.NoError(t, err)

	tlsCert := tls.Certificate{
		Certificate: [][]byte{certBytes},
		PrivateKey:  privateKey,
		Leaf:        cert,
	}

	return &tlsCert, privateKey
}

func TestRepositorySave(t *testing.T) {
	db := setupTestDB()
	repo := NewRepository(db)

	host := "test.example.com"
	// hostModel := proxy.HostModel{ID: cuid2.Generate(), Host: host, ProxyID: "1"}
	// db.Create(&hostModel)

	cert, _ := generateTestCertificate(t)

	err := repo.Save(host, cert)
	assert.NoError(t, err)
}

func TestRepositoryGet(t *testing.T) {
	db := setupTestDB()
	repo := NewRepository(db)

	host := "test.example.com"
	// hostModel := proxy.HostModel{ID: "1", Host: host}
	// db.Create(&hostModel)

	cert, _ := generateTestCertificate(t)
	err := repo.Save(host, cert)
	assert.NoError(t, err)

	retrieved, err := repo.Get(host)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
}

func TestRepositoryGetExpired(t *testing.T) {
	db := setupTestDB()
	repo := NewRepository(db)

	host := "test.example.com"
	db.Create(&proxy.HostModel{ID: "1", Host: host})

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: host},
		NotBefore:    time.Now().AddDate(-1, 0, 0),
		NotAfter:     time.Now().AddDate(-1, 0, 1),
	}

	certBytes, _ := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	tlsCert := &tls.Certificate{Certificate: [][]byte{certBytes}, PrivateKey: privateKey, Leaf: cert}

	repo.Save(host, tlsCert)
	_, err := repo.Get(host)
	assert.Error(t, err)
	assert.Equal(t, "certificate expired", err.Error())
}

func setupTestDB() *gorm.DB {
	cfg := config.LoadConfig()
	db := database.ConnectPostgres(cfg)
	return db
}
