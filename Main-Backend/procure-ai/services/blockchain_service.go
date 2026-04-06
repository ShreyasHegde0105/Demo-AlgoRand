package services

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"procure-ai/models"

	"gorm.io/gorm"
)

const (
	escrowStatusCreated   = uint64(0)
	escrowStatusFunded    = uint64(1)
	escrowStatusDelivered = uint64(2)
	escrowStatusReleased  = uint64(3)
)

type BlockchainService struct {
	db                  *gorm.DB
	contractProjectPath string
	helperScriptPath    string
	pythonExecutable    string
}

type createEscrowCLIResponse struct {
	OrderID           string `json:"orderId"`
	AppID             uint64 `json:"appId"`
	AppAddress        string `json:"appAddress"`
	BuyerAddress      string `json:"buyerAddress"`
	SellerAddress     string `json:"sellerAddress"`
	AgentAddress      string `json:"agentAddress"`
	ApproverAddress   string `json:"approverAddress"`
	EscrowAmountMicro uint64 `json:"escrowAmountMicroAlgos"`
	QuoteValidUntil   uint64 `json:"quoteValidUntil"`
	AlgorandNetwork   string `json:"algorandNetwork"`
}

type prepareActionCLIResponse struct {
	OrderID         string                       `json:"orderId"`
	AppID           uint64                       `json:"appId"`
	AppAddress      string                       `json:"appAddress"`
	Action          string                       `json:"action"`
	Transactions    []models.PreparedTransaction `json:"transactions"`
	AlgorandNetwork string                       `json:"algorandNetwork"`
}

type escrowStateCLIResponse struct {
	OrderID         string                 `json:"orderId"`
	AppID           uint64                 `json:"appId"`
	AppAddress      string                 `json:"appAddress"`
	AlgorandNetwork string                 `json:"algorandNetwork"`
	State           map[string]interface{} `json:"state"`
	Status          interface{}            `json:"status"`
}

func NewBlockchainService(db *gorm.DB) *BlockchainService {
	contractProjectPath := resolveContractProjectPath()
	return &BlockchainService{
		db:                  db,
		contractProjectPath: contractProjectPath,
		helperScriptPath:    filepath.Join(contractProjectPath, "scripts", "escrow_cli.py"),
		pythonExecutable:    resolveContractPythonExecutable(contractProjectPath),
	}
}

func (s *BlockchainService) LockFunds(orderID string) (string, error) {
	return s.writeLegacyTransaction("lock", orderID)
}

func (s *BlockchainService) ReleasePayment(orderID string) (string, error) {
	return s.writeLegacyTransaction("release", orderID)
}

func (s *BlockchainService) CreateEscrow(req models.CreateEscrowRequest) (*models.CreateEscrowResponse, error) {
	var cliResp createEscrowCLIResponse
	if err := s.runHelper(&cliResp,
		"create-escrow",
		"--order-id", req.OrderID,
		"--buyer", req.BuyerAddress,
		"--seller", req.SellerAddress,
		"--agent", req.AgentAddress,
		"--approver", req.ApproverAddress,
		"--amount", strconv.FormatUint(req.EscrowAmountMicro, 10),
		"--quote-valid-until", strconv.FormatUint(req.QuoteValidUntil, 10),
	); err != nil {
		return nil, err
	}

	return &models.CreateEscrowResponse{
		OrderID:           cliResp.OrderID,
		AppID:             cliResp.AppID,
		AppAddress:        cliResp.AppAddress,
		BuyerAddress:      cliResp.BuyerAddress,
		SellerAddress:     cliResp.SellerAddress,
		AgentAddress:      cliResp.AgentAddress,
		ApproverAddress:   cliResp.ApproverAddress,
		EscrowAmountMicro: cliResp.EscrowAmountMicro,
		QuoteValidUntil:   cliResp.QuoteValidUntil,
		AlgorandNetwork:   cliResp.AlgorandNetwork,
		Status:            "escrow_created",
	}, nil
}

func (s *BlockchainService) PrepareSelectSupplier(order *models.Order, selectedSupplier string, quoteID string) (*models.PrepareEscrowActionResponse, error) {
	var cliResp prepareActionCLIResponse
	if err := s.runHelper(&cliResp,
		"prepare-select-supplier",
		"--order-id", order.ID,
		"--app-id", strconv.FormatUint(order.AlgorandAppID, 10),
		"--agent", order.AgentAddress,
		"--selected-supplier", selectedSupplier,
		"--quote-id", quoteID,
	); err != nil {
		return nil, err
	}

	return &models.PrepareEscrowActionResponse{
		OrderID:         cliResp.OrderID,
		AppID:           cliResp.AppID,
		AppAddress:      cliResp.AppAddress,
		Action:          cliResp.Action,
		Transactions:    cliResp.Transactions,
		AlgorandNetwork: cliResp.AlgorandNetwork,
		Status:          order.Status,
	}, nil
}

func (s *BlockchainService) PrepareApprove(order *models.Order) (*models.PrepareEscrowActionResponse, error) {
	var cliResp prepareActionCLIResponse
	if err := s.runHelper(&cliResp,
		"prepare-approve",
		"--order-id", order.ID,
		"--app-id", strconv.FormatUint(order.AlgorandAppID, 10),
		"--approver", order.ApproverAddress,
	); err != nil {
		return nil, err
	}

	return &models.PrepareEscrowActionResponse{
		OrderID:         cliResp.OrderID,
		AppID:           cliResp.AppID,
		AppAddress:      cliResp.AppAddress,
		Action:          cliResp.Action,
		Transactions:    cliResp.Transactions,
		AlgorandNetwork: cliResp.AlgorandNetwork,
		Status:          order.Status,
	}, nil
}

