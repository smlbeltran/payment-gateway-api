package request

type Refund struct {
	TransactionId string `json:"transaction_id,omitempty"`
	Amount        int    `json:"amount,omitempty"`
}
