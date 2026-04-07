export type Vendor = {
  id: number
  name: string
  price: number
  trust: number
  deliveryDays: number
  stock: number
  minOrderQty: number
  location: string
  paymentTerms: string
  reliabilityScore: number
  category: string
}

export type RankedVendor = {
  rank: number
  vendor: Vendor
  estimatedTotal: number
  scoreBreakdown: {
    price: number
    delivery: number
    trust: number
    reliability: number
    final: number
  }
  reason: string
}

export type ProcurementRecommendationResponse = {
  recommendationId?: string
  recommendedVendor?: RankedVendor
  topVendors: RankedVendor[]
  rejectedVendors: { vendor: Vendor; reasons: string[] }[]
  appliedWeights: {
    price: number
    delivery: number
    trust: number
    reliability: number
  }
  summary: string
}

export type Order = {
  id: string
  vendor: string
  vendorId: number
  category: string
  quantity: number
  unitPrice: number
  amount: number
  status: string
  recommendationId?: string
  buyerAddress?: string
  sellerAddress?: string
  agentAddress?: string
  approverAddress?: string
  escrowAmountMicroAlgos?: number
  algorandAppId?: number
  algorandAppAddress?: string
  algorandNetwork?: string
  selectedSupplier?: string
  quoteId?: string
  quoteValidUntil?: number
  escrowApproved: boolean
  fundingTxId?: string
  releaseTxId?: string
  settlementSupplier?: string
  settlementAmountMicroAlgos?: number
  settlementTxId?: string
  createdAt: string
  updatedAt: string
}

export type CreateEscrowResponse = {
  orderId: string
  appId: number
  appAddress: string
  status: string
  algorandNetwork: string
}

export type PreparedTransaction = {
  index: number
  description: string
  txnBase64: string
}

export type PrepareEscrowActionResponse = {
  orderId: string
  appId: number
  appAddress: string
  action: string
  transactions: PreparedTransaction[]
  algorandNetwork: string
  status: string
}

export type OrderActionResponse = {
  message: string
  order?: Order
  status: string
}
