package utils

import (
	"golang.org/x/crypto/bcrypt"
	"regexp"
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

// CheckEmailValid CheckEmail ^[a-zA-Z0-9._%+\-]+：匹配邮箱用户名部分，允许字母、数字、点、下划线、百分号、加号和减号。
// @：邮箱地址中必须包含一个 @ 符号。
// [a-zA-Z0-9.\-]+：匹配域名部分，允许字母、数字、点和减号。
// \.[a-zA-Z]{2,}$：匹配顶级域名，必须以点开头，后跟至少两个字母。
func CheckEmailValid(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}
func CheckPasswordValid(password string) bool {
	lengthRegex := `^{6,}`
	UpperRegex := `[A-Z]`
	LowerRegex := `[a-z]` //默认匹配一次
	NumberRegex := `[0-9]`
	hasSymbolRegex := `[\W_]+` //匹配一次或多次，优先多次
	// 编译正则表达式
	lengthMatch := regexp.MustCompile(lengthRegex)
	upperMatch := regexp.MustCompile(UpperRegex)
	lowerMatch := regexp.MustCompile(LowerRegex)
	digitMatch := regexp.MustCompile(NumberRegex)
	symbolMatch := regexp.MustCompile(hasSymbolRegex)
	if !lengthMatch.MatchString(password) {
		return false
	}
	count := 0
	if upperMatch.MatchString(password) {
		count++
	}
	if lowerMatch.MatchString(password) {
		count++
	}
	if digitMatch.MatchString(password) {
		count++
	}
	if symbolMatch.MatchString(password) {
		count++
	}
	if count >= 3 {
		return true
	} else {
		return false
	}
}
