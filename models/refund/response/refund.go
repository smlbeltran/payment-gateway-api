package response

type AccountRefundResponse struct {
	Amount   int    `json:"amount,omitempty"`
	Currency string `json:"currency,omitempty"`
}
