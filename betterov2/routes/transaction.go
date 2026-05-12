package routes

import (
	"betterov2/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterTransactionRoutes(router *gin.Engine, c *controllers.TransactionController) {
	tranGroup := router.Group("/transactions")

	tranGroup.GET("", c.GetTransactions)
	tranGroup.POST("", c.PostTransaction)
	tranGroup.PUT("/:id", c.PutTransaction)
	tranGroup.DELETE("/:id", c.DeleteTransaction)
	tranGroup.GET("/autocomplete", c.SearchTransactions)
}