func (s *BlockchainService) PrepareFund(order *models.Order) (*models.PrepareEscrowActionResponse, error) {
	var cliResp prepareActionCLIResponse
	if err := s.runHelper(&cliResp,
		"prepare-fund",
		"--order-id", order.ID,
		"--app-id", strconv.FormatUint(order.AlgorandAppID, 10),
		"--buyer", order.BuyerAddress,
		"--amount", strconv.FormatUint(order.EscrowAmountMicro, 10),
	); err != nil {
		return nil, err
	}

	return &models.PrepareEscrowActionResponse{
		OrderID:         cliResp.OrderID,
		AppID:           cliResp.AppID,
		AppAddress:      cliResp.AppAddress,
		Action:          cliResp.Action,
		Transactions:    cliResp.Transactions,
		AlgorandNetwork: cliResp.AlgorandNetwork,
		Status:          order.Status,
	}, nil
}

func (s *BlockchainService) PrepareRelease(order *models.Order) (*models.PrepareEscrowActionResponse, error) {
	var cliResp prepareActionCLIResponse
	if err := s.runHelper(&cliResp,
		"prepare-release",
		"--order-id", order.ID,
		"--app-id", strconv.FormatUint(order.AlgorandAppID, 10),
		"--buyer", order.BuyerAddress,
		"--approver", order.ApproverAddress,
	); err != nil {
		return nil, err
	}

	return &models.PrepareEscrowActionResponse{
		OrderID:         cliResp.OrderID,
		AppID:           cliResp.AppID,
		AppAddress:      cliResp.AppAddress,
		Action:          cliResp.Action,
		Transactions:    cliResp.Transactions,
		AlgorandNetwork: cliResp.AlgorandNetwork,
		Status:          order.Status,
	}, nil
}

func (s *BlockchainService) GetEscrowState(order *models.Order) (*escrowStateCLIResponse, error) {
	var cliResp escrowStateCLIResponse
	if err := s.runHelper(&cliResp,
		"get-state",
		"--order-id", order.ID,
		"--app-id", strconv.FormatUint(order.AlgorandAppID, 10),
	); err != nil {
		return nil, err
	}
	return &cliResp, nil
}

func (s *BlockchainService) runHelper(out interface{}, args ...string) error {
	if _, err := os.Stat(s.helperScriptPath); err != nil {
		return fmt.Errorf("missing contract helper script at %s", s.helperScriptPath)
	}

	cmdArgs := append([]string{s.helperScriptPath}, args...)
	cmd := exec.Command(s.pythonExecutable, cmdArgs...)
	cmd.Dir = s.contractProjectPath
	cmd.Env = append(os.Environ(), s.workspaceScopedPythonEnv()...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("contract helper failed: %s", message)
	}

	if err := json.Unmarshal(stdout.Bytes(), out); err != nil {
		return fmt.Errorf("failed to parse contract helper output: %w", err)
	}
	return nil
}

func (s *BlockchainService) workspaceScopedPythonEnv() []string {
	localAppData := filepath.Join(s.contractProjectPath, ".localappdata")
	appData := filepath.Join(s.contractProjectPath, ".appdata")
	return []string{
		"LOCALAPPDATA=" + localAppData,
		"APPDATA=" + appData,
	}
}

func (s *BlockchainService) writeLegacyTransaction(prefix, orderID string) (string, error) {
	if orderID == "" {
		return "", fmt.Errorf("orderId is required")
	}

	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	txID := fmt.Sprintf("%s-%s-%s", prefix, orderID, hex.EncodeToString(randomBytes))
	if err := s.db.Model(&models.Order{}).Where("id = ?", orderID).Update("payment_tx_id", txID).Error; err != nil {
		return "", err
	}

	return txID, nil
}

func resolveContractProjectPath() string {
	if configured := os.Getenv("PROCURE_CONTRACTS_PATH"); configured != "" {
		return configured
	}

	candidates := []string{
		"../procure-contracts",
		"procure-contracts",
		filepath.Join("..", "procure-contracts"),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			absPath, absErr := filepath.Abs(candidate)
			if absErr == nil {
				return absPath
			}
			return candidate
		}
	}

	return filepath.Join("..", "procure-contracts")
}

func resolveContractPythonExecutable(contractProjectPath string) string {
	if configured := os.Getenv("PROCURE_CONTRACTS_PYTHON"); configured != "" {
		if filepath.IsAbs(configured) {
			return configured
		}
		if !strings.EqualFold(configured, "python") && !strings.EqualFold(configured, "python.exe") {
			return configured
		}
	}

	candidates := []string{
		filepath.Join(contractProjectPath, ".venv", "Scripts", "python.exe"),
		filepath.Join(contractProjectPath, ".venv", "Scripts", "python"),
		filepath.Join(contractProjectPath, ".venv", "bin", "python3"),
		filepath.Join(contractProjectPath, ".venv", "bin", "python"),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			if absPath, absErr := filepath.Abs(candidate); absErr == nil {
				return absPath
			}
			return candidate
		}
	}

	if configured := os.Getenv("PROCURE_CONTRACTS_PYTHON"); configured != "" {
		return configured
	}

	return "python"
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
