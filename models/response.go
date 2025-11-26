package models

type SuccessResponse[T any] struct {
	Status string `json:"status" example:"ok"`
	Data   T      `json:"data"`
}

type FailureResponse struct {
	Status string `json:"status" example:"error"`
	Error  string `json:"error" example:"description"`
}
