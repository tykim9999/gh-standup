# gh-standup

A GitHub CLI extension that generates AI-powered standup reports using GitHub activity data.

## Installation

### Prerequisites

- [GitHub CLI](https://cli.github.com/) installed and authenticated
- Go 1.21+ (for building from source)
- GitHub token with appropriate permissions (automatically used by GitHub CLI)

### Install from Source

```bash
# Clone the repository
git clone https://github.com/your-username/gh-standup.git
cd gh-standup

# Build and install
make build
gh extension install .
```

### Install from GitHub

```bash
gh extension install your-username/gh-standup
```

## Usage

### Basic Usage

Generate a standup report for yesterday's activity:

```bash
gh standup
```

### Advanced Options

```bash
# Generate report for specific user
gh standup --user octocat

# Generate report for specific repository
gh standup --repo owner/repo

# Look back multiple days
gh standup --days 3

# Use a different AI model
gh standup --model anthropic/claude-3.5-sonnet

# Combine options
gh standup --user octocat --repo microsoft/vscode --days 2 --model openai/gpt-4o
```

## Authentication with Organizations

To ensure the extension can access your organization repositories and activities:

```bash
# Authenticate with GitHub CLI (if not already done)
gh auth login

# Authenticate with your organizations
gh auth refresh -h github.com -s read:org

# To authenticate with a GitHub Enterprise instance
gh auth login --hostname your-enterprise-instance.com
gh auth refresh -h your-enterprise-instance.com -s read:org
```

This ensures your GitHub token includes the necessary organization access permissions.