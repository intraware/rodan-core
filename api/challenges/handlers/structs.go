package handlers

type challengeItem struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

type challengeDetail struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Author     string `json:"author"`
	Desc       string `json:"desc"`
	Category   int8   `json:"category"`
	Difficulty int8   `json:"difficulty"`
	Points     int    `json:"points"`
	Solved     bool   `json:"solved"`
}

type submitFlagRequest struct {
	Flag string `json:"flag" binding:"required"`
}

type submitFlagResponse struct {
	Correct bool   `json:"correct"`
	Message string `json:"message"`
}

type challengeConfigResponse struct {
	ID       uint     `json:"id"`
	Links    []string `json:"links,omitempty"`
	Ports    []int    `json:"ports,omitempty"`
	TimeLeft int64    `json:"timeleft,omitempty"`
	IsStatic bool     `json:"is_static"`
}
