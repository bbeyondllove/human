package http

import (
	"encoding/json"
	"human/app/core/model"
	"human/library/ecode"
	"human/library/log"
	"human/library/render"
	"io"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
)

func GetRedisKey(c *gin.Context) {
	r := render.New(c)
	key := c.Query("key")
	if key == "" {
		r.JSON(nil, ecode.RequestErr)
		return
	}
	val, err := srv.GetRedisKey(c, key)
	r.JSON(val, err)
}

func SetRedisKey(c *gin.Context) {
	r := render.New(c)
	key := c.Query("key")
	if len(key) == 0 {
		r.JSON(nil, ecode.RequestErr)
		return
	}
	val := c.Query("val")
	if len(val) == 0 {
		r.JSON(nil, ecode.RequestErr)
		return
	}
	expire := com.StrTo(c.Query("expire")).MustInt64()
	if expire <= 0 || expire > 300 {
		r.JSON(nil, ecode.RequestErr)
		return
	}
	err := srv.SetRedisKey(c, key, val, expire)
	r.JSON(nil, err)
}

// @Summary 获取注册码
// @Produce json
// @Param username query string true "用户名（邮箱）"
// @Success 200 {object} render.JSON
// @Router /api/getcode [post]
func GetCode(c *gin.Context) {
	r := render.New(c)
	v := new(struct {
		UserName string `form:"username" binding:"required,min=1,max=30"`
	})

	err := c.Bind(v)
	if err != nil {
		r.JSON("{}", ecode.RequestErr)
		return
	}

	ret, err := srv.GetCode(c, v.UserName)
	r.JSON(ret, err)

}

// @Summary 修改密码
// @Produce json
// @Param username query string true "用户名（邮箱）"
// @Param password query string true "新密码"
// @Param code query string true "验证码"
// @Success 200 {object} render.JSON
// @Router /api/resetPassword [post]
func ResetPassword(c *gin.Context) {
	r := render.New(c)
	v := new(struct {
		UserName string `form:"username" binding:"required,min=1,max=30"`
		Password string `form:"password" binding:"required,min=1,max=30"`
		Code     string `form:"code" binding:"required,min=1,max=30"`
	})

	err := c.Bind(v)
	if err != nil {
		r.JSON("{}", ecode.RequestErr)
		return
	}

	ret, err := srv.ResetPassword(c, v.UserName, v.Password, v.Code)
	if ret.(int64) == 1 {
		r.JSON("密码重置成功", ecode.OK)
	} else if ret.(int64) == 0 {
		r.JSON("密码重置失败", ecode.CodeError)
	} else {
		r.JSON("密码重置失败", ecode.ServerErr)
	}
}

// @Summary 注册
// @Produce json
// @Param username query string true "用户名"
// @Param password query string true "密码"
// @Param code query string true "注册码"
// @Success 200 {object} render.JSON
// @Router /api/register [post]
func Register(c *gin.Context) {
	r := render.New(c)

	v := new(struct {
		UserName string `form:"username" binding:"required,min=1,max=30"`
		Password string `form:"password" binding:"required,min=1,max=30"`
		Code     string `form:"code" binding:"required,min=1,max=30"`
	})

	err := c.Bind(v)
	if err != nil {
		r.JSON("{}", ecode.RequestErr)
		return
	}

	ret, err := srv.TxAddUserUpdateBear(c, v.UserName, v.Password, v.Code)
	if ret != nil && err != nil {
		if ret.([]*model.TUser)[0].UserId == -1 {
			r.JSON("{}", ecode.CodeError)
		} else {
			r.JSON("{}", ecode.UserExist)
		}
		return
	}

	if err != nil {
		r.JSON("{}", ecode.UserLimited)
		return
	}

	r.JSON(ret, err)
}

// @Summary 登录
// @Produce json
// @Param username query string true "用户名"
// @Param password query string true "密码"
// @Success 200 {object} render.JSON
// @Router /api/login [post]
func Login(c *gin.Context) {
	r := render.New(c)

	v := new(struct {
		UserName string `form:"username" binding:"required,min=1,max=30"`
		Password string `form:"password" binding:"required,min=1,max=30"`
	})

	err := c.Bind(v)
	if err != nil {
		r.JSON("{}", ecode.RequestErr)
		return
	}

	ret, err := srv.Login(c, v.UserName, v.Password)
	if ret == nil {
		r.JSON("{}", ecode.AccessDenied)
		return
	}
	r.JSON(ret, err)
}

