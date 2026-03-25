package routes

import (
	"betterov2/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterAccountRoutes(router *gin.Engine) {
	accGroup := router.Group("/accounts")

	accGroup.GET("", controllers.GetAccounts)
	accGroup.GET("/:id", controllers.GetAccount)
	accGroup.POST("", controllers.PostAccounts)
	accGroup.PUT("/:id", controllers.PutAccount)
	accGroup.DELETE("/:id", controllers.DeleteAccount)
	accGroup.GET("/:id/transactions", controllers.GetAccountTransactions)
}
