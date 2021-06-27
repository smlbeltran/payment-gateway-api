package response

type Verfication struct {
	VerificationId string `json:"verification_id,omitempty"`
	Status         int    `json:"status,omitempty"`
	Amount         int    `json:"amount,omitempty"`
	Currency       string `json:"currency,omitempty"`
}
