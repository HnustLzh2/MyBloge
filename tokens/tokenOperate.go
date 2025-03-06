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

func InitEnv() {
	SECRET_KEY = os.Getenv("SECRET_KEY")
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
func VerifyToken(userToken string) (MyClaims, error) {
	//从token中解析出claims
	token, err := jwt.ParseWithClaims(userToken, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return MyClaims{}, err
	}
	//拿出claims
	claims, ok := token.Claims.(*MyClaims)
	if !ok {
		return MyClaims{}, fmt.Errorf("invalid tokens")
	}
	if claims.ExpiresAt < time.Now().Unix() {
		return MyClaims{}, fmt.Errorf("tokens expired")
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
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization")
		if tokenString == "" {
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "token is empty"})
			c.Abort()
			return
		}
		claims, err := VerifyToken(tokenString)
		if err != nil {
			// 如果 Access Token 过期，尝试使用 Refresh Token 刷新
			var ve *jwt.ValidationError
			if errors.As(err, &ve) && ve.Errors == jwt.ValidationErrorExpired {
				refreshToken := c.Request.Header.Get("RefreshToken")
				if refreshToken == "" {
					c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "refresh token is empty"})
					c.Abort()
					return
				}
				// 刷新 Token
				newAccessToken, newRefreshToken, err := RefreshToken(refreshToken)
				if err != nil {
					c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
					c.Abort()
					return
				}
				// 返回新的 Token
				c.Header("Authorization", newAccessToken)
				c.Header("Refresh-Token", newRefreshToken)
				//给user设置token
				user, err := db.FindUserByEmail(claims.Email)
				if err != nil {
					c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
					c.Abort()
					return
				}
				UpdateToken(newAccessToken, newRefreshToken, &user)
				err = db.UpdateUser(user)
				if err != nil {
					c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "error updating user"})
					return
				}
			}
		}
		// 设置用户信息到上下文
		c.Set("email", claims.Email)
		c.Set("name", claims.Name)
		c.Next()
	}
}
