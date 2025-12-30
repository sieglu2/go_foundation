#!/bin/bash

if [ $# -eq 0 ]; then
    echo "Usage: $0 <suffix1> [suffix2] [suffix3] ..."
    echo "Example: $0 .js .py xyz.zip"
    exit 1
fi

patterns=""
for suffix in "$@"; do
    if [ -z "$patterns" ]; then
        patterns="-name \"*$suffix\""
    else
        patterns="$patterns -o -name \"*$suffix\""
    fi
done

eval "find . -type f \\( $patterns \\) -print0" | \
while IFS= read -r -d '' file; do
    echo "// $file"
    cat "$file"
    echo
done
