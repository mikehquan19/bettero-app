package routes

import (
	"betterov2/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterAccountRoutes(router *gin.Engine, c *controllers.AccountController) {
	accGroup := router.Group("/accounts")

	accGroup.GET("", c.GetAccounts)
	accGroup.GET("/:id", c.GetAccount)
	accGroup.POST("", c.PostAccounts)
	accGroup.PUT("/:id", c.PutAccount)
	accGroup.DELETE("/:id", c.DeleteAccount)
	accGroup.GET("/:id/transactions", c.GetAccountTransactions)
	accGroup.GET("/:id/histories", c.GetAccountHistories)
	accGroup.GET("/:id/summary", c.GetAccountSummary)
}
