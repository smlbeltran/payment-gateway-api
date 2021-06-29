package response

type CaptureResponse struct {
	Amount   int    `json:"amount"`
	Captured int    `json:"captured"`
	Currency string `json:"currency"`
}
