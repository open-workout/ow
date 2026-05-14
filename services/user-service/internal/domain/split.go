package domain

type SplitElement struct {
	Muscles []string `json:"muscles"`
	Title   string   `json:"title"`
}

type Split struct {
	Elements []SplitElement `json:"elements"`
}
