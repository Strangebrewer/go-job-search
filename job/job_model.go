package job

type CreateJobRequest struct {
	RecruiterID       string `json:"recruiterId"`
	JobTitle          string `json:"jobTitle"`
	WorkFrom          string `json:"workFrom"`
	DateApplied       string `json:"dateApplied"`
	CompanyName       string `json:"companyName"`
	CompanyAddress    string `json:"companyAddress"`
	CompanyCity       string `json:"companyCity"`
	CompanyState      string `json:"companyState"`
	PointOfContact    string `json:"pointOfContact"`
	PocTitle          string `json:"pocTitle"`
	PrimaryLink       string `json:"primaryLink"`
	PrimaryLinkText   string `json:"primaryLinkText"`
	SecondaryLink     string `json:"secondaryLink"`
	SecondaryLinkText string `json:"secondaryLinkText"`
	Status            string `json:"status"`
}

type UpdateJobRequest struct {
	RecruiterID       string   `json:"recruiterId"`
	JobTitle          string   `json:"jobTitle"`
	WorkFrom          string   `json:"workFrom"`
	DateApplied       string   `json:"dateApplied"`
	CompanyName       string   `json:"companyName"`
	CompanyAddress    string   `json:"companyAddress"`
	CompanyCity       string   `json:"companyCity"`
	CompanyState      string   `json:"companyState"`
	PointOfContact    string   `json:"pointOfContact"`
	PocTitle          string   `json:"pocTitle"`
	Interviews        []string `json:"interviews"`
	Comments          []string `json:"comments"`
	Status            string   `json:"status"`
	Archived          bool     `json:"archived"`
	PrimaryLink       string   `json:"primaryLink"`
	PrimaryLinkText   string   `json:"primaryLinkText"`
	SecondaryLink     string   `json:"secondaryLink"`
	SecondaryLinkText string   `json:"secondaryLinkText"`
}

type JobFilter struct {
	Company         string
	RecruiterID     string
	Status          string
	WorkFrom        string
	DateMin         string
	DateMax         string
	IncludeArchived bool
	IncludeDeclined bool
	SortBy          string
	SortDir         string
}
