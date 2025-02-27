package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// EncodePassword 加密
func EncodePassword(userPassword string) (string, error) {
	enPassword, err := bcrypt.GenerateFromPassword([]byte(userPassword), 14)
	return string(enPassword), err
}

// CheckOutPassword 解密并验证
func CheckOutPassword(enPassword string, userPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(enPassword), []byte(userPassword))
	return err == nil
}
