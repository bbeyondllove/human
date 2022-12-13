package service

import (
	"context"
	"encoding/json"
	"errors"
	"human/app/core/conf"
	"human/app/core/dao"
	"human/app/core/model"
	"human/library/log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

func init() {
	os.MkdirAll("logs", 0755)
	os.MkdirAll("model", 0755)
}

var (
	_instService *Service
)

// Service struct
type Service struct {
	c   *conf.Config
	dao *dao.Dao
}

func CetConfig(s *Service) *conf.Config {
	return s.c
}

// New init
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: dao.New(c),
	}
	_instService = s
	return
}

// New init
func GetDao() *dao.Dao {
	return _instService.dao
}

// Ping ping the resource.
func (s *Service) Ping(c context.Context) (err error) {
	// TODO
	return
}

// Close close the resource.
func (s *Service) Close() {
	s.dao.Close()
}

func (s *Service) process(c context.Context, conf *conf.Config, scan *model.TScan, wg *sync.WaitGroup) error {
	url := conf.HttpClient["in3d"].GetAddr + scan.ScanId
	ret, err := s.dao.In3dClient.ReqData(http.MethodGet, url, nil, scan.Bearer)
	fields := make(map[string]string, 0)
	where := make(map[string]string, 0)

	defer func() {
		wg.Done()
	}()

	if err != nil {
		return err
	}

	retmap := make(map[string]interface{}, 0)
	err = json.Unmarshal([]byte(ret), &retmap)
	if err != nil {
		return err
	}

	if _, ok := retmap["status"]; !ok {
		err = errors.New("get dowdload url error")
		return err
	}

	status := retmap["status"].(string)
	if status != "completed" {
		fields["status"] = "3"
		where["id"] = strconv.Itoa(int(scan.Id))
		s.dao.UpdateScan(c, fields, where)
		return errors.New("not completed")
	}

	defer func() {
		if err != nil {
			fields["status"] = "5"
			where["id"] = strconv.Itoa(int(scan.Id))
			s.dao.UpdateScan(c, fields, where)
		}
	}()

	url1 := url + "/result?type=fbx"
	url2 := url + "/result?type=preview_png"
	down_fbx := ""
	down_pre := ""
	ret, err = s.dao.In3dClient.ReqData(http.MethodGet, url1, nil, scan.Bearer)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(ret), &retmap)
	if err != nil {
		return err
	}

	down_fbx = retmap["url"].(string)
	ret, err = s.dao.In3dClient.ReqData(http.MethodGet, url2, nil, scan.Bearer)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(ret), &retmap)
	if err != nil {
		return err
	}
	down_pre = retmap["url"].(string)
	log.Info("down_fbx=%v,down_pre=%v\n", down_fbx, down_pre)
	destdir := "model/" + strconv.FormatInt(scan.UserId, 10) + "/"

	filefbx, _ := s.dao.In3dClient.Download(down_fbx, int(scan.UserId), "")
	filepre, _ := s.dao.In3dClient.Download(down_pre, int(scan.UserId), "")
	if filefbx == "" || filepre == "" {
		return errors.New("filefbx is empty   or  filepre is  empty")
	}

	os.Rename(filepre, destdir+filepre)
	starturl := s.dao.Conf.HttpClient["aspose"].UploadStartAddr
	runurl := s.dao.Conf.HttpClient["aspose"].RunAddr
	geturl := s.dao.Conf.HttpClient["aspose"].GetAddr

	ret, _, err = s.dao.AsposeClient.TranFile(filefbx, starturl, runurl, geturl)
	if err != nil || ret == "" {
		if ret == "" {
			err = errors.New("TranFile is empty")
		}
		os.Rename(filefbx, destdir+filefbx)
		return err
	}

	log.Info("TranFile return %v\n", ret)
	os.Rename(filefbx, destdir+filefbx)
	filename := scan.ScanId + ".zip"
	tranfbx, _ := s.dao.AsposeClient.Download(ret, int(scan.UserId), filename)
	if tranfbx == "" {
		return errors.New("tranfbx is empty   or  filepre is  empty")
	}

	err = nil
	log.Info("Download return %v\n", tranfbx)
	os.Rename(tranfbx, destdir+tranfbx)
	fields["result_model_fbx"] = destdir + tranfbx
	fields["result_model_pre"] = destdir + filepre
	fields["status"] = "4"
	where["id"] = strconv.Itoa(int(scan.Id))
	s.dao.UpdateScan(c, fields, where)

	return nil
}

func (s *Service) Scan(conf *conf.Config) {

	for true {
		c := context.Background()
		list, _ := s.dao.GetScan(c)
		if len(list) == 0 {
			time.Sleep(time.Second * 10)
			continue
		}

		wg := new(sync.WaitGroup)
		wg.Add(len(list))
		for _, v := range list {
			c = context.Background()
			go s.process(c, conf, v, wg)
		}
		wg.Wait()
	}
}
