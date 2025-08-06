package handlers

type errorResponse struct {
	Error string `json:"error" example:"Something went wrong"`
}

type successResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}