package main

import (
    "github.com/gin-gonic/gin"
    "github.com/dgrijalva/jwt-go"
    "net/http"
    "time"
)

type User struct {
    ID       uint64 `json:"id"`
    Email    string `json:"email"`
    Password string `json:"password"`
}
type GameResult struct {
    ID        uint64    `json:"id"`
    UserID    uint64    `json:"user_id"`
    Score     uint64    `json:"score"`
    CreatedAt time.Time `json:"created_at"`
}


func main() {
    // 初始化 Gin 框架
    r := gin.Default()

    // 用户注册
    r.POST("/register", func(c *gin.Context) {
        var user User
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // 在此处添加用户注册逻辑，例如将用户信息保存到数据库中
        // ...

        c.Status(http.StatusOK)
    })

    // 用户登录
    r.POST("/login", func(c *gin.Context) {
        var user User
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // 在此处添加用户登录逻辑，例如从数据库中查找用户信息并验证密码是否正确
        // ...

        // 生成 JWT
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
            "id": user.ID,
            "exp": time.Now().Add(time.Hour * 24).Unix(),
        })
        tokenString, err := token.SignedString([]byte("your-secret-key"))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"token": tokenString})
    })

    // 游戏成绩保存
    r.POST("/game_result", func(c *gin.Context) {
        var gameResult GameResult
        if err := c.ShouldBindJSON(&gameResult); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // 在此处添加保存游戏成绩逻辑，例如将游戏成绩信息保存到数据库中
        // ...

        c.Status(http.StatusOK)
    })

    // 启动服务器
    r.Run(":8080")
}
