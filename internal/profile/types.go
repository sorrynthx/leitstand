package profile

// Links holds the developer's common public URLs.
type Links struct {
	Web      string `json:"web"`
	Git      string `json:"git"`
	LinkedIn string `json:"linkedin"`
}

// LangProfile holds localized developer story and career philosophy.
type LangProfile struct {
	Name        string `json:"name"`
	StoryTitle  string `json:"story_title"`
	StoryP1     string `json:"story_p1"`
	StoryP2     string `json:"story_p2"`
	StoryP3     string `json:"story_p3"`
	CareerTitle string `json:"career_title"`
	CareerDesc  string `json:"career_desc"`
	TechTitle   string `json:"tech_title"`
	TechP1      string `json:"tech_p1"`
	TechP2      string `json:"tech_p2"`
	TechP3      string `json:"tech_p3"`
}

// Payload represents the root structure of the encrypted developer profile.
type Payload struct {
	Links    Links                  `json:"links"`
	Profiles map[string]LangProfile `json:"profiles"`
}
