package models

type Account struct {
	ID      int
	Person  string
	Number  string
	Balance int
	Type    string `json:"type"`
}

type AccountCreationResponse struct {
	Number string `json:"number"`
	Type   string `json:"type"`
}

type AccountUpdateRequest struct {
	Person string
	Number string  `json:"number"`
	Sum    float64 `json:"sum"`
}

type AccountTransferRequest struct {
	Person string
	From   string  `json:"from"`
	To     string  `json:"to"`
	Sum    float64 `json:"sum"`
}
