package request

type Capture struct {
	AuthorizationId string `json:"authorization_id"`
	Amount          int    `json:"amount"`
}
