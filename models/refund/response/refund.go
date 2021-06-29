package response

type AccountRefundResponse struct {
	CaptureAmount int    `json:"capture_amount"`
	RefundAmount  int    `json:"refund_amount"`
	Currency      string `json:"currency"`
}