// @Summary 获取in3d访问id
// @Produce json
// @Param jwt  header  string true "登录token"
// @Success 200 {object} render.JSON
// @Router /api/getid [post]
func GetId(c *gin.Context) {
	r := render.New(c)
	Authorization := c.Request.Header.Get("bearer")
	userid := c.Request.Header.Get("userid")
	ret, err := srv.GetId(c, Authorization, userid)
	b := json.RawMessage([]byte(ret.(string)))
	r.JSON(b, err)
}

// @Summary 获取in3d扫描
// @Produce json
// @Param jwt  header  string true "登录token"
// @Param id query string true "in3d访问id"
// @Success 200 {object} render.JSON
// @Router /api/getscan [get]
func GetScan(c *gin.Context) {
	r := render.New(c)
	Authorization := c.Request.Header.Get("bearer")
	v := new(struct {
		Id string `form:"id" binding:"required,min=1,max=300"`
	})
	err := c.Bind(v)
	if err != nil {
		r.JSON(nil, ecode.RequestErr)
		return
	}
	ret, err := srv.GetScan(c, Authorization, v.Id)
	r.JSON(ret, err)
}

// @Summary 获取in3d扫描结果
// @Produce json
// @Param jwt  header  string true "登录token"
// @Param id query string true "in3d访问id"
// @Success 200 {object} render.JSON
// @Router /api/getscanResult [get]
func GetScanResult(c *gin.Context) {
	r := render.New(c)
	Authorization := c.Request.Header.Get("bearer")
	v := new(struct {
		Id string `form:"id" binding:"required,min=1,max=300"`
	})

	err := c.Bind(v)
	if err != nil {
		r.JSON(nil, ecode.RequestErr)
		return
	}

	ret, err := srv.GetScanResult(c, Authorization, v.Id)
	r.JSON(ret, err)

}

// @Summary 上传文件
// @Produce json
// @Param jwt  header  string true "登录token"
// @Param id query string true "in3d访问id"
// @Param file formData file true "file"
// @Success 200 {object} render.JSON
// @Router /api/upload [put]
func UploadFile(c *gin.Context) {
	r := render.New(c)
	Authorization := c.Request.Header.Get("bearer")
	userid := c.Request.Header.Get("userid")
	user_id, _ := strconv.ParseInt(userid, 10, 64)

	id := c.Query("id")
	if id == "" {
		r.JSON(nil, ecode.RequestErr)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		r.JSON(nil, ecode.RequestErr)
		return
	}

	//获取文件名
	filename := header.Filename
	destdir := "model/" + userid + "/"
	os.MkdirAll(destdir, 777)
	//写入文件
	out, err := os.Create(destdir + filename)
	if err != nil {
		r.JSON(nil, ecode.ServerErr)
		return
	}

	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		r.JSON(nil, ecode.ServerErr)
		return
	}

	//ret, reterr := "上传成功", err
	//log.Info("Upload  %s %d %d %s", Authorization, user_id, id, destdir+filename)
	ret, reterr := srv.Upload(c, Authorization, user_id, id, destdir+filename)
	log.Info("Upload  ret [%v]%v", reterr, ret)
	if ret == nil {
		r.JSON(nil, reterr)
		return
	}

	buf, _ := json.Marshal(ret)
	b := json.RawMessage(buf)
	r.JSON(b, reterr)

}

// @Summary 获取用户信息
// @Produce json
// @Param jwt  header  string true "登录token"
// @Success 200 {object} render.JSON
// @Router /api/GetUserScanInfo [get]
func GetUserScanInfo(c *gin.Context) {
	r := render.New(c)
	Authorization := c.Request.Header.Get("bearer")
	userid := c.Request.Header.Get("userid")
	user_id, _ := strconv.ParseInt(userid, 10, 64)
	ret, err := srv.GetUserScanInfo(c, Authorization, user_id)
	r.JSON(ret, err)
}

// @Summary 获取用户模型列表
// @Produce json
// @Param jwt  header  string true "登录token"
// @Success 200 {object} render.JSON
// @Router /api/getuserModel [get]
func GetUserModel(c *gin.Context) {
	r := render.New(c)
	userid := c.Request.Header.Get("userid")
	user_id, _ := strconv.ParseInt(userid, 10, 64)
	ret, err := srv.GetUserModel(c, user_id)
	r.JSON(ret, err)
}
