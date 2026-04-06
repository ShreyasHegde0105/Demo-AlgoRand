package services

import (
	"fmt"

	"procure-ai/models"
)

type ProcurementService struct {
	agentService      *AgentService
	orderService      *OrderService
	blockchainService *BlockchainService
}

func NewProcurementService(agentService *AgentService, orderService *OrderService, blockchainService *BlockchainService) *ProcurementService {
	return &ProcurementService{
		agentService:      agentService,
		orderService:      orderService,
		blockchainService: blockchainService,
	}
}

func (s *ProcurementService) CreateOrder(req models.CreateOrderRequest) (*models.Order, error) {
	selected, session, err := s.agentService.ValidateSelectedVendor(req.RecommendationID, req.Vendor)
	if err != nil {
		return nil, err
	}
	if req.Quantity != session.Quantity {
		return nil, fmt.Errorf("quantity %d does not match recommendation quantity %d", req.Quantity, session.Quantity)
	}
	req.Quantity = session.Quantity
	req.SelectionReason = selected.Reason
	req.AgentScore = selected.ScoreBreakdown.Final
	req.ShortlistSnapshot = session.ShortlistSnapshot
	return s.orderService.CreateOrder(req)
}

func (s *ProcurementService) ApproveOrder(orderID string) (*models.OrderActionResponse, error) {
	order, err := s.orderService.ApproveOrder(orderID)
	if err != nil {
		return nil, err
	}

	return &models.OrderActionResponse{
		Message: "order approved",
		Order:   order,
		Status:  order.Status,
	}, nil
}

func (s *ProcurementService) LockFunds(orderID string) (*models.PaymentActionResponse, error) {
	order, err := s.orderService.GetOrder(orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != "approved" {
		return nil, fmt.Errorf("order %s must be approved before locking funds", orderID)
	}

	txID, err := s.blockchainService.LockFunds(orderID)
	if err != nil {
		return nil, err
	}

	order, err = s.orderService.MarkFundsLocked(orderID)
	if err != nil {
		return nil, err
	}

	return &models.PaymentActionResponse{
		OrderID: order.ID,
		TxID:    txID,
		Status:  order.Status,
	}, nil
}

func (s *ProcurementService) ReleasePayment(orderID string) (*models.PaymentActionResponse, error) {
	order, err := s.orderService.GetOrder(orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != "delivered" {
		return nil, fmt.Errorf("order %s must be delivered before payment release", orderID)
	}

	txID, err := s.blockchainService.ReleasePayment(orderID)
	if err != nil {
		return nil, err
	}

	order, err = s.orderService.MarkPaymentReleased(orderID)
	if err != nil {
		return nil, err
	}

	return &models.PaymentActionResponse{
		OrderID: order.ID,
		TxID:    txID,
		Status:  order.Status,
	}, nil
}

func (s *ProcurementService) ConfirmDelivery(orderID string) (*models.ConfirmDeliveryResponse, error) {
	order, err := s.orderService.ConfirmDelivery(orderID)
	if err != nil {
		return nil, err
	}

	txID, err := s.blockchainService.ReleasePayment(orderID)
	if err != nil {
		return nil, err
	}

	order, err = s.orderService.MarkPaymentReleased(orderID)
	if err != nil {
		return nil, err
	}

	return &models.ConfirmDeliveryResponse{
		Message: "delivery confirmed and payment released",
		TxID:    txID,
		Order:   order,
		Status:  order.Status,
	}, nil
}
