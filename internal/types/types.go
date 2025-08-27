package types

type ErrorResponse struct {
	Error string `json:"error" example:"Something went wrong"`
}

type SuccessResponse struct {
	Message string `json:"message" example:"Something went right"`
}
