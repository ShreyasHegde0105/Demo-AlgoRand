package routes

import (
	"github.com/gin-gonic/gin"
	"procure-ai/controllers"
)

func RegisterRoutes(router *gin.Engine, controller *controllers.Controller) {
	router.GET("/vendors", controller.GetVendors)
	router.POST("/select-vendor", controller.SelectVendor)
	router.POST("/agent/recommend-vendors", controller.RecommendVendors)
	router.POST("/create-order", controller.CreateOrder)
	router.POST("/approve-order", controller.ApproveOrder)
	router.POST("/lock-funds", controller.LockFunds)
	router.POST("/release-payment", controller.ReleasePayment)
	router.POST("/generate-qr", controller.GenerateQR)
	router.POST("/verify-qr", controller.VerifyQR)
	router.POST("/confirm-delivery", controller.ConfirmDelivery)
}
