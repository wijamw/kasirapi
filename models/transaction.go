package models

type Transaction struct {
	ID          int                 `json:"id"`
	TotalAmount int                 `json:"total_amount"`
	Details     []TransactionDetail `json:"details"`
}

type TransactionDetail struct {
	ID           	int    `json:"id"`
	TransactionID	int    `json:"transaction_id"`
	CategoryID		int    `json:"category_id"`
	CategoryName	string `json:"category_name"`
	Quantity      	int    `json:"quantity"`
}

type CheckoutRequest struct {
	Items []CheckoutItem `json:"items"`
}

type CheckoutItem struct {
	CategoryID int `json:"category_id"`
	Quantity  int `json:"quantity"`
}

type DailyCategoryReport struct {
    TotalTransaction   int    `json:"total_transaction"`
    MostPickedCategory string `json:"most_picked_category"`
    TotalPicked        int    `json:"total_picked"`
}