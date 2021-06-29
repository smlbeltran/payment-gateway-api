package response

type CaptureResponse struct {
	Amount   int    `json:"amount,omitempty"`
	Captured int    `json:"captured,omitempty"`
	Currency string `json:"currency,omitempty"`
}
