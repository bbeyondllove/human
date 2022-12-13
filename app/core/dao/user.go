package dao

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"human/app/core/model"
	"human/library/auth"
	"human/library/ecode"
	"human/library/log"
	"math/rand"
	"net/smtp"
	"strings"
	"time"

	"github.com/jordan-wright/email"
)

const (
	TUSER          = "t_user"
	TCODE          = "t_code"
	MAX_COUNT      = int64(2)
	RECOVER_SECOND = 3600 * 24
)

var (
	_insertUserSQL = fmt.Sprintf("INSERT INTO %s(user_name, password)VALUES(?, ?)", TUSER)
	_insertCodeSQL = fmt.Sprintf("INSERT INTO %s(user_name, code)VALUES(?, ?)", TCODE)
)

// data list
func (d *Dao) GetUsers(c context.Context, username string, password string) ([]*model.TUser, error) {
	sql := fmt.Sprintf("SELECT user_id,user_name, password from %s where user_name = '%s' ", TUSER, username)

	list := make([]*model.TUser, 0)

	if len(password) > 0 {
		pass := ecode.MD5(username + password)
		sql += fmt.Sprintf(" and password = '%s' ", pass)
	}

	rows, err := d.db.Query(c, sql)
	if err != nil {
		log.Error("d.db.Query(%s) error(%v)", sql, err)
		return list, err
	}
	for rows.Next() {
		tmp := new(model.TUser)
		err = rows.Scan(&tmp.UserId, &tmp.UserName, &tmp.Password)
		if err != nil {
			log.Error("rows.Scan error(%v)", err)
			continue
		}
		list = append(list, tmp)
	}
	return list, nil
}

