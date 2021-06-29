package response

type VoidResponse struct {
	Status   int    `json:"status"`
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}
