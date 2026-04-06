package models

type ProcurementRequest struct {
	Category        string   `json:"category" binding:"required"`
	Quantity        int      `json:"quantity" binding:"required,gt=0"`
	Budget          float64  `json:"budget" binding:"omitempty,gt=0"`
	MaxDeliveryDays int      `json:"maxDeliveryDays" binding:"omitempty,gte=0"`
	PreferredCities []string `json:"preferredCities"`
	TopN            int      `json:"topN" binding:"omitempty,gte=1,lte=10"`
}

type AgentWeights struct {
	Price       float64 `json:"price"`
	Delivery    float64 `json:"delivery"`
	Trust       float64 `json:"trust"`
	Reliability float64 `json:"reliability"`
}

type VendorScoreBreakdown struct {
	Price       float64 `json:"price"`
	Delivery    float64 `json:"delivery"`
	Trust       float64 `json:"trust"`
	Reliability float64 `json:"reliability"`
	Final       float64 `json:"final"`
}

type RankedVendor struct {
	Rank           int                  `json:"rank"`
	Vendor         Vendor               `json:"vendor"`
	EstimatedTotal float64              `json:"estimatedTotal"`
	ScoreBreakdown VendorScoreBreakdown `json:"scoreBreakdown"`
	Reason         string               `json:"reason"`
}

type RejectedVendor struct {
	Vendor  Vendor    `json:"vendor"`
	Reasons []string  `json:"reasons"`
}

type ProcurementRecommendationResponse struct {
	RecommendationID string              `json:"recommendationId,omitempty"`
	RecommendedVendor *RankedVendor      `json:"recommendedVendor,omitempty"`
	TopVendors        []RankedVendor     `json:"topVendors"`
	RejectedVendors   []RejectedVendor   `json:"rejectedVendors"`
	AppliedWeights    AgentWeights       `json:"appliedWeights"`
	Summary           string             `json:"summary"`
}
