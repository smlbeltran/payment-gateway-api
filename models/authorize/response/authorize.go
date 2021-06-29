package response

type Authorize struct {
	AuthorizationId string `json:"authorization_id"`
	Status          int    `json:"status"`
	Amount          int    `json:"amount"`
	Currency        string `json:"currency"`
}
