package routes

import (
	"betterov2/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.Engine, c *controllers.AuthController) {
	group := router.Group("/auth")

	group.POST("/register", c.Register)
	group.POST("/login", c.Login)
	group.POST("/logout", c.Logout)
}
