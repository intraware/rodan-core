package handlers

type errorResponse struct {
	Error string `json:"error" example:"Something went wrong"`
}

type challengeItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type challengeDetail struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Desc       string   `json:"desc"`
	Category   int8     `json:"category"`
	Difficulty int8     `json:"difficulty"`
	Points     int      `json:"points"`
	Solved     bool     `json:"solved"`
	Links      []string `json:"links,omitempty"`
}

type ContainerResponse struct {
	Flag  string   `json:"flag"`
	Ports []int    `json:"ports"`
	Links []string `json:"links"`
}

type submitFlagRequest struct {
	Flag string `json:"flag" binding:"required"`
}

type submitFlagResponse struct {
	Correct bool   `json:"correct"`
	Message string `json:"message"`
}
