package response

type AccountBillingResponse struct {
	Amount   int    `json:"amount,omitempty"`
	Currency string `json:"currency,omitempty"`
}
