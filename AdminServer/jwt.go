package AdminServer

import (
	"bflog/config"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

//var jwtKey = []byte("111") // 请替换为实际的密钥

// Claims 定义了 JWT 的声明结构
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成一个 JWT token
func GenerateJWT(username string) (string, error) {
	//logrus.Info([]byte(config.GetBase().Server.Seckey))
	// 设置 JWT 过期时间为 24 小时
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// 使用 HS256 算法生成 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.GetBase().Server.Seckey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ValidateJWT 验证 JWT token
func ValidateJWT(tokenStr string) (*Claims, error) {

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetBase().Server.Seckey), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, err
		}
		return nil, err
	}
	if !token.Valid {
		return nil, err
	}
	return claims, nil
}
