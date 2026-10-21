// Package postgres contains the PostgreSQL persistence model for admin accounts.
package postgres

import "time"

const (
	StatusEnabled  = "enabled"
	StatusDisabled = "disabled"

	SexUnknown = "unknown"
	SexMale    = "male"
	SexFemale  = "female"
)

// AdminUser maps to admin_user. It is a persistence model, not an HTTP DTO.
// Constraints, indexes and identity generation must be created by goose SQL;
// this model must not be used with AutoMigrate.
type AdminUser struct {
	ID           int64   `gorm:"column:id;primaryKey;autoIncrement;<-:create"`
	Username     string  `gorm:"column:username;type:varchar(64);not null"`
	PasswordHash string  `gorm:"column:password_hash;type:text;not null" json:"-"`
	Nickname     string  `gorm:"column:nickname;type:varchar(64);not null"`
	Remark       *string `gorm:"column:remark;type:varchar(500)"`
	Email        *string `gorm:"column:email;type:varchar(254)"`
	Mobile       *string `gorm:"column:mobile;type:varchar(32)"`
	Sex          string  `gorm:"column:sex;type:varchar(16);not null;default:unknown"`
	Avatar       *string `gorm:"column:avatar;type:varchar(2048)"`
	Status       string  `gorm:"column:status;type:varchar(16);not null;default:enabled"`

	// Only successful logins update this pair; nil means no successful login yet.
	LastLoginIP *string    `gorm:"column:last_login_ip;type:inet"`
	LastLoginAt *time.Time `gorm:"column:last_login_at;type:timestamptz(3)"`

	// Defaults apply on INSERT. Updates must explicitly set UpdatedAt and UpdatedBy.
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz(3);not null;default:CURRENT_TIMESTAMP;autoCreateTime:false;<-:create"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz(3);not null;default:CURRENT_TIMESTAMP;autoUpdateTime:false"`
	CreatedBy *int64    `gorm:"column:created_by;type:bigint;<-:create"`
	UpdatedBy *int64    `gorm:"column:updated_by;type:bigint"`
}

func (AdminUser) TableName() string {
	return "admin_user"
}
