package model

import (
	"time"
)

type TScan struct {
	Id              int64  `json:"id"`
	Bearer          string `json:"bearer"`
	UserId          int64  `json:"user_id"`
	ScanId          string `json:"scan_id"`
	ScanResult      string `json:"scan_result"`
	AvailableResult string `json:"available_results"`
	HeadPaths       string `json:"head_paths"`
	BodyPaths       string `json:"body_path"`
	ModelFbx        string `json:"result_model_fbx"`
	ModelPre        string `json:"result_model_pre"`
	Process         string `json:"process"`
	Status          int64  `json:"status"`

	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}
