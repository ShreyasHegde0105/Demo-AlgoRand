package services

import (
	"encoding/json"
	"fmt"
	"strconv"

	"procure-ai/models"
)

const demoEscrowAmountMicro uint64 = 10_000

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
	if order.AlgorandAppID != 0 {
		return nil, fmt.Errorf("order %s uses Algorand escrow; use /blockchain/prepare-fund and /blockchain/confirm-fund", orderID)
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
	if order.AlgorandAppID != 0 {
		return nil, fmt.Errorf("order %s uses Algorand escrow; use /blockchain/prepare-release and /blockchain/confirm-release", orderID)
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

	return &models.ConfirmDeliveryResponse{
		Message: "delivery confirmed",
		TxID:    "",
		Order:   order,
		Status:  order.Status,
	}, nil
}

func (s *ProcurementService) CreateEscrow(req models.CreateEscrowRequest) (*models.CreateEscrowResponse, error) {
	order, err := s.orderService.GetOrder(req.OrderID)
	if err != nil {
		return nil, err
	}
	if order.AlgorandAppID != 0 {
		return nil, fmt.Errorf("order %s already has escrow app %d", order.ID, order.AlgorandAppID)
	}

	req.EscrowAmountMicro = normaliseDemoEscrowAmount(req.EscrowAmountMicro)

	response, err := s.blockchainService.CreateEscrow(req)
	if err != nil {
		return nil, err
	}

	if _, err := s.orderService.SaveEscrowDetails(
		order.ID,
		req.BuyerAddress,
		req.SellerAddress,
		req.AgentAddress,
		req.ApproverAddress,
		req.EscrowAmountMicro,
		response.AppID,
		response.AppAddress,
		response.AlgorandNetwork,
		req.QuoteValidUntil,
	); err != nil {
		return nil, err
	}

	return response, nil
}

func normaliseDemoEscrowAmount(requested uint64) uint64 {
	if requested == 0 {
		return demoEscrowAmountMicro
	}
	if requested > demoEscrowAmountMicro {
		return demoEscrowAmountMicro
	}
	return requested
}

func (s *ProcurementService) PrepareSelectSupplier(req models.SetSelectedSupplierRequest) (*models.PrepareEscrowActionResponse, error) {
	order, err := s.orderService.GetOrder(req.OrderID)
	if err != nil {
		return nil, err
	}
	if order.AlgorandAppID == 0 {
		return nil, fmt.Errorf("order %s does not have an escrow app yet", req.OrderID)
	}
	if order.AgentAddress == "" {
		return nil, fmt.Errorf("order %s is missing agent wallet address", req.OrderID)
	}

	return s.blockchainService.PrepareSelectSupplier(order, req.SelectedSupplier, req.QuoteID)
}

func (s *ProcurementService) ConfirmSelectedSupplier(orderID string) (*models.EscrowActionResponse, error) {
	order, err := s.orderService.GetOrder(orderID)
	if err != nil {
		return nil, err
	}
	if order.AlgorandAppID == 0 {
		return nil, fmt.Errorf("order %s does not have an escrow app yet", orderID)
	}

	state, err := s.blockchainService.GetEscrowState(order)
	if err != nil {
		return nil, err
	}

	selectedSupplier, err := readEscrowString(state.State, "selected_supplier")
	if err != nil {
		return nil, err
	}
	quoteID, err := readEscrowString(state.State, "quote_id")
	if err != nil {
		return nil, err
	}
	if selectedSupplier == "" || quoteID == "" {
		return nil, fmt.Errorf("escrow app %d does not have supplier selection saved yet", order.AlgorandAppID)
	}

	order, err = s.orderService.SaveSelectedSupplier(order.ID, selectedSupplier, quoteID)
	if err != nil {
		return nil, err
	}

	return &models.EscrowActionResponse{
		Message: "selected supplier confirmed on-chain",
		Order:   order,
		Status:  order.Status,
	}, nil
}

func (s *ProcurementService) PrepareApproveEscrow(orderID string) (*models.PrepareEscrowActionResponse, error) {
	order, err := s.orderService.GetOrder(orderID)
	if err != nil {
		return nil, err
	}
	if order.AlgorandAppID == 0 {
		return nil, fmt.Errorf("order %s does not have an escrow app yet", orderID)
	}
	if order.ApproverAddress == "" {
		return nil, fmt.Errorf("order %s is missing approver wallet address", orderID)
	}

	return s.blockchainService.PrepareApprove(order)
}

func (s *ProcurementService) ConfirmApproveEscrow(orderID string) (*models.EscrowActionResponse, error) {
	order, err := s.orderService.GetOrder(orderID)
	if err != nil {
		return nil, err
	}
	if order.AlgorandAppID == 0 {
		return nil, fmt.Errorf("order %s does not have an escrow app yet", orderID)
	}

	state, err := s.blockchainService.GetEscrowState(order)
	if err != nil {
		return nil, err
	}
	approved, err := readEscrowBool(state.State, "approved")
	if err != nil {
		return nil, err
	}
	if !approved {
		return nil, fmt.Errorf("escrow app %d has not been approved yet", order.AlgorandAppID)
	}

	order, err = s.orderService.MarkEscrowApproved(order.ID)
	if err != nil {
		return nil, err
	}

	return &models.EscrowActionResponse{
		Message: "escrow approval confirmed on-chain",
		Order:   order,
		Status:  order.Status,
	}, nil
}

func (s *ProcurementService) PrepareFund(orderID string) (*models.PrepareEscrowActionResponse, error) {
	order, err := s.orderService.GetOrder(orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != "approved" {
		return nil, fmt.Errorf("order %s must be approved before preparing escrow funding", orderID)
	}
	if order.AlgorandAppID == 0 {
		return nil, fmt.Errorf("order %s does not have an escrow app yet", orderID)
	}
	if order.BuyerAddress == "" {
		return nil, fmt.Errorf("order %s is missing buyer wallet address", orderID)
	}
	if order.EscrowAmountMicro == 0 {
		return nil, fmt.Errorf("order %s is missing escrow amount in microAlgos", orderID)
	}

	return s.blockchainService.PrepareFund(order)
}

func (s *ProcurementService) ConfirmFund(req models.ConfirmEscrowActionRequest) (*models.PaymentActionResponse, error) {
	order, err := s.orderService.GetOrder(req.OrderID)
	if err != nil {
		return nil, err
	}
	if order.AlgorandAppID == 0 {
		return nil, fmt.Errorf("order %s does not have an escrow app yet", req.OrderID)
	}

	state, err := s.blockchainService.GetEscrowState(order)
	if err != nil {
		return nil, err
	}
	status, err := normaliseEscrowStatus(state.Status)
	if err != nil {
		return nil, err
	}
	if status != escrowStatusFunded {
		return nil, fmt.Errorf("escrow app %d is not funded yet", order.AlgorandAppID)
	}

	if _, err := s.orderService.RecordFundingTransaction(order.ID, req.TxID); err != nil {
		return nil, err
	}
	order, err = s.orderService.MarkFundsLocked(order.ID)
	if err != nil {
		return nil, err
	}

	return &models.PaymentActionResponse{
		OrderID: order.ID,
		TxID:    req.TxID,
		Status:  order.Status,
	}, nil
}

func (s *ProcurementService) PrepareRelease(orderID string) (*models.PrepareEscrowActionResponse, error) {
	order, err := s.orderService.GetOrder(orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != "funds_locked" && order.Status != "delivered" {
		return nil, fmt.Errorf("order %s must be funds_locked or delivered before preparing release", orderID)
	}
	if order.AlgorandAppID == 0 {
		return nil, fmt.Errorf("order %s does not have an escrow app yet", orderID)
	}
	if order.BuyerAddress == "" {
		return nil, fmt.Errorf("order %s is missing buyer wallet address", orderID)
	}
	if order.ApproverAddress == "" {
		return nil, fmt.Errorf("order %s is missing approver wallet address", orderID)
	}
	if order.SelectedSupplier == "" {
		return nil, fmt.Errorf("order %s does not have a selected supplier saved yet", orderID)
	}
	if !order.EscrowApproved {
		return nil, fmt.Errorf("order %s escrow has not been approved on-chain yet", orderID)
	}

	return s.blockchainService.PrepareRelease(order)
}

func (s *ProcurementService) ConfirmRelease(req models.ConfirmEscrowActionRequest) (*models.PaymentActionResponse, error) {
	order, err := s.orderService.GetOrder(req.OrderID)
	if err != nil {
		return nil, err
	}
	if order.AlgorandAppID == 0 {
		return nil, fmt.Errorf("order %s does not have an escrow app yet", req.OrderID)
	}

	state, err := s.blockchainService.GetEscrowState(order)
	if err != nil {
		return nil, err
	}
	status, err := normaliseEscrowStatus(state.Status)
	if err != nil {
		return nil, err
	}
	if status != escrowStatusReleased {
		return nil, fmt.Errorf("escrow app %d has not released payment yet", order.AlgorandAppID)
	}

	if _, err := s.orderService.RecordReleaseTransaction(order.ID, req.TxID); err != nil {
		return nil, err
	}
	settlementSupplier, err := readEscrowString(state.State, "settlement_supplier")
	if err != nil {
		return nil, err
	}
	settlementAmount, err := readEscrowUint64(state.State, "settlement_amount")
	if err != nil {
		return nil, err
	}
	settlementTxn, err := readEscrowString(state.State, "settlement_txn")
	if err != nil {
		return nil, err
	}
	if settlementTxn == "" {
		settlementTxn = req.TxID
	}
	if _, err := s.orderService.SaveSettlementDetails(order.ID, settlementSupplier, settlementAmount, settlementTxn); err != nil {
		return nil, err
	}
	if order.Status == "funds_locked" {
		if _, err := s.orderService.ConfirmDelivery(order.ID); err != nil {
			return nil, err
		}
	}
	order, err = s.orderService.MarkPaymentReleased(order.ID)
	if err != nil {
		return nil, err
	}

	return &models.PaymentActionResponse{
		OrderID: order.ID,
		TxID:    req.TxID,
		Status:  order.Status,
	}, nil
}

func readEscrowUint64(state map[string]interface{}, key string) (uint64, error) {
	value, ok := state[key]
	if !ok {
		return 0, fmt.Errorf("escrow state missing %s", key)
	}
	return normaliseEscrowStatus(value)
}

func readEscrowBool(state map[string]interface{}, key string) (bool, error) {
	value, err := readEscrowUint64(state, key)
	if err != nil {
		return false, err
	}
	return value == 1, nil
}

func readEscrowString(state map[string]interface{}, key string) (string, error) {
	value, ok := state[key]
	if !ok {
		return "", fmt.Errorf("escrow state missing %s", key)
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case json.Number:
		return v.String(), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func normaliseEscrowStatus(value interface{}) (uint64, error) {
	switch v := value.(type) {
	case float64:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	case uint64:
		return v, nil
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return 0, err
		}
		return uint64(parsed), nil
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("unexpected escrow status type %T", value)
	}
}
