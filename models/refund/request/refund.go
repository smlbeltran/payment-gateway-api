package request

type Refund struct {
	AuthorizationId string `json:"authorization_id"`
	Amount          int    `json:"amount"`
}
