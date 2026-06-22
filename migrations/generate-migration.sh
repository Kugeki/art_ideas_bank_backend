#!/bin/sh

if [ -z "$1" ]; then
    echo "Please provide a name for this migration."
    exit 1
fi


command -v goose || {
    echo >&2 "Goose command not found. Have you installed goose?";
    echo >&2 "https://github.com/pressly/goose?tab=readme-ov-file#install";
    exit 1;
}

goose create "$1" sql
