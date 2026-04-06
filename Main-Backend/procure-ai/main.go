package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"procure-ai/controllers"
	"procure-ai/db"
	"procure-ai/routes"
	"procure-ai/services"
)

func main() {
	database, err := db.NewPostgresDB()
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(database); err != nil {
		log.Fatal(err)
	}

	if err := db.SeedVendors(database); err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	vendorService := services.NewVendorService(database)
	agentService := services.NewAgentService(database, vendorService)
	orderService := services.NewOrderService(database, vendorService)
	blockchainService := services.NewBlockchainService(database)
	procurementService := services.NewProcurementService(agentService, orderService, blockchainService)
	qrService := services.NewQRService(database, orderService)

	controller := controllers.NewController(
		vendorService,
		agentService,
		orderService,
		procurementService,
		qrService,
	)

	routes.RegisterRoutes(router, controller)

	log.Println("Autonomous Procurement Agent backend running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
