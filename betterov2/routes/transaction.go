package routes

import (
	"betterov2/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterTransactionRoutes(router *gin.Engine) {
	tranGroup := router.Group("/transactions")

	tranGroup.GET("", controllers.GetTransactions)
	tranGroup.POST("", controllers.PostTransaction)
	tranGroup.PUT("/:id", controllers.PutTransaction)
	tranGroup.DELETE("/:id", controllers.DeleteTransaction)
	tranGroup.GET("/autocomplete", controllers.SearchTransactions)
}
