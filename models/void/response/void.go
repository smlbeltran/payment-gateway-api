package response

type VoidResponse struct {
	Status   int    `json:"status,omitempty"`
	Amount   int    `json:"amount,omitempty"`
	Currency string `json:"currency,omitempty"`
}
