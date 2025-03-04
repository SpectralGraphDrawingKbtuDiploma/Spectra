package internal

type GraphDTO struct {
	ID    *string `json:"id"`
	Edges []int   `json:"edges"`
}

type TaskStatus struct {
	ID     string    `json:"id"`
	Status string    `json:"status"`
	Result []float64 `json:"result"`
}
