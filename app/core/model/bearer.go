package model

import (
	"time"
)

type TBearer struct {
	Bearer     string    `json:"bearer"`
	UserId     int64     `json:"user_id"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}
