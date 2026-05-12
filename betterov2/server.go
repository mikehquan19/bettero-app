package main

import (
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

	var database *pgxpool.Pool = setup.ConnectDB()
	accountService := services.NewAccountService(database)

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	routes.RegisterAccountRoutes(router, accountService)
	routes.RegisterTransactionRoutes(router)
	routes.RegisterSummaryRoutes(router)
	routes.RegisterBillRoutes(router)

	router.Run(":8080")
}
