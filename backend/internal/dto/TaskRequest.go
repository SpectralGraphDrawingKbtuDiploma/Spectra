package dto

type JobRequest struct {
	ID      string  `json:"id"`
	Content *string `json:"content"`
}

type JobResponse struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result *string `json:"result"`
	Error  *string `json:"error"`
}
