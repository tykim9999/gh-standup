package github

import (
	"fmt"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/gh-standup/internal/types"
)

type Client struct {
	client *api.RESTClient
}

func NewClient() (*Client, error) {
	fmt.Print("  Connecting to GitHub API... ")
	client, err := api.DefaultRESTClient()
	if err != nil {
		fmt.Println("Failed")
		return nil, err
	}
	fmt.Println("Done")
	
	return &Client{client: client}, nil
}

func (c *Client) GetCurrentUser() (string, error) {
	var user struct {
		Login string `json:"login"`
	}

	err := c.client.Get("user", &user)
	if err != nil {
		return "", err
	}

	return user.Login, nil
}

// CollectActivity gathers activity data from GitHub API
func (c *Client) CollectActivity(username, repo string, startDate, endDate time.Time) ([]types.GitHubActivity, error) {
	var activities []types.GitHubActivity

	// Collect commits (may be slow or fail)
	fmt.Print("  🔍 Searching for commits... ")
	commits, err := c.getCommits(username, repo, startDate, endDate)
	if err != nil {
		fmt.Printf("⚠️  Skipped (search may be restricted)\n")
	} else {
		fmt.Printf("✅ Found %d commits\n", len(commits))
		activities = append(activities, commits...)
	}

	// Collect pull requests
	fmt.Print("  🔍 Searching for pull requests... ")
	prs, err := c.getPullRequests(username, repo, startDate, endDate)
	if err != nil {
		fmt.Println("❌")
		return nil, fmt.Errorf("failed to get pull requests: %w", err)
	}
	fmt.Printf("✅ Found %d pull requests\n", len(prs))
	activities = append(activities, prs...)

	// Collect issues
	fmt.Print("  🔍 Searching for issues... ")
	issues, err := c.getIssues(username, repo, startDate, endDate)
	if err != nil {
		fmt.Println("❌")
		return nil, fmt.Errorf("failed to get issues: %w", err)
	}
	fmt.Printf("✅ Found %d issues\n", len(issues))
	activities = append(activities, issues...)

	// Collect reviews
	fmt.Print("  🔍 Searching for code reviews... ")
	reviews, err := c.getReviews(username, startDate, endDate)
	if err != nil {
		fmt.Println("❌")
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}
	fmt.Printf("✅ Found %d reviews\n", len(reviews))
	activities = append(activities, reviews...)

	return activities, nil
}

func (c *Client) getCommits(username, repo string, startDate, endDate time.Time) ([]types.GitHubActivity, error) {
	var activities []types.GitHubActivity

	// Search for commits by author
	query := fmt.Sprintf("author:%s committer-date:%s..%s", 
		username, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	if repo != "" {
		query += fmt.Sprintf(" repo:%s", repo)
	}

	var searchResult struct {
		Items []struct {
			SHA        string `json:"sha"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
			Commit struct {
				Message string `json:"message"`
				Author  struct {
					Date time.Time `json:"date"`
				} `json:"author"`
			} `json:"commit"`
			HTMLURL string `json:"html_url"`
		} `json:"items"`
	}

	// The commits search API often fails or is slow due to GitHub's rate limiting
	// If it fails, we'll return an error so the caller can handle it gracefully
	escapedQuery := strings.ReplaceAll(query, " ", "%20")
	err := c.client.Get(fmt.Sprintf("search/commits?q=%s&sort=committer-date&order=desc", escapedQuery), &searchResult)
	if err != nil {
		// Return error so caller knows commits search failed
		return activities, fmt.Errorf("commits search failed (this is common due to GitHub API restrictions): %w", err)
	}

	for _, item := range searchResult.Items {
		activities = append(activities, types.GitHubActivity{
			Type:        "commit",
			Repository:  item.Repository.FullName,
			Title:       strings.Split(item.Commit.Message, "\n")[0],
			Description: item.Commit.Message,
			URL:         item.HTMLURL,
			CreatedAt:   item.Commit.Author.Date,
		})
	}

	return activities, nil
}

func (c *Client) getPullRequests(username, repo string, startDate, endDate time.Time) ([]types.GitHubActivity, error) {
	var activities []types.GitHubActivity

	// Search for pull requests
	query := fmt.Sprintf("author:%s created:%s..%s", 
		username, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	if repo != "" {
		query += fmt.Sprintf(" repo:%s", repo)
	}

	var searchResult struct {
		Items []struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
			Body   string `json:"body"`
			State  string `json:"state"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
			HTMLURL   string    `json:"html_url"`
			CreatedAt time.Time `json:"created_at"`
		} `json:"items"`
	}

	escapedQuery := strings.ReplaceAll(query, " ", "%20")
	err := c.client.Get(fmt.Sprintf("search/issues?q=%s+type:pr&sort=created&order=desc", escapedQuery), &searchResult)
	if err != nil {
		return activities, nil
	}

	for _, item := range searchResult.Items {
		activities = append(activities, types.GitHubActivity{
			Type:        "pull_request",
			Repository:  item.Repository.FullName,
			Title:       fmt.Sprintf("PR #%d: %s", item.Number, item.Title),
			Description: item.Body,
			URL:         item.HTMLURL,
			CreatedAt:   item.CreatedAt,
		})
	}

	return activities, nil
}

func (c *Client) getIssues(username, repo string, startDate, endDate time.Time) ([]types.GitHubActivity, error) {
	var activities []types.GitHubActivity

	// Search for issues created by user
	query := fmt.Sprintf("author:%s created:%s..%s", 
		username, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	if repo != "" {
		query += fmt.Sprintf(" repo:%s", repo)
	}

	var searchResult struct {
		Items []struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
			Body   string `json:"body"`
			State  string `json:"state"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
			HTMLURL   string    `json:"html_url"`
			CreatedAt time.Time `json:"created_at"`
		} `json:"items"`
	}

	escapedQuery := strings.ReplaceAll(query, " ", "%20")
	err := c.client.Get(fmt.Sprintf("search/issues?q=%s+type:issue&sort=created&order=desc", escapedQuery), &searchResult)
	if err != nil {
		return activities, nil
	}

	for _, item := range searchResult.Items {
		activities = append(activities, types.GitHubActivity{
			Type:        "issue",
			Repository:  item.Repository.FullName,
			Title:       fmt.Sprintf("Issue #%d: %s", item.Number, item.Title),
			Description: item.Body,
			URL:         item.HTMLURL,
			CreatedAt:   item.CreatedAt,
		})
	}

	return activities, nil
}

func (c *Client) getReviews(username string, startDate, endDate time.Time) ([]types.GitHubActivity, error) {
	var activities []types.GitHubActivity

	// Search for pull requests reviewed by user
	query := fmt.Sprintf("reviewed-by:%s created:%s..%s", 
		username, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	var searchResult struct {
		Items []struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
			HTMLURL   string    `json:"html_url"`
			CreatedAt time.Time `json:"created_at"`
		} `json:"items"`
	}

	escapedQuery := strings.ReplaceAll(query, " ", "%20")
	err := c.client.Get(fmt.Sprintf("search/issues?q=%s+type:pr&sort=created&order=desc", escapedQuery), &searchResult)
	if err != nil {
		return activities, nil
	}

	for _, item := range searchResult.Items {
		activities = append(activities, types.GitHubActivity{
			Type:        "review",
			Repository:  item.Repository.FullName,
			Title:       fmt.Sprintf("Reviewed PR #%d: %s", item.Number, item.Title),
			Description: fmt.Sprintf("Reviewed pull request: %s", item.Title),
			URL:         item.HTMLURL,
			CreatedAt:   item.CreatedAt,
		})
	}

	return activities, nil
}
