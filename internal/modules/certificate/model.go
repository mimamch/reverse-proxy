package certificate

import "time"

type Certificate struct {
	ID        string    `gorm:"primaryKey;column:id" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	HostID    string    `gorm:"column:host_id" json:"host_id"`

	Cert      string    `gorm:"column:cert" json:"cert"`
	Key       string    `gorm:"column:key" json:"key"`
	ExpiresAt time.Time `gorm:"column:expires_at" json:"expires_at"`
}

func (Certificate) TableName() string {
	return "certificates"
}
