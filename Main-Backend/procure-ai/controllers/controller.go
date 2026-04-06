package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"procure-ai/models"
	"procure-ai/services"
)

type Controller struct {
	vendorService      *services.VendorService
	agentService       *services.AgentService
	orderService       *services.OrderService
	procurementService *services.ProcurementService
	qrService          *services.QRService
}

func NewController(
	vendorService *services.VendorService,
	agentService *services.AgentService,
	orderService *services.OrderService,
	procurementService *services.ProcurementService,
	qrService *services.QRService,
) *Controller {
	return &Controller{
		vendorService:      vendorService,
		agentService:       agentService,
		orderService:       orderService,
		procurementService: procurementService,
		qrService:          qrService,
	}
}

func (ctl *Controller) GetVendors(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"vendors": ctl.vendorService.GetVendors(),
	})
}

func (ctl *Controller) SelectVendor(c *gin.Context) {
	var req models.SelectVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recommendation, err := ctl.vendorService.SelectBestVendor(req.Vendors)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, recommendation)
}

func (ctl *Controller) RecommendVendors(c *gin.Context) {
	var req models.ProcurementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recommendation, err := ctl.agentService.RecommendVendors(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctl.agentService.SaveRecommendationSession(req, recommendation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, recommendation)
}

func (ctl *Controller) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := ctl.procurementService.CreateOrder(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (ctl *Controller) ApproveOrder(c *gin.Context) {
	var req models.ApproveOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := ctl.procurementService.ApproveOrder(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctl *Controller) LockFunds(c *gin.Context) {
	var req models.PaymentActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := ctl.procurementService.LockFunds(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctl *Controller) ReleasePayment(c *gin.Context) {
	var req models.PaymentActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := ctl.procurementService.ReleasePayment(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctl *Controller) GenerateQR(c *gin.Context) {
	var req models.GenerateQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := ctl.qrService.GenerateQR(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctl *Controller) VerifyQR(c *gin.Context) {
	var req models.VerifyQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := ctl.qrService.VerifyQR(req.OrderID, req.QRCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctl *Controller) ConfirmDelivery(c *gin.Context) {
	var req models.ConfirmDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := ctl.procurementService.ConfirmDelivery(req.OrderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
