package model

import (
	"time"
)

type TUser struct {
	UserId     int64     `json:"user_id"`
	UserName   string    `json:"user_name"`
	Password   string    `json:"password"`
	Gender     string    `json:"gender"`
	IdCard     string    `json:"id_card"`
	Phone      string    `json:"phone"`
	Province   string    `json:"province"`
	City       string    `json:"city"`
	Area       string    `json:"area"`
	Address    string    `json:"address"`
	OpenId     string    `json:"open_id"`
	AvatarUrl  string    `json:"avatar_url"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

type TCode struct {
	Id         int64     `json:"id"`
	UserName   string    `json:"user_name"`
	Code       string    `json:"code"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

type TLoginInfo struct {
	UserId int64  `json:"userId"`
	Bearer string `json:"bearer"`
	Token  string `json:"token"`
}

type TUserScanInfo struct {
	UserId      int64 `json:"userId"`
	MaxCount    int64 `json:"max_count"`
	LeftCount   int64 `json:"left_count"`
	RecoverTime int64 `json:"recover_time"`
}

type TUserModel struct {
	UserId   int64  `json:"userId"`
	ScanId   string `json:"scan_id"`
	Status   int64  `json:"status"`
	ModelFbx string `json:"result_model_fbx"`
	ModePre  string `json:"result_model_pre"`
}
