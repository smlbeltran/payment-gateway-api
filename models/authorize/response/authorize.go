package response

type Verfication struct {
	AuthorizationId string `json:"authorization_id,omitempty"`
	Status          int    `json:"status,omitempty"`
	Amount          int    `json:"amount,omitempty"`
	Currency        string `json:"currency,omitempty"`
}
