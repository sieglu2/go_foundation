#!/bin/bash

# Function to delete a credential
delete_credential() {
    local account=$1
    local service=$2
    local provider=$3

    echo "Deleting key for $provider..."
    if security delete-generic-password -a "$account" -s "$service" 2>/dev/null; then
        echo "✓ Successfully deleted API key for $provider"
    else
        echo "✗ Failed to delete API key for $provider (key might not exist)"
    fi
}

# Delete chatgpt credentials
delete_credential "my_openai" "chatgpt" "chatgpt"

# Delete claude credentials
delete_credential "my_anthropic" "claude" "claude"

# Delete deepseek credentials
delete_credential "my_deepseek" "deepseek" "deepseek"

# Delete minimax credentials
delete_credential "my_hailuoai" "minimax" "minimax"

echo "Operation completed"
