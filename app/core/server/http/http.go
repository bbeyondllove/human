package http

import (
	"fmt"
	"human/app/core/conf"
	"human/app/core/middleware/access"
	"human/app/core/service"
	_ "human/docs"
	"human/library/log"
	"net/http"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	srv *service.Service
)

//token = "pUgaL2oHHLX0qn8mVLx9MItyQUJz0EGIQmnlvYNh7ZTFPOjrDgQQ2DCdUU3bQcJ3JPuupwmUGuNemEg_Gz81MQ"

// New init
func New(c *conf.Config, s *service.Service) (httpSrv *http.Server) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	//engine.Use(access.SlowAccess())
	route(engine)
	readTimeout := conf.Conf.App.ReadTimeout
	writeTimeout := conf.Conf.App.WriteTimeout
	endPoint := fmt.Sprintf(":%d", conf.Conf.App.HttpPort)
	maxHeaderBytes := 1 << 20
	httpSrv = &http.Server{
		Addr:           endPoint,
		Handler:        engine,
		ReadTimeout:    time.Duration(readTimeout),
		WriteTimeout:   time.Duration(writeTimeout),
		MaxHeaderBytes: maxHeaderBytes,
	}
	srv = s
	ginpprof.Wrapper(engine)
	go func() {
		// service connections
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("srv.ListenAndServe() error(%v) | config(%v)", err, c)
			panic(err)
		}
	}()
	return
}

func route(e *gin.Engine) {
	e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	downdir := conf.ConfPath + "/tran/"
	e.StaticFS("/tran", gin.Dir(downdir, true))

	api := e.Group("/api")

	api.Use(access.CorsMiddleware())
	{
		//验证码
		api.POST("/getcode", GetCode)

		//注册
		api.POST("/register", Register)

		//登录
		api.POST("/login", Login)

		//重置密码
		api.POST("/resetPassword", ResetPassword)

		api.Use(access.SlowAccess())

		//create new scan
		api.POST("/getid", GetId)

		//get scan
		api.GET("/getscan", GetScan)

		//get scan result
		api.GET("/getscanResult", GetScanResult)

		//userinfo
		api.GET("/GetUserScanInfo", GetUserScanInfo)

		//userinfo
		api.GET("/getuserModel", GetUserModel)

		//upload file
		api.PUT("/upload", UploadFile)

	}
}
