package utils

import (
	"cloud-storage/common"
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func HashPassword(password string) (string, error) {
	// 生成16字节的随机盐值
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	// 组合密码和盐值
	combined := password + string(salt)

	// 使用SHA-256哈希
	hash := sha256.Sum256([]byte(combined))

	// 将盐值和哈希值编码为base64
	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])

	// 返回格式：盐值$哈希值
	return fmt.Sprintf("%s$%s", saltBase64, hashBase64), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hashedPassword string) bool {
	// 分割盐值和哈希值
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 2 {
		return false
	}

	saltBase64 := parts[0]
	expectedHashBase64 := parts[1]

	// 解码盐值
	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return false
	}

	// 组合密码和盐值
	combined := password + string(salt)

	// 计算哈希值
	hash := sha256.Sum256([]byte(combined))
	actualHashBase64 := base64.StdEncoding.EncodeToString(hash[:])

	// 比较哈希值
	return actualHashBase64 == expectedHashBase64
}

var jwtSecret = []byte("your_secret_key")

const tokenExpiration = 12 * time.Hour

type CustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenToken 只传入用户名即可，同时存储到Redis
func GenToken(username string) (string, error) {
	claims := CustomClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	// 将token存储到Redis，key格式：user:login:token:username
	tokenKey := common.UserLoginTokenKey + ":" + username
	err = common.RS.Set(context.Background(), tokenKey, tokenStr, common.UserLoginTokenExpireTime)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

// ParseTokenToUsername 解析出用户名
func ParseTokenToUsername(tokenStr string) string {
	claims, err := ParseToken(tokenStr)
	if err != nil {
		return ""
	}
	return claims.Username
}

// ParseToken 解析并验证Redis中的token
func ParseToken(tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// 验证 token是否在Redis中存在
		tokenKey := common.UserLoginTokenKey + ":" + claims.Username
		storedToken, err := common.RS.Get(context.Background(), tokenKey)
		if err != nil || strings.Compare(storedToken, tokenStr) != 0 {
			return nil, fmt.Errorf("在Redis中不存在该 token或token不匹配")
		}
		return claims, nil
	}

	return nil, err
}

// RefreshToken 刷新token并更新Redis
func RefreshToken(tokenStr string) (string, error) {
	claims, err := ParseToken(tokenStr)
	if err != nil {
		return "", err
	}

	// 生成新 token
	newToken, err := GenToken(claims.Username)
	if err != nil {
		return "", err
	}

	tokenKey := common.UserLoginTokenKey + ":" + claims.Username
	err = common.RS.Set(context.Background(), tokenKey, newToken, tokenExpiration)
	if err != nil {
		return "", err
	}

	return newToken, nil
}

// InvalidateToken 使token失效（从Redis中删除）
func InvalidateToken(username string) error {
	tokenKey := common.UserLoginTokenKey + ":" + username
	_, err := common.RS.Remove(context.Background(), tokenKey, "")
	return err
}

// ShouldRefreshToken 检查token是否需要刷新
func ShouldRefreshToken(claims *CustomClaims) bool {
	// 当token剩余有效期小于6小时时，需要刷新
	return time.Until(claims.ExpiresAt.Time) < 30*time.Minute
}

// CalculateMd5 计算文件的MD5值
func CalculateMd5(fileHeader *multipart.FileHeader) string {
	f, err := fileHeader.Open()
	if err != nil {
		return ""
	}
	defer f.Close()
	hash := md5.New()
	io.Copy(hash, f)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
