package api

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type WithdrawRequest struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}

type OrderListResponse struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int    `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at"`
}

type UserBalanceResponse struct {
	Current   int `json:"current"`
	Withdrawn int `json:"withdrawn"`
}

type WithdrawalListResponse struct {
	Order       string `json:"order"`
	Sum         int    `json:"sum"`
	ProcessedAt string `json:"processed_at"`
}
