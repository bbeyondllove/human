package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"human/app/core/model"
	"human/library/log"
	"net/http"
	"strconv"
)

// GetRedisKey
func (s *Service) GetRedisKey(c context.Context, key string) (val string, err error) {
	return s.dao.GetKey(c, key)
}

// SetRedisKey
func (s *Service) SetRedisKey(c context.Context, key, val string, expire int64) (err error) {
	return s.dao.SetKey(c, key, val, expire)
}

func (s *Service) GetCode(c context.Context, username string) (ret interface{}, err error) {
	return s.dao.GetCode(c, username)
}

func (s *Service) ResetPassword(c context.Context, username, password, code string) (ret interface{}, err error) {
	return s.dao.ResetPassword(c, username, password, code)
}

// TxAddUserUpdateBear
func (s *Service) TxAddUserUpdateBear(c context.Context, username string, password string, code string) (ret interface{}, err error) {
	tx, err := s.dao.BeginTran(c)
	if err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return nil, err
	}

	ret, suberr := s.dao.Register(c, username, password, code)
	if suberr != nil || ret == nil {
		log.Error("s.dao.Register(%v) error(%s %s)", username, password)
		tx.Rollback()
		return ret, suberr
	}

	tUserBear := ret.(*model.TLoginInfo)
	bearer := &model.TBearer{
		Bearer: tUserBear.Bearer,
		UserId: tUserBear.UserId,
	}

	rid, rerr := s.dao.UpdateBearer(c, bearer)
	if rerr != nil || rid == 0 {
		log.Error("s.dao.UpdateBearer(%v) error(%v)", bearer, err)
		tx.Rollback()
		return nil, errors.New("s.dao.UpdateBearer error")
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return ret, err
}

// Login
func (s *Service) Login(c context.Context, username string, password string) (ret interface{}, err error) {
	return s.dao.Login(c, username, password)

}

// GetId
func (s *Service) GetId(ctx context.Context, bearer string, userid string) (interface{}, error) {
	user_id, _ := strconv.ParseInt(userid, 10, 64)

	url := fmt.Sprintf("%snew?config=head_body", s.dao.Conf.HttpClient["in3d"].GetAddr)
	res, err := s.dao.In3dClient.ReqData(http.MethodPost, url, nil, bearer)
	fmt.Printf("response Body:%+v\n", res)
	if err != nil {
		log.Error("In3dClient.Get(%s) error(%v)", url, err)
		return res, err
	}

	rsp := make(map[string]string, 0)
	//res = "{\"id\":\"41a9eda6-542c-465b-af7f-3b6594e68f25\"}"
	err = json.Unmarshal([]byte(res), &rsp)
	if err == nil {
		scheduler := &model.TScanScheduler{
			Bearer: bearer,
			UserId: user_id,
			ScanId: rsp["id"],
		}
		s.dao.UpdateScheduler(ctx, scheduler)
	}

	return res, nil
}

// GetScan
func (s *Service) GetScan(ctx context.Context, bearer string, id string) (interface{}, error) {
	url := fmt.Sprintf("%s%s", s.dao.Conf.HttpClient["in3d"].GetAddr, id)
	res, err := s.dao.In3dClient.ReqData(http.MethodGet, url, nil, bearer)

	fmt.Printf("response Body:%+v\n", res)
	if err != nil {
		log.Error("In3dClient.Get(%s) error(%v)", url, err)
	}

	return res, err
}

// GetScanResult
func (s *Service) GetScanResult(ctx context.Context, bearer string, id string) (interface{}, error) {
	url := fmt.Sprintf("%s%s", s.dao.Conf.HttpClient["in3d"].GetAddr, id)
	res, err := s.dao.In3dClient.ReqData(http.MethodGet, url, nil, bearer)
	fmt.Printf("response Body:%+v\n", res)
	if err != nil {
		log.Error("In3dClient.Get(%s) error(%v)", url, err)
		return res, err
	}

	return res, nil
}

// upload
func (s *Service) Upload(ctx context.Context, bearer string, user_id int64, id string, filename string) (interface{}, error) {
	starturl := fmt.Sprintf("%s%s", s.dao.Conf.HttpClient["in3d"].UploadStartAddr, id)
	doneurl := fmt.Sprintf("%s%s?&etag=", s.dao.Conf.HttpClient["in3d"].UploadDoneAddr, id)
	runurl := fmt.Sprintf("%s%s", s.dao.Conf.HttpClient["in3d"].RunAddr, id)
	scan := &model.TScan{
		UserId: user_id,
		ScanId: id,
		Status: int64(0),
	}
	lstid, serr := s.dao.AddScan(ctx, scan)
	if serr != nil {
		return "", serr
	}

	fields := make(map[string]string, 0)
	where := make(map[string]string, 0)
	where["id"] = strconv.Itoa(int(lstid))
	res, err := s.dao.In3dClient.Upload(http.MethodPost, bearer, id, filename, starturl, doneurl, runurl)
	fmt.Printf("response Body:%+v\n", res)
	if err != nil || res == "" {
		log.Error("In3dClient.Upload(%s %s %s) error(%v)", starturl, doneurl, runurl, err)
		fields["status"] = "2"
		s.dao.UpdateScan(ctx, fields, where)
		return res, err
	}

	fields["status"] = "1"
	s.dao.UpdateScan(ctx, fields, where)
	return res, nil
}

func (s *Service) GetUserScanInfo(ctx context.Context, bearer string, userid int64) (interface{}, error) {
	var userbear model.TUserScanInfo
	userbear.UserId = userid
	ret, _ := s.dao.GetUserScanInfo(ctx, bearer, userid)
	retary := ret.([]int64)
	userbear.MaxCount, userbear.LeftCount, userbear.RecoverTime = retary[0], retary[1], retary[2]

	return userbear, nil
}

func (s *Service) GetUserModel(ctx context.Context, userid int64) (interface{}, error) {
	return s.dao.GetUserModel(ctx, userid)

}
