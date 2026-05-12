package routes

import (
	"betterov2/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterSummaryRoutes(router *gin.Engine, s *controllers.SummaryController) {
	summaryGroup := router.Group("/summary")

	summaryGroup.GET("", s.GetSummary)
}
