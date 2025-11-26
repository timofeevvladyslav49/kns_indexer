package models

import "time"

type Username struct {
	Username  string    `json:"username" example:"username" db:"username"`
	Address   string    `json:"address" example:"keeta_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" db:"address"`
	Owner     string    `json:"owner" example:"keeta_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" db:"owner"`
	CID       *string   `json:"cid,omitempty" example:"Qmaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" db:"cid"`
	IsPrimary bool      `json:"isPrimary" example:"false" db:"is_primary"`
	Timestamp time.Time `json:"timestamp" example:"2025-11-25T11:22:33.123Z" db:"timestamp"`
}
