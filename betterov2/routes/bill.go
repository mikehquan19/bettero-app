package routes

import (
	"betterov2/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterBillRoutes(router *gin.Engine, c *controllers.BillController) {
	billGroup := router.Group("/bills")

	billGroup.GET("", c.GetBills)
	billGroup.POST("", c.PostBill)
	billGroup.PUT("/:id", c.PutBill)
	billGroup.DELETE("/:id", c.DeleteBill)
}
