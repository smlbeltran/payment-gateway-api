package request

type Account struct {
	AuthorizationId string `json:"authorization_id,omitempty"`
	Amount          int    `json:"amount,omitempty"`
}
