package dto

type TaskRequest struct {
	ID    string `json:"id"`
	Edges []int  `json:"edges"`
}

type TaskResponse struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result *string `json:"result"`
	Error  *string `json:"error"`
}
