package routes

import (
	"betterov2/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterBillRoutes(router *gin.Engine) {
	billGroup := router.Group("/bills")

	billGroup.GET("", controllers.GetBills)
	billGroup.POST("", controllers.PostBill)
	billGroup.PUT("/:id", controllers.PutBill)
	billGroup.DELETE("/:id", controllers.DeleteBill)
}
