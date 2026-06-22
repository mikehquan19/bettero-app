package main

import (
	"betterov2/controllers"
	"betterov2/repositories"
	"betterov2/routes"
	"betterov2/services"
	"betterov2/setup"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Connect to database
	var db = setup.ConnectDB()

	// Initialize repositories
	accountRepo := repositories.NewAccountRepo()
	accountHistoryRepo := repositories.NewAccountHistoryRepo()
	transactionRepo := repositories.NewTransactionRepo()
	billRepo := repositories.NewBillRepo()

	// Dependency injection
	accountService := services.NewAccountService(db, accountRepo, accountHistoryRepo, transactionRepo)
	transactionService := services.NewTransactionService(db, transactionRepo, accountRepo)
	billService := services.NewBillService(db, billRepo, accountRepo, transactionRepo)
	summaryService := services.NewSummaryService(db)

	accountController := controllers.NewAccountController(accountService, summaryService)
	transactionController := controllers.NewTransactionController(transactionService)
	billController := controllers.NewBillController(billService)
	summaryController := controllers.NewSummaryController(summaryService)

	// Initialize the router
	router := gin.Default()

	// Logger
	router.Use(gin.Logger())
	log.SetPrefix("[APP] ")
	log.SetOutput(gin.DefaultWriter)

	// Recovers the panics and returns the 500. Note that we should not panic in the app anyway
	router.Use(gin.Recovery())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	routes.RegisterAccountRoutes(router, accountController)
	routes.RegisterTransactionRoutes(router, transactionController)
	routes.RegisterSummaryRoutes(router, summaryController)
	routes.RegisterBillRoutes(router, billController)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
