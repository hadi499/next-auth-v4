package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"server-token/controllers"
	"server-token/database"
	"server-token/middleware"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Ganti * dengan domain tertentu jika perlu
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Jika method OPTIONS, langsung response 200 OK
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func main() {
	database.ConnectDatabase()

	r := gin.Default()
	r.Use(CORSMiddleware())
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
	r.GET("/user", controllers.GetUserByEmail)

	authRoutes := r.Group("/")
	authRoutes.Use(middleware.AuthMiddleware())
	{
		authRoutes.GET("/", controllers.Home)
		authRoutes.POST("/logout", controllers.Logout)
	}
	r.Run(":8080")

}

