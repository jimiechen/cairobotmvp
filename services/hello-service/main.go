package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/hello", func(c *gin.Context) {
		name := c.DefaultQuery("name", "World")
		message := "Hello, " + name + "!"

		c.JSON(http.StatusOK, gin.H{
			"result": gin.H{
				"code": 0,
				"msg":  "success",
			},
			"message":   message,
			"timestamp": time.Now().Unix(),
		})
	})

	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
