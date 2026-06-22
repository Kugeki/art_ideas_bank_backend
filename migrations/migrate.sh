#!/bin/sh

if [ -z "${POSTGRES_URL}" ]; then
    echo "Please, provide a POSTGRES_URL environment variable."
    exit 1
fi


command -v goose || {
    echo >&2 "Goose command not found. Have you installed goose?";
    echo >&2 "https://github.com/pressly/goose?tab=readme-ov-file#install";
    exit 1;
}


retries=5
code=1
until [ ${retries} -eq 0 ]; do
    goose postgres "${POSTGRES_URL}" up
    code=$?
    if [ ${code} -eq 0 ]; then
      break
    fi
    retries="$((retries-1))"
    sleep 2
done

if [ ${code} -ne 0 ]; then
  exit 1
fi

exit 0
