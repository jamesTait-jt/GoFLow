#!/bin/sh
set -e

# Check if the argument is provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <handlers-directory>"
    exit 1
fi

ls -la
cd handlers
echo "in handlers!!!"
ls -la
pwd

# Compile all Go files in the specified directory
for file in "$1"/*.go; do
    # Get the base name without the extension for output
    output_file="$1/compiled/$(basename "$file" .go).so"
    go build -buildmode=plugin -o "$output_file" "$file"
done