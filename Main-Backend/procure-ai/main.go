package main

import (
	"log"
	"os"
	"path/filepath"

	"procure-ai/controllers"
	"procure-ai/db"
	"procure-ai/routes"
	"procure-ai/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func main() {
	loadEnvFile()

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
	router.Use(corsMiddleware())

	vendorService := services.NewVendorService(database)
	agentService := services.NewAgentService(database, vendorService)
	orderService := services.NewOrderService(database, vendorService)
	blockchainService := services.NewBlockchainService(database)
	procurementService := services.NewProcurementService(agentService, orderService, blockchainService)
	qrService := services.NewQRService(database, orderService)
	geminiParserService := services.NewGeminiParserServiceFromEnv()

	controller := controllers.NewController(
		vendorService,
		agentService,
		orderService,
		procurementService,
		qrService,
		geminiParserService,
	)

	routes.RegisterRoutes(router, controller)

	log.Println("Autonomous Procurement Agent backend running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func loadEnvFile() {
	// Support starting the binary from either project root or parent workspace.
	candidates := []string{
		".env",
		filepath.Join("procure-ai", ".env"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if err := godotenv.Overload(path); err != nil {
			log.Printf("warning: failed to load %s: %v", path, err)
			continue
		}
		log.Printf("loaded environment variables from %s", path)
		return
	}
}
