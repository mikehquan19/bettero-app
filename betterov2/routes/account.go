package routes

import (
	"betterov2/controllers"
	"betterov2/services"

	"github.com/gin-gonic/gin"
)

func RegisterAccountRoutes(router *gin.Engine, accService *services.AccountService) {
	accGroup := router.Group("/accounts")

	accGroup.GET("", controllers.GetAccounts(accService))
	accGroup.GET("/:id", controllers.GetAccount(accService))
	accGroup.POST("", controllers.PostAccounts(accService))
	accGroup.PUT("/:id", controllers.PutAccount(accService))
	accGroup.DELETE("/:id", controllers.DeleteAccount(accService))
	accGroup.GET("/:id/transactions", controllers.GetAccountTransactions(accService))
}
