package models

type ProcurementRequest struct {
	Category        string   `json:"category" binding:"required"`
	Quantity        int      `json:"quantity" binding:"required,gt=0"`
	Budget          float64  `json:"budget" binding:"omitempty,gt=0"`
	MaxDeliveryDays int      `json:"maxDeliveryDays" binding:"omitempty,gte=0"`
	PreferredCities []string `json:"preferredCities"`
	TopN            int      `json:"topN" binding:"omitempty,gte=1,lte=10"`
}

type NaturalLanguageRecommendationRequest struct {
	Prompt string `json:"prompt" binding:"required"`
	TopN   int    `json:"topN" binding:"omitempty,gte=1,lte=10"`
}

type ParsedProcurementRequest struct {
	Category        string   `json:"category"`
	Quantity        int      `json:"quantity"`
	UnitBudget      float64  `json:"unitBudget,omitempty"`
	TotalBudget     float64  `json:"totalBudget,omitempty"`
	MaxDeliveryDays int      `json:"maxDeliveryDays,omitempty"`
	PreferredCities []string `json:"preferredCities,omitempty"`
}

type NaturalLanguageRecommendationResponse struct {
	Prompt         string                             `json:"prompt"`
	ParsedRequest  ParsedProcurementRequest           `json:"parsedRequest"`
	Request        ProcurementRequest                 `json:"request"`
	Recommendation *ProcurementRecommendationResponse `json:"recommendation"`
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
	Vendor  Vendor   `json:"vendor"`
	Reasons []string `json:"reasons"`
}

type ProcurementRecommendationResponse struct {
	RecommendationID  string           `json:"recommendationId,omitempty"`
	RecommendedVendor *RankedVendor    `json:"recommendedVendor,omitempty"`
	TopVendors        []RankedVendor   `json:"topVendors"`
	RejectedVendors   []RejectedVendor `json:"rejectedVendors"`
	AppliedWeights    AgentWeights     `json:"appliedWeights"`
	Summary           string           `json:"summary"`
}
