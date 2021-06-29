package request

type CreditCard struct {
	CardNumber      int    `json:"card_number"`
	CardExpiryMonth int    `json:"card_expiry_month"`
	CardExpiryYear  int    `json:"card_expiry_year"`
	CVV             int    `json:"cvv"`
	Amount          int    `json:"amount"`
	Currency        string `json:"currency"`
}
