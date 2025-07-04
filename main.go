package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

const extensionName = "standup"

var rootCmd = &cobra.Command{
	Use:   extensionName,
	Short: "Generate AI-powered standup reports",
	Long:  "A GitHub CLI extension that generates standup reports using GitHub Models and GitHub API data",
	RunE:  runStandup,
}

var (
	flagDays  int
	flagModel string
	flagRepo  string
	flagUser  string
)

func init() {
	rootCmd.Flags().IntVarP(&flagDays, "days", "d", 1, "Number of days to look back for activity")
	rootCmd.Flags().StringVarP(&flagModel, "model", "m", "openai/gpt-4o", "GitHub Models model to use")
	rootCmd.Flags().StringVarP(&flagRepo, "repo", "r", "", "Repository to generate standup for (owner/repo)")
	rootCmd.Flags().StringVarP(&flagUser, "user", "u", "", "User to generate standup for (defaults to authenticated user)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runStandup(cmd *cobra.Command, args []string) error {
	if flagUser == "" {
		fmt.Print("Getting authenticated GitHub user... ")
		user, err := getCurrentUser()
		if err != nil {
			fmt.Println("Failed")
			return fmt.Errorf("failed to get current user: %w", err)
		}
		flagUser = user
		fmt.Printf("Found user: %s\n", flagUser)
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -flagDays)

	fmt.Printf("Analyzing GitHub activity for %s (%s to %s)\n", 
		flagUser, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	fmt.Print("Collecting GitHub activity data...\n")
	activities, err := collectGitHubActivity(flagUser, flagRepo, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to collect GitHub activity: %w", err)
	}

	if len(activities) == 0 {
		fmt.Println("No GitHub activity found for the specified period.")
		return nil
	}

	fmt.Printf("Found %d activities\n", len(activities))

	commits := 0
	prs := 0
	issues := 0
	reviews := 0
	for _, activity := range activities {
		switch activity.Type {
		case "commit":
			commits++
		case "pull_request":
			prs++
		case "issue":
			issues++
		case "review":
			reviews++
		}
	}
	fmt.Printf("   %d commits, %d pull requests, %d issues, %d reviews\n", commits, prs, issues, reviews)

	fmt.Printf("Generating standup report using %s...\n", flagModel)
	report, err := generateStandupReport(activities, flagModel)
	if err != nil {
		return fmt.Errorf("failed to generate standup report: %w", err)
	}

	fmt.Println("Report generated successfully!")
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("STANDUP REPORT")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println(report)

	return nil
}

func getCurrentUser() (string, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return "", err
	}

	var user struct {
		Login string `json:"login"`
	}

	err = client.Get("user", &user)
	if err != nil {
		return "", err
	}

	return user.Login, nil
}
