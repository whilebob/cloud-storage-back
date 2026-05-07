package common

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"time"
)

// GenerateSalt 生成随机盐值
func GenerateSalt(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	salt := make([]byte, length)
	for i := range salt {
		salt[i] = chars[rand.Intn(len(chars))]
	}
	return string(salt)
}

// SHA256 对字符串进行SHA256加密
func SHA256(str string) string {
	hash := sha256.New()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}

// SHA256WithSalt 对字符串进行带盐值的SHA256加密
func SHA256WithSalt(str string, salt string) string {
	return SHA256(str + salt)
}

// HashPassword 加密密码（自动生成盐值）
func HashPassword(password string) (string, string) {
	salt := GenerateSalt(16)
	return SHA256WithSalt(password, salt), salt
}

// VerifyPassword 验证密码是否正确
func VerifyPassword(password string, hash string, salt string) bool {
	return SHA256WithSalt(password, salt) == hash
}
