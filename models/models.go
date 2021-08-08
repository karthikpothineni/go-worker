package models

// BillingEvent - holds billing event information
type BillingEvent struct {
	UserID     int    `json:"user_id"`
	ProductID  int    `json:"product_id"`
	CallID     string `json:"call_id"`
	AnswerTime string `json:"answer_time"`
	HangupTime string `json:"hangup_time"`
}

// BalanceResponse - holds balance api response
type BalanceResponse struct {
	CallID             string `json:"call_id"`
	TotalConsumedUnits int    `json:"total_consumed_units"`
	ChargeAmount       string `json:"charge_amount"`
}
