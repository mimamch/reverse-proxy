package proxy

import "time"

type Backend struct {
	Scheme string
	Host   string
	Port   int
}

type TargetConfig struct {
	Backends   []Backend
	Headers    map[string]string
	ForceHTTPS bool
}

type SelectedTarget struct {
	Backend    Backend
	Headers    map[string]string
	ForceHTTPS bool
}

type ProxyModel struct {
	ID        string    `gorm:"primaryKey;column:id" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ProxyModel) TableName() string {
	return "proxies"
}

type HostModel struct {
	ID        string    `gorm:"primaryKey;column:id" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	ProxyID    string `gorm:"column:proxy_id" json:"proxy_id"`
	Host       string `gorm:"column:host" json:"host"`
	ForceHTTPS bool   `gorm:"column:force_https" json:"force_https"`
}

func (HostModel) TableName() string {
	return "hosts"
}

type BackendModel struct {
	ID        string    `gorm:"primaryKey;column:id" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	Scheme  string `gorm:"column:scheme" json:"scheme"`
	Host    string `gorm:"column:host" json:"host"`
	Port    int    `gorm:"column:port" json:"port"`
	ProxyID string `gorm:"column:proxy_id" json:"proxy_id"`
	Enabled bool   `gorm:"column:enabled" json:"enabled"`
}

func (BackendModel) TableName() string {
	return "backends"
}

type HeadersModel struct {
	ID        string    `gorm:"primaryKey;column:id" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	ProxyID string `gorm:"column:proxy_id" json:"proxy_id"`
	Key     string `gorm:"column:key" json:"key"`
	Value   string `gorm:"column:value" json:"value"`
}

func (HeadersModel) TableName() string {
	return "headers"
}
