package access

import (
	"human/app/core/service"
	"human/library/auth"
	"human/library/ecode"
	"human/library/log"
	"human/library/render"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	_family          = "access"
	_slowLogDuration = time.Second
)

// SlowAccess handler record slow access
func SlowAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		JWTAuth(c)
		// Process request
		c.Next()

		// Stop timer
		latency := time.Since(start)
		if latency > _slowLogDuration {
			clientIP := c.ClientIP()
			method := c.Request.Method
			statusCode := c.Writer.Status()
			log.Warn("%s|%v|%d|%v|%s|%s|%s|%s|",
				_family,
				start.Format("2006-01-02 15:04:05"),
				statusCode,
				latency,
				clientIP,
				method,
				path,
				raw)
		}
	}
}

// CorsMiddleware 解决跨域的方法
func CorsMiddleware() gin.HandlerFunc {
	corsConf := cors.Config{
		MaxAge:                 12 * time.Hour,
		AllowBrowserExtensions: true,
	}

	// 在開發環境時，允許所有 origins、所有 methods 和多數的 headers
	corsConf.AllowAllOrigins = true
	corsConf.AllowMethods = []string{"GET", "POST", "DELETE", "OPTIONS", "PUT"}
	corsConf.AllowHeaders = []string{"Authorization", "Content-Type", "Upgrade", "Origin",
		"Connection", "Accept-Encoding", "Accept-Language", "Host"}

	return cors.New(corsConf)

}

func JWTAuth(c *gin.Context) {
	r := render.New(c)
	token := c.Request.Header.Get("jwt")
	if len(token) == 0 {
		// 无token直接拒绝
		c.Abort()
		r.JSON(nil, ecode.AccessDenied)
		//c.String(ecode.RequestErr, "无权限")
		return
	}
	// 校验token
	claims, err := auth.ParseToken(token)
	list, serr := service.GetDao().GetUserBearer(c, claims.User.UserId)
	if serr != nil || len(list) == 0 {
		c.Abort()
		r.JSON(nil, ecode.AccessDenied)
	}

	if err != nil {
		if strings.Contains(err.Error(), "expired") {
			// 若过期调用续签函数
			newToken, _ := auth.RenewToken(claims)
			if newToken != "" {
				// 续签成功給返回头设置一个newtoken字段
				c.Header("newtoken", newToken)
				c.Request.Header.Set("jwt", newToken)
				c.Request.Header.Set("userid", strconv.FormatInt(claims.User.UserId, 10))
				c.Request.Header.Set("bearer", list[0].Bearer)
				c.Next()
				return
			}
		}
		// Token验证失败或续签失败直接拒绝请求
		c.Abort()
		r.JSON(nil, ecode.AccessDenied)
		return
	}
	// token未过期继续执行1其他中间件

	c.Request.Header.Set("userid", strconv.FormatInt(claims.User.UserId, 10))
	c.Request.Header.Set("bearer", list[0].Bearer)
	c.Next()

}
