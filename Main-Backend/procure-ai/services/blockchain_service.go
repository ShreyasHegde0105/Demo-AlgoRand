package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"procure-ai/models"

	"gorm.io/gorm"
)

type BlockchainService struct {
	db *gorm.DB
}

func NewBlockchainService(db *gorm.DB) *BlockchainService {
	return &BlockchainService{db: db}
}

func (s *BlockchainService) LockFunds(orderID string) (string, error) {
	return s.writeTransaction("lock", orderID)
}

func (s *BlockchainService) ReleasePayment(orderID string) (string, error) {
	return s.writeTransaction("release", orderID)
}

func (s *BlockchainService) writeTransaction(prefix, orderID string) (string, error) {
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