func GenValidateCode(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < width; i++ {
		fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}

func SendMail(mail string) (string, error) {
	e := email.NewEmail()

	mailUserName := "yuxi@uvisiontech.cn" //邮箱账号
	mailPassword := "J9KZqjSx8zp9H5RY"    //邮箱授权码

	code := GenValidateCode(6) //发送的验证码
	Subject := "Uvision Tech"  //发送的主题

	e.From = "宇晰科技<yuxi@uvisiontech.cn>"
	e.To = append(e.To, mail)
	e.Subject = Subject
	//e.HTML = []byte("Hello,<br />Your verification code is :<br /><h1>" + code + "</h1><br />If you didn’t ask to verify this address, you can ignore this email.<br />Thanks,<br />Your UvisionTech-developer team")
	e.HTML = []byte("<br />Hello,<br /><br />Your verification code is :" + code + "<br /><br />If you didn’t ask to verify this address, you can ignore this email.<br /><br />Thanks,<br /><br />Your UvisionTech-developer team<br /><br />")
	err := e.SendWithTLS("hwsmtp.exmail.qq.com:465", smtp.PlainAuth("", mailUserName, mailPassword, "hwsmtp.exmail.qq.com"),
		&tls.Config{InsecureSkipVerify: true, ServerName: "hwsmtp.exmail.qq.com"})
	fmt.Printf("%v\n", err)

	return code, err
}

func (d *Dao) GetCode(c context.Context, username string) (interface{}, error) {
	code, err := SendMail(username)
	msg := "验证码发送成功"
	if err != nil {
		msg = "验证码发送失败"
	} else {
		d.AddCode(c, username, code)
	}

	return msg, err
}

func (d *Dao) ResetPassword(c context.Context, username, password, code string) (interface{}, error) {
	ret, err := d.QueryCode(c, username, code)
	codelist := ret.([]*model.TCode)
	if err != nil || len(codelist) == 0 {
		return int64(0), err
	}

	return d.UpdateUser(c, username, password)
}

// Update
func (d *Dao) UpdateUser(c context.Context, username, password string) (RowsAffected int64, err error) {

	sql := fmt.Sprintf("update %s set `password`=? where user_name=?", TUSER)
	pass := ecode.MD5(username + password)
	_, err = d.db.Exec(c, sql, pass, username)
	if err != nil {
		log.Error("UpdateUser error(%v) : %v", err, username)
		RowsAffected = int64(-1)
		return
	}
	RowsAffected, err = 1, err
	return
}

func (d *Dao) AddCode(c context.Context, username string, code string) (interface{}, error) {
	rsp := new(model.TCode)
	rsp.UserName = username

	_, err := d.db.Exec(c, _insertCodeSQL, username, code)
	if err != nil {
		log.Error("AddCode error(%v) : %s:%s", err, username, code)
		return nil, err
	}

	return rsp, nil
}

func (d *Dao) QueryCode(c context.Context, username string, code string) (interface{}, error) {
	sql := fmt.Sprintf("SELECT user_name,code   FROM %s where user_name='%s'  and code='%s' and id=(select max(id) from t_code where user_name='%s')",
		TCODE, username, code, username)
	list := make([]*model.TCode, 0)

	rows, err := d.db.Query(c, sql)
	if err != nil {
		log.Error("d.db.Query(%s) error(%v)", sql, err)
		return list, err
	}
	for rows.Next() {
		tmp := new(model.TCode)
		err = rows.Scan(&tmp.Code, &tmp.Code)
		if err != nil {
			log.Error("rows.Scan error(%v)", err)
			continue
		}
		list = append(list, tmp)
	}
	return list, nil
}

// Register
func (d *Dao) Register(c context.Context, username string, password string, code string) (interface{}, error) {
	list, err := d.GetUsers(c, username, "")
	if len(list) > 0 || err != nil {
		return list, errors.New(ecode.UserExist.Error())
	}

	bear, _ := d.GetBearer(c)
	if len(bear) == 0 {
		return nil, errors.New("no enough bearer!")
	}

	ret, serr := d.QueryCode(c, username, code)
	codelist := ret.([]*model.TCode)
	if serr != nil || len(codelist) == 0 {
		ulist := make([]*model.TUser, 0)
		u := &model.TUser{
			UserId: -1,
		}
		ulist = append(ulist, u)
		return ulist, errors.New("code erorr")
	}

	rsp := new(model.TLoginInfo)
	rsp.Bearer = bear[0].Bearer
	pass := ecode.MD5(username + password)
	res, suberr := d.db.Exec(c, _insertUserSQL, username, pass)
	if suberr != nil {
		log.Error("Register error(%v) : %s:%s", err, username, pass)
		return nil, suberr
	}

	rsp.UserId, err = res.LastInsertId()
	return rsp, nil
}

// Login
func (d *Dao) Login(c context.Context, username string, password string) (interface{}, error) {
	list, err := d.GetUsers(c, username, password)
	if err != nil || len(list) == 0 {
		log.Error("Login error(%v) list: %v", err, list)
		return nil, err
	}

	bearer, suberr := d.GetUserBearer(c, list[0].UserId)
	if suberr != nil || len(bearer) == 0 {
		log.Error("GetUserBearer error(%v) userid: %v", err, list[0].UserId)
		return nil, suberr
	}

	var userbear model.TLoginInfo
	var user auth.User
	user.UserId = list[0].UserId
	userbear.Token, _ = auth.GenerateToken(user)
	userbear.Bearer = bearer[0].Bearer
	userbear.UserId = list[0].UserId

	return userbear, suberr

}

func (d *Dao) GetUserScanInfo(c context.Context, bearer string, userId int64) (interface{}, error) {
	sche := &model.TScanScheduler{
		Bearer: bearer,
		UserId: userId,
	}

	bearlist, _ := d.GetScheduler(c, sche)
	LeftCount := MAX_COUNT - int64(len(bearlist))
	RecoverTime := int64(0)

	if LeftCount < 0 {
		LeftCount = 0
	}
	if LeftCount == 0 {
		maxtime := int64(0)
		for _, v := range bearlist {
			if maxtime < v.UpdateTime.Unix() {
				maxtime = v.UpdateTime.Unix()
			}
		}

		lost_time := time.Now().Unix() - maxtime
		if lost_time >= RECOVER_SECOND {
			RecoverTime = 0
		} else {
			RecoverTime = RECOVER_SECOND - lost_time
		}
	}

	ret := make([]int64, 0)
	ret = append(ret, MAX_COUNT)
	ret = append(ret, LeftCount)
	ret = append(ret, RecoverTime)
	return ret, nil
}

func (d *Dao) GetUserModel(c context.Context, userId int64) (interface{}, error) {
	ret, err := d.GetUserScan(c, userId)
	retary := make([]*model.TUserModel, 0)
	for _, v := range ret {
		s := &model.TUserModel{
			UserId:   userId,
			ScanId:   v.ScanId,
			ModelFbx: v.ModelFbx,
			ModePre:  v.ModelPre,
			Status:   v.Status,
		}
		retary = append(retary, s)
	}

	return retary, err
}
