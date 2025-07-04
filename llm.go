package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type GitHubModelsRequest struct {
	Messages      []Message `json:"messages"`
	Model         string    `json:"model"`
	Temperature   float64   `json:"temperature"`
	TopP          float64   `json:"top_p"`
	Stream        bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GitHubModelsResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func generateStandupReport(activities []GitHubActivity, model string) (string, error) {
	fmt.Print("  Checking GitHub token... ")
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("Failed")
		return "", fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}
	fmt.Println("Done")

	fmt.Print("  Formatting activity data for AI... ")
	activitySummary := formatActivitiesForLLM(activities)
	fmt.Println("Done")

	systemPrompt := `You are an AI assistant helping to generate professional standup reports based on GitHub activity. 

Your task is to create a concise, well-structured standup report that summarizes the developer's work from the previous day(s). The report should be written in first person and include:

1. **Yesterday's Accomplishments**: What was completed/worked on
2. **Today's Plans**: Logical next steps based on the activity (be realistic)
3. **Blockers/Challenges**: Any potential issues or dependencies mentioned

Guidelines:
- Keep it professional but conversational
- Focus on meaningful work rather than trivial commits
- Group related activities together
- Highlight significant contributions like new features, bug fixes, or reviews
- Be concise but informative
- Use bullet points for clarity
- Avoid technical jargon that non-developers wouldn't understand

Format the output as a clean, readable report without any markdown headers.`

	userPrompt := fmt.Sprintf("Based on the following GitHub activity, generate a standup report:\n\n%s", activitySummary)

	request := GitHubModelsRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Model:       model,
		Temperature: 0.7,
		TopP:        1.0,
		Stream:      false,
	}

	fmt.Printf("  Calling GitHub Models API (%s)... ", model)
	response, err := callGitHubModels(request, token)
	if err != nil {
		fmt.Println("Failed")
		return "", err
	}
	fmt.Println("Done")

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response generated from the model")
	}

	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}

func formatActivitiesForLLM(activities []GitHubActivity) string {
	if len(activities) == 0 {
		return "No GitHub activity found for the specified period."
	}

	var builder strings.Builder
	
	commits := make([]GitHubActivity, 0)
	prs := make([]GitHubActivity, 0)
	issues := make([]GitHubActivity, 0)
	reviews := make([]GitHubActivity, 0)

	for _, activity := range activities {
		switch activity.Type {
		case "commit":
			commits = append(commits, activity)
		case "pull_request":
			prs = append(prs, activity)
		case "issue":
			issues = append(issues, activity)
		case "review":
			reviews = append(reviews, activity)
		}
	}

	if len(commits) > 0 {
		builder.WriteString("COMMITS:\n")
		for _, commit := range commits {
			builder.WriteString(fmt.Sprintf("- [%s] %s\n", commit.Repository, commit.Title))
			if commit.Description != commit.Title {
				lines := strings.Split(commit.Description, "\n")
				if len(lines) > 1 && lines[1] != "" {
					builder.WriteString(fmt.Sprintf("  Description: %s\n", strings.TrimSpace(lines[1])))
				}
			}
		}
		builder.WriteString("\n")
	}

	if len(prs) > 0 {
		builder.WriteString("PULL REQUESTS:\n")
		for _, pr := range prs {
			builder.WriteString(fmt.Sprintf("- [%s] %s\n", pr.Repository, pr.Title))
			if pr.Description != "" && len(pr.Description) < 200 {
				builder.WriteString(fmt.Sprintf("  Description: %s\n", strings.TrimSpace(pr.Description)))
			}
		}
		builder.WriteString("\n")
	}

	if len(issues) > 0 {
		builder.WriteString("ISSUES:\n")
		for _, issue := range issues {
			builder.WriteString(fmt.Sprintf("- [%s] %s\n", issue.Repository, issue.Title))
			if issue.Description != "" && len(issue.Description) < 200 {
				builder.WriteString(fmt.Sprintf("  Description: %s\n", strings.TrimSpace(issue.Description)))
			}
		}
		builder.WriteString("\n")
	}

	if len(reviews) > 0 {
		builder.WriteString("CODE REVIEWS:\n")
		for _, review := range reviews {
			builder.WriteString(fmt.Sprintf("- [%s] %s\n", review.Repository, review.Title))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

func callGitHubModels(request GitHubModelsRequest, token string) (*GitHubModelsResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://models.github.ai/inference/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response GitHubModelsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}
