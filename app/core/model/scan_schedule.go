package model

import (
	"time"
)

type TScanScheduler struct {
	Bearer string `json:"bearer"`
	UserId int64  `json:"user_id"`
	ScanId string `json:"scan_id"`
	Count  int64  `json:"count"`

	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}
