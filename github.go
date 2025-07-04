package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

type GitHubActivity struct {
	Type        string    `json:"type"`
	Repository  string    `json:"repository"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"created_at"`
}

func collectGitHubActivity(username, repo string, startDate, endDate time.Time) ([]GitHubActivity, error) {
	fmt.Print("  Connecting to GitHub API... ")
	client, err := api.DefaultRESTClient()
	if err != nil {
		fmt.Println("Failed")
		return nil, err
	}
	fmt.Println("Done")

	var activities []GitHubActivity

	fmt.Print("  Searching for commits... ")
	commits, err := getCommits(client, username, repo, startDate, endDate)
	if err != nil {
		fmt.Println("Failed")
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}
	fmt.Printf("Done. Found %d commits\n", len(commits))
	activities = append(activities, commits...)

	// Collect pull requests
	fmt.Print("  Searching for pull requests... ")
	prs, err := getPullRequests(client, username, repo, startDate, endDate)
	if err != nil {
		fmt.Println("Failed")
		return nil, fmt.Errorf("failed to get pull requests: %w", err)
	}
	fmt.Printf("Done. Found %d pull requests\n", len(prs))
	activities = append(activities, prs...)

	fmt.Print("  Searching for issues... ")
	issues, err := getIssues(client, username, repo, startDate, endDate)
	if err != nil {
		fmt.Println("Failed")
		return nil, fmt.Errorf("failed to get issues: %w", err)
	}
	fmt.Printf("Done. Found %d issues\n", len(issues))
	activities = append(activities, issues...)

	fmt.Print("  Searching for code reviews... ")
	reviews, err := getReviews(client, username, startDate, endDate)
	if err != nil {
		fmt.Println("Failed")
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}
	fmt.Printf("Done. Found %d reviews\n", len(reviews))
	activities = append(activities, reviews...)

	return activities, nil
}

func getCommits(client *api.RESTClient, username, repo string, startDate, endDate time.Time) ([]GitHubActivity, error) {
	var activities []GitHubActivity

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

	err := client.Get(fmt.Sprintf("search/commits?q=%s&sort=committer-date&order=desc", query), &searchResult)
	if err != nil {
		// Commits search might fail due to permissions, continue silently
		return activities, nil
	}

	for _, item := range searchResult.Items {
		activities = append(activities, GitHubActivity{
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

func getPullRequests(client *api.RESTClient, username, repo string, startDate, endDate time.Time) ([]GitHubActivity, error) {
	var activities []GitHubActivity

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

	err := client.Get(fmt.Sprintf("search/issues?q=%s+type:pr&sort=created&order=desc", query), &searchResult)
	if err != nil {
		return activities, nil
	}

	for _, item := range searchResult.Items {
		activities = append(activities, GitHubActivity{
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

func getIssues(client *api.RESTClient, username, repo string, startDate, endDate time.Time) ([]GitHubActivity, error) {
	var activities []GitHubActivity

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

	err := client.Get(fmt.Sprintf("search/issues?q=%s+type:issue&sort=created&order=desc", query), &searchResult)
	if err != nil {
		return activities, nil
	}

	for _, item := range searchResult.Items {
		activities = append(activities, GitHubActivity{
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

func getReviews(client *api.RESTClient, username string, startDate, endDate time.Time) ([]GitHubActivity, error) {
	var activities []GitHubActivity

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

	err := client.Get(fmt.Sprintf("search/issues?q=%s+type:pr&sort=created&order=desc", query), &searchResult)
	if err != nil {
		return activities, nil
	}

	for _, item := range searchResult.Items {
		activities = append(activities, GitHubActivity{
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
