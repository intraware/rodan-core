package handlers

type errorResponse struct {
	Error string `json:"error" example:"Something went wrong"`
}