package common

import (
	"CloudStorage/config"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type UserClaims struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var secret = config.AppConfig.JWT.Secret

// GenAccessToken 生成访问令牌，过期时间较短（如2小时）
func GenAccessToken(id int, username string) string {
	accessTokenExpires := time.Duration(config.AppConfig.JWT.AccessTokenExpires) * time.Hour
	claims := &UserClaims{
		ID:       id,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExpires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return tokenStr
}

// GenRefreshToken 生成刷新令牌，过期时间较长（如7天）
func GenRefreshToken(id int, username string) string {
	refreshTokenExpires := time.Duration(config.AppConfig.JWT.RefreshTokenExpires) * time.Hour
	claims := &UserClaims{
		ID:       id,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenExpires)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return tokenStr
}

// RefreshToken 使用RefreshToken刷新AccessToken，返回新的AccessToken和新的RefreshToken
func RefreshToken(refreshTokenStr string) (string, string, error) {
	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(refreshTokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return "", "", err
	}
	if !token.Valid {
		return "", "", errors.New("refresh token 已过期或无效")
	}
	newAccessToken := GenAccessToken(claims.ID, claims.Username)
	newRefreshToken := GenRefreshToken(claims.ID, claims.Username)
	return newAccessToken, newRefreshToken, nil
}

// VerifyRefreshToken 验证RefreshToken的有效性
func VerifyRefreshToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("refresh token 已过期")
	}
	claims := token.Claims.(*UserClaims)
	return claims, nil
}

func VerifyAccessToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token 已过期")
	}
	claims := token.Claims.(*UserClaims)
	return claims, nil
}
