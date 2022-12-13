package auth

import (
	"errors"
	"human/app/core/conf"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// jwt中payload的数据
type User struct {
	UserId int64
}

type MyClaims struct {
	User User
	jwt.StandardClaims
}

func GenerateToken(userInfo User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(conf.Conf.Auth.ExpireTime) * time.Minute) //有效期
	claims := &MyClaims{
		User: userInfo,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "human",
		},
	}
	// 生成Token，指定签名算法和claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 签名
	if tokenString, err := token.SignedString([]byte(conf.Conf.Auth.SignKey)); err != nil {
		return "", err
	} else {
		return tokenString, nil
	}

}

func RenewToken(claims *MyClaims) (string, error) {
	// 若token过期不超过最大超时时间则给它续签
	if withinLimit(claims.ExpiresAt, int64(conf.Conf.Auth.MaxTimeOut)) {
		return GenerateToken(claims.User)
	}
	return "", errors.New("登录已过期")
}

func ParseToken(tokenString string) (*MyClaims, error) {
	claims := &MyClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(conf.Conf.Auth.SignKey), nil
	})
	// 若token只是过期claims是有数据的，若token无法解析claims无数据
	return claims, err
}

func ParseToken2(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(conf.Conf.Auth.SignKey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("token无法解析")
}

// 计算过期时间是否超过l
func withinLimit(s int64, l int64) bool {
	e := time.Now().Unix()
	// println(e - s)
	return e-s < l
}
