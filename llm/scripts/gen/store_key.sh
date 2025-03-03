#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <provider> <secret_key>"
    echo "Supported providers: chatgpt,claude,deepseek,gemini,minimax"
    exit 1
fi

provider=$1
secret_key=$2

case "$provider" in
    "chatgpt")
        account_name="my_openai"
        service_name="chatgpt"
        ;;
    "claude")
        account_name="my_anthropic"
        service_name="claude"
        ;;
    "deepseek")
        account_name="my_deepseek"
        service_name="deepseek"
        ;;
    "gemini")
        account_name="my_google"
        service_name="gemini"
        ;;
    "minimax")
        account_name="my_hailuoai"
        service_name="minimax"
        ;;
    *)
        echo "Error: Invalid provider. Supported providers: chatgpt,claude,deepseek,gemini,minimax"
        exit 1
        ;;
esac

security add-generic-password -U -a "$account_name" -s "$service_name" -w "$secret_key"

echo "Successfully stored API key for $provider"
