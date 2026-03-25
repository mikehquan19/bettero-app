package main

import (
	"betterov2/routes"
	"betterov2/setup"

	"github.com/gin-gonic/gin"
)

func main() {
	// Connect to database
	setup.ConnectDB()

	router := gin.Default()

	routes.RegisterAccountRoutes(router)
	routes.RegisterTransactionRoutes(router)
	routes.RegisterSummaryRoutes(router)
	routes.RegisterBillRoutes(router)

	router.Run(":8080")
}
