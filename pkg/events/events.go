package events

type AuthorizationRequested struct {
	TransactionID string  `json:"transaction_id"`
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
}

type AuthorizationSucceeded struct {
	TransactionID string `json:"transaction_id"`
}

type AuthorizationFailed struct {
	TransactionID string `json:"transaction_id"`
	Reason        string `json:"reason"`
}

type LedgerUpdateSucceeded struct {
	TransactionID string `json:"transaction_id"`
}

type LedgerUpdateFailed struct {
	TransactionID string `json:"transaction_id"`
	Reason        string `json:"reason"`
}