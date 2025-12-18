package proxy

import (
	"strings"

	"gorm.io/gorm"
)

type Repository interface {
	GetTargetConfig(domain string) (*TargetConfig, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetTargetConfig(domain string) (*TargetConfig, error) {

	incomingHost := strings.ToLower(domain)

	var host HostModel

	if strings.Contains(incomingHost, "*") {
		incomingHost = strings.ReplaceAll(incomingHost, "*", "%")
		if err := r.db.Where("host LIKE ?", incomingHost).First(&host).Error; err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Where("host = ?", incomingHost).First(&host).Error; err != nil {
			return nil, err
		}
	}

	// if host not found, return nil
	if host.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var backends []BackendModel
	var headers []HeadersModel
	errChan := make(chan error, 2)

	go func() {
		errChan <- r.db.Where("proxy_id = ?", host.ProxyID).Where("enabled = ?", true).Find(&backends).Error
	}()

	go func() {
		errChan <- r.db.Where("proxy_id = ?", host.ProxyID).Find(&headers).Error
	}()

	for range 2 {
		if err := <-errChan; err != nil {
			return nil, err
		}
	}

	backendsObj := []Backend{}
	for _, backend := range backends {
		backendsObj = append(backendsObj, Backend{
			Scheme: backend.Scheme,
			Host:   backend.Host,
			Port:   backend.Port,
		})
	}

	headersMap := make(map[string]string)
	for _, header := range headers {
		headersMap[header.Key] = header.Value
	}

	return &TargetConfig{
		Backends:   backendsObj,
		Headers:    headersMap,
		ForceHTTPS: host.ForceHTTPS,
	}, nil
}
