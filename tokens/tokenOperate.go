package tokens

import (
	"MyBloge/model"
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

var SECRET_KEY = os.Getenv("SECRET_KEY") //密钥

func GenerateToken(Email string, name string) (SignedToken string, refreshToken string, err error) {
	claims := MyClaims{
		Email: Email,
		Name:  name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}
	refreshClaims := MyClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 168).Unix(),
		},
	}
	refreshTokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}
	return tokenString, refreshTokenString, nil
}

// UpdateToken 更新user
func UpdateToken(signedToken string, refreshToken string, user *model.User) error {
	user.Token = signedToken
	user.RefreshToken = refreshToken
	return nil
}
func VerifyToken(userToken string) (MyClaims, error) {
	token, err := jwt.ParseWithClaims(userToken, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return MyClaims{}, err
	}
	claims, ok := token.Claims.(*MyClaims)
	if !ok {
		return MyClaims{}, fmt.Errorf("invalid tokens")
	}
	if claims.ExpiresAt < time.Now().Unix() {
		return MyClaims{}, fmt.Errorf("tokens expired")
	}
	return *claims, nil
}
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization")
		if tokenString == "" {
			c.IndentedJSON(http.StatusUnauthorized, "tokens is empty!")
			c.Abort()
			return
		}
		claims, err := VerifyToken(tokenString)
		if err != nil {
			c.IndentedJSON(http.StatusUnauthorized, "tokens is invalid!")
			c.Abort()
			return
		}
		c.Set("email", claims.Email)
		c.Set("name", claims.Name)
		c.Next()
	}
}
