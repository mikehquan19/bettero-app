package routes

import (
	"betterov2/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterSummaryRoutes(router *gin.Engine) {
	summaryGroup := router.Group("/summary")

	summaryGroup.GET("", controllers.GetSummary)
}
