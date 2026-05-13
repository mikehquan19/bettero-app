package main

import (
	"betterov2/controllers"
	"betterov2/routes"
	"betterov2/services"
	"betterov2/setup"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Connect to database
	setup.ConnectDB()

	router := gin.Default()

	var db *pgxpool.Pool = setup.ConnectDB()

	// Dependency injection
	accountService := services.NewAccountService(db)
	transactionService := services.NewTransactionService(db)
	billService := services.NewBillService(db)
	summaryService := services.NewSummaryService(db)

	accountController := controllers.NewAccountController(accountService)
	transactionController := controllers.NewTransactionController(transactionService)
	billController := controllers.NewBillController(billService)
	summaryController := controllers.NewSummaryController(summaryService)

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	routes.RegisterAccountRoutes(router, accountController)
	routes.RegisterTransactionRoutes(router, transactionController)
	routes.RegisterSummaryRoutes(router, summaryController)
	routes.RegisterBillRoutes(router, billController)

	router.Run(":8080")
}
