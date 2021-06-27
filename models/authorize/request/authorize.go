package request

type CreditCard struct {
	CardNumber      int    `json:"card_number,omitempty"`
	CardExpiryMonth int    `json:"card_expiry_month,omitempty"`
	CardExpiryYear  int    `json:"card_expiry_year,omitempty"`
	CVV             int    `json:"cvv,omitempty"`
	Amount          int    `json:"amount,omitempty"`
	Currency        string `json:"currency,omitempty"`
}
