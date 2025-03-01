#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <provider>"
    echo "Supported providers: chatgpt,claude,deepseek,minimax"
    exit 1
fi

provider=$1

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
    "minimax")
        account_name="my_hailuoai"
        service_name="minimax"
        ;;
    *)
        echo "Error: Invalid provider. Supported providers: chatgpt,claude,deepseek,minimax"
        exit 1
        ;;
esac

if security find-generic-password -a "$account_name" -s "$service_name" >/dev/null 2>&1; then
    echo "API key exists for $provider"
    exit 0
else
    echo "No API key found for $provider"
    exit 1
fi
