package main

import (
	"strings"
	"testing"
	"time"
)

func TestFormatActivitiesForLLM(t *testing.T) {
	activities := []GitHubActivity{
		{
			Type:        "commit",
			Repository:  "test/repo",
			Title:       "Fix bug in user authentication",
			Description: "Fix bug in user authentication\n\nThis commit resolves the issue where users couldn't log in",
			URL:         "https://github.com/test/repo/commit/abc123",
			CreatedAt:   time.Now(),
		},
		{
			Type:        "pull_request",
			Repository:  "test/repo",
			Title:       "PR #123: Add new feature",
			Description: "This PR adds a new feature for better user experience",
			URL:         "https://github.com/test/repo/pull/123",
			CreatedAt:   time.Now(),
		},
	}

	result := formatActivitiesForLLM(activities)

	if result == "" {
		t.Error("Expected non-empty result")
	}

	if !strings.Contains(result, "COMMITS:") {
		t.Error("Expected result to contain COMMITS section")
	}

	if !strings.Contains(result, "PULL REQUESTS:") {
		t.Error("Expected result to contain PULL REQUESTS section")
	}

	if !strings.Contains(result, "Fix bug in user authentication") {
		t.Error("Expected result to contain commit title")
	}

	if !strings.Contains(result, "PR #123: Add new feature") {
		t.Error("Expected result to contain PR title")
	}
}

func TestFormatActivitiesForLLMEmpty(t *testing.T) {
	activities := []GitHubActivity{}
	result := formatActivitiesForLLM(activities)

	expected := "No GitHub activity found for the specified period."
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
