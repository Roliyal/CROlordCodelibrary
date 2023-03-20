package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type GameResult struct {
	Username   string `json:"username"`
	Score      int    `json:"score"`
	CreateTime int64  `json:"create_time"`
}

var secretKey = []byte("secret_key")

func main() {
	r := gin.Default()

	// 登录认证接口
	r.POST("/login", func(c *gin.Context) {
		var user User
		if err := c.BindJSON(&user); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 检查用户名和密码是否正确，这里省略具体实现

		// 生成 JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": user.Username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(), // 过期时间为 24 小时
		})

		tokenString, err := token.SignedString(secretKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	})

	// 需要登录认证的接口
	auth := r.Group("/", JWTAuthMiddleware())
	{
		// 保存游戏成绩接口
		auth.POST("/game/result", func(c *gin.Context) {
			var result GameResult
			if err := c.BindJSON(&result); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// 保存游戏成绩到数据库，这里省略具体实现

			c.Status(http.StatusOK)
		})
	}

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

// JWT 认证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		tokenString := authHeader[len("Bearer "):]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}


		c.Next()
	}
}
