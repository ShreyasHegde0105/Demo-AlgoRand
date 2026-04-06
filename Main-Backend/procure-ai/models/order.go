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
