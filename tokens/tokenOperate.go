package tokens

import (
	"MyBloge/db"
	"MyBloge/model"
	"errors"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"time"
)

type MyClaims struct {
	Email string
	Name  string
	jwt.StandardClaims
}

var SECRET_KEY string // 密钥
var TokenExpireError error

// init 函数会在包被导入时自动执行	***
func init() {
	SECRET_KEY = os.Getenv("SECRET_KEY")
	if SECRET_KEY == "" {
		// 如果环境变量未设置，可以提供一个默认值或报错
		SECRET_KEY = "default-secret-key"
	}
	TokenExpireError = errors.New("TokenExpireError")
}

func GenerateToken(email string, name string) (accessToken string, refreshToken string, err error) {
	// Access Token 的过期时间较短
	accessTokenClaims := MyClaims{
		Email: email,
		Name:  name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(), // 1 小时过期
			Issuer:    "your-app",
		},
	}
	// Refresh Token 的过期时间较长
	refreshTokenClaims := MyClaims{
		Email: email,
		Name:  name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 168).Unix(), // 7 天过期
			Issuer:    "your-app",
		},
	}
	// 生成 Access Token
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}
	// 生成 Refresh Token
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// UpdateToken 更新user
func UpdateToken(signedToken string, refreshToken string, user *model.User) {
	user.Token = signedToken
	user.RefreshToken = refreshToken
}

// VerifyToken 验证Token并解析Claims
func VerifyToken(userToken string) (MyClaims, error) {
	token, err := jwt.ParseWithClaims(userToken, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		// 如果是Token过期错误
		var ve *jwt.ValidationError
		if errors.As(err, &ve) && ve.Errors&jwt.ValidationErrorExpired != 0 {
			// 拿出claims（即使Token过期，claims仍然可以被解析出来）
			claims, ok := token.Claims.(*MyClaims)
			if !ok {
				return MyClaims{}, fmt.Errorf("invalid token claims")
			}
			// 返回带有用户信息的claims，并标记为Token过期错误
			return MyClaims{Email: claims.Email, Name: claims.Name}, TokenExpireError
		}
		// 如果是其他错误，直接返回空的claims和错误
		return MyClaims{}, err
	}
	claims, ok := token.Claims.(*MyClaims)
	if !ok {
		return MyClaims{}, fmt.Errorf("invalid token claims")
	}
	return *claims, nil
}
func RefreshToken(refreshToken string) (newAccessToken string, newRefreshToken string, err error) {
	// 解析 Refresh Token
	claims := MyClaims{}
	_, err = jwt.ParseWithClaims(refreshToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return "", "", err
	}
	// 验证 Refresh Token 是否过期
	if claims.ExpiresAt < time.Now().Unix() {
		return "", "", fmt.Errorf("refresh token expired")
	}
	// 生成新的 Access Token 和 Refresh Token
	newAccessToken, newRefreshToken, err = GenerateToken(claims.Email, claims.Name)
	if err != nil {
		return "", "", err
	}
	return newAccessToken, newRefreshToken, nil
}

// NullifyTokenCookiesAndHeader 移除所有的cookies喝Header
func NullifyTokenCookiesAndHeader(c *gin.Context) {
	c.Writer.Header().Del("Authorization")
	c.Writer.Header().Del("RefreshToken")
	c.SetCookie("Authorization", "", -1, "/", "", false, true)
	c.SetCookie("RefreshToken", "", -1, "/", "", false, true)
}
func SetInfoAtHeaderAndCookies(c *gin.Context, authToken string, refreshToken string, email string, name string) {
	c.Header("Authorization", authToken)
	c.Header("Refresh-Token", refreshToken)
	c.SetCookie("authToken", authToken, 60*60*24, "/", "", false, true)
	c.SetCookie("refreshToken", refreshToken, 60*60*24, "/", "", false, true)
	c.Set("email", email)
	c.Set("name", name)
}
func Authorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization")
		refreshToken := c.Request.Header.Get("RefreshToken")
		if tokenString == "" {
			NullifyTokenCookiesAndHeader(c)
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "token is empty"})
			c.Abort()
			return
		}
		if refreshToken == "" {
			NullifyTokenCookiesAndHeader(c)
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "refresh token is empty"})
			c.Abort()
			return
		}
		claims, err := VerifyToken(tokenString)
		if err != nil {
			// 如果 Access Token 过期，尝试使用 Refresh Token 刷新
			if errors.Is(err, TokenExpireError) {
				//检查RefreshToken是不是合法或者是不是过期
				_, err := VerifyToken(refreshToken)
				if err != nil {
					NullifyTokenCookiesAndHeader(c)
					c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "refresh token is invalid"})
					c.Abort()
				}
				// 刷新 Token
				newAccessToken, newRefreshToken, err := RefreshToken(refreshToken)
				if err != nil {
					NullifyTokenCookiesAndHeader(c)
					c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
					c.Abort()
					return
				}
				//给user设置token
				user, err := db.FindUserByEmail(claims.Email)
				if err != nil {
					NullifyTokenCookiesAndHeader(c)
					c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "database Error"})
					c.Abort()
					return
				}
				UpdateToken(newAccessToken, newRefreshToken, &user)
				err = db.UpdateUser(user)
				if err != nil {
					NullifyTokenCookiesAndHeader(c)
					c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "error updating user"})
					return
				}
				SetInfoAtHeaderAndCookies(c, newAccessToken, newRefreshToken, claims.Email, claims.Name)
				c.Next()
			} else {
				NullifyTokenCookiesAndHeader(c)
				c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "token is invalid"})
				c.Abort()
				return
			}
		}
		// 设置用户信息到上下文
		SetInfoAtHeaderAndCookies(c, tokenString, refreshToken, claims.Email, claims.Name)
		c.Next()
	}
}
