package main

type OrderRequestBody struct {
	Sources []string `json:"sources"`
}

type OrderStatusBody struct {
	NewStatus string `json:"newStatus"`
}
