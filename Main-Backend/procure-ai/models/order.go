package models

import "time"

type Order struct {
	ID                 string    `json:"id" gorm:"primaryKey;size:32"`
	Vendor             string    `json:"vendor" gorm:"not null"`
	VendorID           uint      `json:"vendorId" gorm:"index;not null"`
	VendorRef          Vendor    `json:"-" gorm:"foreignKey:VendorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Category           string    `json:"category" gorm:"not null;default:''"`
	Quantity           int       `json:"quantity" gorm:"not null;default:0"`
	UnitPrice          float64   `json:"unitPrice" gorm:"not null;default:0"`
	Amount             float64   `json:"amount" gorm:"not null"`
	Status             string    `json:"status" gorm:"not null"`
	RecommendationID   string    `json:"recommendationId,omitempty" gorm:"index;default:''"`
	SelectionReason    string    `json:"selectionReason,omitempty" gorm:"type:text;not null;default:''"`
	AgentScore         float64   `json:"agentScore,omitempty" gorm:"not null;default:0"`
	ShortlistSnapshot  string    `json:"shortlistSnapshot,omitempty" gorm:"type:text;not null;default:''"`
	PaymentTxID        string    `json:"paymentTxId,omitempty"`
	BuyerAddress       string    `json:"buyerAddress,omitempty" gorm:"not null;default:''"`
	SellerAddress      string    `json:"sellerAddress,omitempty" gorm:"not null;default:''"`
	AgentAddress       string    `json:"agentAddress,omitempty" gorm:"not null;default:''"`
	ApproverAddress    string    `json:"approverAddress,omitempty" gorm:"not null;default:''"`
	EscrowAmountMicro  uint64    `json:"escrowAmountMicroAlgos,omitempty" gorm:"not null;default:0"`
	AlgorandAppID      uint64    `json:"algorandAppId,omitempty" gorm:"not null;default:0"`
	AlgorandAppAddress string    `json:"algorandAppAddress,omitempty" gorm:"not null;default:''"`
	AlgorandNetwork    string    `json:"algorandNetwork,omitempty" gorm:"not null;default:''"`
	SelectedSupplier   string    `json:"selectedSupplier,omitempty" gorm:"not null;default:''"`
	QuoteID            string    `json:"quoteId,omitempty" gorm:"not null;default:''"`
	QuoteValidUntil    uint64    `json:"quoteValidUntil,omitempty" gorm:"not null;default:0"`
	EscrowApproved     bool      `json:"escrowApproved" gorm:"not null;default:false"`
	FundingTxID        string    `json:"fundingTxId,omitempty" gorm:"not null;default:''"`
	ReleaseTxID        string    `json:"releaseTxId,omitempty" gorm:"not null;default:''"`
	RefundTxID         string    `json:"refundTxId,omitempty" gorm:"not null;default:''"`
	SettlementSupplier string    `json:"settlementSupplier,omitempty" gorm:"not null;default:''"`
	SettlementAmount   uint64    `json:"settlementAmountMicroAlgos,omitempty" gorm:"not null;default:0"`
	SettlementTxID     string    `json:"settlementTxId,omitempty" gorm:"not null;default:''"`
	QR                 *QR       `json:"qr,omitempty" gorm:"foreignKey:OrderID;references:ID"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type CreateOrderRequest struct {
	RecommendationID  string  `json:"recommendationId" binding:"required"`
	Vendor            string  `json:"vendor" binding:"required"`
	Quantity          int     `json:"quantity" binding:"required,gt=0"`
	SelectionReason   string  `json:"selectionReason"`
	AgentScore        float64 `json:"agentScore"`
	ShortlistSnapshot string  `json:"shortlistSnapshot"`
}

type ApproveOrderRequest struct {
	OrderID string `json:"orderId" binding:"required"`
}

type PaymentActionRequest struct {
	OrderID string `json:"orderId" binding:"required"`
}

type ConfirmDeliveryRequest struct {
	OrderID string `json:"orderId" binding:"required"`
}

type PaymentActionResponse struct {
	OrderID string `json:"orderId"`
	TxID    string `json:"txID"`
	Status  string `json:"status"`
}

type CreateEscrowRequest struct {
	OrderID           string `json:"orderId" binding:"required"`
	BuyerAddress      string `json:"buyerAddress" binding:"required"`
	SellerAddress     string `json:"sellerAddress" binding:"required"`
	AgentAddress      string `json:"agentAddress" binding:"required"`
	ApproverAddress   string `json:"approverAddress" binding:"required"`
	EscrowAmountMicro uint64 `json:"escrowAmountMicroAlgos" binding:"required,gt=0"`
	QuoteValidUntil   uint64 `json:"quoteValidUntil" binding:"required,gt=0"`
}

type CreateEscrowResponse struct {
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
	Status            string `json:"status"`
}

type SetSelectedSupplierRequest struct {
	OrderID          string `json:"orderId" binding:"required"`
	SelectedSupplier string `json:"selectedSupplier" binding:"required"`
	QuoteID          string `json:"quoteId" binding:"required"`
}

type EscrowActionResponse struct {
	Message string `json:"message"`
	Order   *Order `json:"order"`
	Status  string `json:"status"`
}

type PrepareEscrowActionRequest struct {
	OrderID string `json:"orderId" binding:"required"`
}

type ConfirmEscrowActionRequest struct {
	OrderID string `json:"orderId" binding:"required"`
	TxID    string `json:"txID" binding:"required"`
}

type PreparedTransaction struct {
	Index       int    `json:"index"`
	Description string `json:"description"`
	TxnBase64   string `json:"txnBase64"`
}

type PrepareEscrowActionResponse struct {
	OrderID         string                `json:"orderId"`
	AppID           uint64                `json:"appId"`
	AppAddress      string                `json:"appAddress"`
	Action          string                `json:"action"`
	Transactions    []PreparedTransaction `json:"transactions"`
	AlgorandNetwork string                `json:"algorandNetwork"`
	Status          string                `json:"status"`
}

type ConfirmDeliveryResponse struct {
	Message string `json:"message"`
	TxID    string `json:"txID"`
	Order   *Order `json:"order"`
	Status  string `json:"status"`
}

type OrderActionResponse struct {
	Message string `json:"message"`
	Order   *Order `json:"order"`
	Status  string `json:"status"`
}
