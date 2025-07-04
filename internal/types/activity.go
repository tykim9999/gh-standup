package types

import "time"

type GitHubActivity struct {
	Type        string    `json:"type"`
	Repository  string    `json:"repository"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"created_at"`
}
