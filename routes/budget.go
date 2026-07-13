package routes

import (
	"betterov2/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterBudgetRoutes(router *gin.Engine, c *controllers.BudgetController) {
	budgetGroup := router.Group("/budget_plan")
	budgetGroup.GET("/:intervalType", c.GetBudgetPlan)
	budgetGroup.POST("", c.PostBudgetPlan)
	budgetGroup.PUT("/:intervalType", c.PutBudgetPlan)
	budgetGroup.DELETE("/:intervalType", c.DeleteBudgetPlan)
}
