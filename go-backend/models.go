package main

type Profile struct {
	ID              int64   `json:"id"`
	Name            *string `json:"name"`
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	CoverPhotoURL   *string `json:"coverPhotoUrl"`
	ProfilePhotoURL *string `json:"profilePhotoUrl"`
	Email           *string `json:"email"`
	Location        *string `json:"location"`
	GithubURL       *string `json:"githubUrl"`
	LinkedinURL     *string `json:"linkedinUrl"`
}

type Project struct {
	ID           int64    `json:"id"`
	Title        *string  `json:"title"`
	Description  *string  `json:"description"`
	Technologies []string `json:"technologies"`
	ImageURL     *string  `json:"imageUrl"`
	LiveURL      *string  `json:"liveUrl"`
	RepoURL      *string  `json:"repoUrl"`
	DisplayOrder int      `json:"displayOrder"`
}

type Education struct {
	ID           int64   `json:"id"`
	Institution  *string `json:"institution"`
	Degree       *string `json:"degree"`
	FieldOfStudy *string `json:"fieldOfStudy"`
	StartYear    *string `json:"startYear"`
	EndYear      *string `json:"endYear"`
	Description  *string `json:"description"`
	DisplayOrder int     `json:"displayOrder"`
}

type Experience struct {
	CompanyLogoPath *string  `json:"companyLogoPath"`
	ID              int64    `json:"id"`
	Company         *string  `json:"company"`
	Role            *string  `json:"role"`
	Location        *string  `json:"location"`
	StartDate       *string  `json:"startDate"`
	EndDate         *string  `json:"endDate"`
	Highlights      []string `json:"highlights"`
	DisplayOrder    int      `json:"displayOrder"`
}

type Portfolio struct {
	EducationPhotoPath *string      `json:"educationPhotoPath"`
	Profile            Profile      `json:"profile"`
	Projects           []Project    `json:"projects"`
	Education          []Education  `json:"education"`
	Experience         []Experience `json:"experience"`
}
