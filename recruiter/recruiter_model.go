package recruiter

type CreateRecruiterRequest struct {
	Name    string `json:"name"`
	Company string `json:"company"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Rating  int32  `json:"rating"`
}

type UpdateRecruiterRequest struct {
	Name     string   `json:"name"`
	Company  string   `json:"company"`
	Phone    string   `json:"phone"`
	Email    string   `json:"email"`
	Rating   int32    `json:"rating"`
	Comments []string `json:"comments"`
	Archived bool     `json:"archived"`
}
