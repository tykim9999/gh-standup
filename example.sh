#!/bin/bash

# Example usage script for gh-standup extension

echo "Building gh-standup extension..."
go build -o gh-standup ./cmd/standup

echo "Testing help command..."
./gh-standup --help

echo ""
echo "Example commands you can run:"
echo ""
echo "1. Generate standup for yesterday:"
echo "   gh standup"
echo ""
echo "2. Generate standup for specific user:"
echo "   gh standup --user octocat"
echo ""
echo "3. Generate standup for specific repo:"
echo "   gh standup --repo microsoft/vscode"
echo ""
echo "4. Generate standup for last 3 days:"
echo "   gh standup --days 3"
echo ""
echo "5. Use different AI model:"
echo "   gh standup --model anthropic/claude-3.5-sonnet"
echo ""
echo "6. Combine options:"
echo "   gh standup --user octocat --repo owner/repo --days 2"
