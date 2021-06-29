package response

type AccountRefundResponse struct {
	CaptureAmount int    `json:"capture_amount,omitempty"`
	RefundAmount  *int   `json:"refund_amount,omitempty"`
	Currency      string `json:"currency,omitempty"`
}
