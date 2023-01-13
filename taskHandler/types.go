package main

type OrderRequestBody struct {
	Sources []string `json:"sources"`
}

type OrderStatusBody struct {
	NewStatus string `json:"newStatus"`
}

type Archive struct {
	ID          string      `json:"id"`
	Content     []FileGroup `json:"content"`
	Meta        string      `json:"meta"`
	TimeCreated string      `json:"timeCreated"`
	TimeUpdated string      `json:"timeUpdated"`
	Status      string      `json:"status"`
	Sources     []string    `json:"sources"`
}

type FileGroup struct {
	SourceID string   `json:"sourceID"`
	Files    []string `json:"files"`
}
