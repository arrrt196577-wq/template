// Package postgres contains the PostgreSQL persistence model for members.
package postgres

import "time"

const (
	StatusEnabled  = "enabled"
	StatusDisabled = "disabled"

	SexUnknown = "unknown"
	SexMale    = "male"
	SexFemale  = "female"
)

// MemberUser maps to member_user and must not be exposed as an HTTP DTO.
// Schema changes and constraints belong in goose SQL, not AutoMigrate.
type MemberUser struct {
	// Account information. PasswordHash contains an encoded Argon2id hash.
	ID               int64      `gorm:"column:id;primaryKey;autoIncrement;<-:create"`
	Username         string     `gorm:"column:username;type:varchar(64);not null"`
	Mobile           *string    `gorm:"column:mobile;type:varchar(32)"`
	Email            *string    `gorm:"column:email;type:varchar(254)"`
	PasswordHash     string     `gorm:"column:password_hash;type:text;not null" json:"-"`
	Status           string     `gorm:"column:status;type:varchar(16);not null;default:enabled"`
	RegisterIP       *string    `gorm:"column:register_ip;type:inet;<-:create"`
	RegisterTerminal *int32     `gorm:"column:register_terminal;type:integer;<-:create"`
	LastLoginIP      *string    `gorm:"column:last_login_ip;type:inet"`
	LastLoginAt      *time.Time `gorm:"column:last_login_at;type:timestamptz(3)"`

	// Profile information. Birthday is a calendar date, not a timestamp.
	Nickname string     `gorm:"column:nickname;type:varchar(64);not null"`
	Avatar   *string    `gorm:"column:avatar;type:varchar(2048)"`
	Name     *string    `gorm:"column:name;type:varchar(64)"`
	Sex      string     `gorm:"column:sex;type:varchar(16);not null;default:unknown"`
	Birthday *time.Time `gorm:"column:birthday;type:date"`
	AreaID   *int64     `gorm:"column:area_id;type:bigint"`
	Mark     *string    `gorm:"column:mark;type:varchar(500)"`

	// Membership data is reserved here; balance changes need separate business rules.
	Point      int64   `gorm:"column:point;type:bigint;not null;default:0"`
	TagIDs     []int64 `gorm:"column:tag_ids;type:jsonb;serializer:json;not null;default:'[]'"`
	LevelID    *int64  `gorm:"column:level_id;type:bigint"`
	Experience int64   `gorm:"column:experience;type:bigint;not null;default:0"`
	GroupID    *int64  `gorm:"column:group_id;type:bigint"`

	// Defaults apply on INSERT; subsequent updates must set audit fields explicitly.
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz(3);not null;default:CURRENT_TIMESTAMP;autoCreateTime:false;<-:create"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz(3);not null;default:CURRENT_TIMESTAMP;autoUpdateTime:false"`
	CreatedBy *int64    `gorm:"column:created_by;type:bigint;<-:create"`
	UpdatedBy *int64    `gorm:"column:updated_by;type:bigint"`
}

func (MemberUser) TableName() string {
	return "member_user"
}
