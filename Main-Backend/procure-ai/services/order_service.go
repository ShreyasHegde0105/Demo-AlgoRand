package services

import (
	"errors"
	"fmt"

	"procure-ai/models"

	"gorm.io/gorm"
)

type OrderService struct {
	db            *gorm.DB
	vendorService *VendorService
}

func NewOrderService(db *gorm.DB, vendorService *VendorService) *OrderService {
	return &OrderService{
		db:            db,
		vendorService: vendorService,
	}
}

func (s *OrderService) CreateOrder(req models.CreateOrderRequest) (*models.Order, error) {
	if req.Vendor == "" {
		return nil, errors.New("vendor is required")
	}
	if req.Quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}

	vendor, err := s.vendorService.GetVendorByName(req.Vendor)
	if err != nil {
		return nil, err
	}
	if req.Quantity > vendor.Stock {
		return nil, fmt.Errorf("requested quantity %d exceeds vendor stock %d", req.Quantity, vendor.Stock)
	}
	if req.Quantity < vendor.MinOrderQty {
		return nil, fmt.Errorf("requested quantity %d is below vendor minimum order quantity %d", req.Quantity, vendor.MinOrderQty)
	}

	var count int64
	if err := s.db.Model(&models.Order{}).Count(&count).Error; err != nil {
		return nil, err
	}

	orderID := fmt.Sprintf("ORD-%04d", count+1)
	amount := vendor.Price * float64(req.Quantity)
	order := &models.Order{
		ID:                orderID,
		Vendor:            vendor.Name,
		VendorID:          vendor.ID,
		Category:          vendor.Category,
		Quantity:          req.Quantity,
		UnitPrice:         vendor.Price,
		Amount:            amount,
		Status:            "pending_approval",
		RecommendationID:  req.RecommendationID,
		SelectionReason:   req.SelectionReason,
		AgentScore:        req.AgentScore,
		ShortlistSnapshot: req.ShortlistSnapshot,
	}

	if err := s.db.Create(order).Error; err != nil {
		return nil, err
	}

	return s.GetOrder(orderID)
}

func (s *OrderService) GetOrder(orderID string) (*models.Order, error) {
	var order models.Order
	if err := s.db.Preload("QR").First(&order, "id = ?", orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("order %s not found", orderID)
		}
		return nil, err
	}

	return &order, nil
}

func (s *OrderService) ConfirmDelivery(orderID string) (*models.Order, error) {
	order, err := s.GetOrder(orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != "funds_locked" {
		return nil, fmt.Errorf("order %s must be funds_locked before delivery confirmation", orderID)
	}

	if err := s.db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", "delivered").Error; err != nil {
		return nil, err
	}

	return s.GetOrder(orderID)
}

func (s *OrderService) ApproveOrder(orderID string) (*models.Order, error) {
	order, err := s.GetOrder(orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != "pending_approval" {
		return nil, fmt.Errorf("order %s is not pending approval", orderID)
	}

	if err := s.db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", "approved").Error; err != nil {
		return nil, err
	}

	return s.GetOrder(orderID)
}

func (s *OrderService) MarkFundsLocked(orderID string) (*models.Order, error) {
	order, err := s.GetOrder(orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != "approved" {
		return nil, fmt.Errorf("order %s must be approved before locking funds", orderID)
	}

	if err := s.db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", "funds_locked").Error; err != nil {
		return nil, err
	}

	return s.GetOrder(orderID)
}

func (s *OrderService) MarkPaymentReleased(orderID string) (*models.Order, error) {
	order, err := s.GetOrder(orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != "delivered" {
		return nil, fmt.Errorf("order %s must be delivered before payment release", orderID)
	}

	if err := s.db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", "payment_released").Error; err != nil {
		return nil, err
	}

	return s.GetOrder(orderID)
}
