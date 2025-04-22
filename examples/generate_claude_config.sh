#!/bin/bash

# Script to generate a claude_config.json file with correct paths
# This should be run from the examples directory

set -e  # Exit on error

# Get the parent directory of the examples folder
PARENT_DIR="$(cd .. && pwd)"

# Define the output file
CONFIG_FILE="claude_config.json"

# Check if we're in the examples directory
if [[ "$(basename $(pwd))" != "examples" ]]; then
  echo "Error: This script must be run from the examples directory"
  exit 1
fi

# Check if binaries exist
if [[ ! -f "$PARENT_DIR/bin/piglatin-mcp" ]]; then
  echo "Warning: piglatin-mcp binary not found. Make sure to build it first with 'make build'"
fi

if [[ ! -f "$PARENT_DIR/bin/temporal-mcp" ]]; then
  echo "Warning: temporal-mcp binary not found. Make sure to build it first with 'make build'"
fi

# Generate the JSON configuration file
cat > "$CONFIG_FILE" << EOF
{
  "mcpServers": {
    "piglatin-mcp": {
      "command": "$PARENT_DIR/bin/piglatin-mcp",
      "args": [],
      "env": {}
    },
    "temporal-mcp": {
      "command": "$PARENT_DIR/bin/temporal-mcp",
      "args": ["--config", "$PARENT_DIR/config.yml"],
      "env": {}
    }
  }
}
EOF

echo "Generated $CONFIG_FILE with correct paths"

# Add file to .gitignore if it's not already there
GITIGNORE_FILE="$PARENT_DIR/.gitignore"

if [[ -f "$GITIGNORE_FILE" ]]; then
  if ! grep -q "examples/$CONFIG_FILE" "$GITIGNORE_FILE"; then
    echo "Adding $CONFIG_FILE to .gitignore"
    echo "examples/$CONFIG_FILE" >> "$GITIGNORE_FILE"
  else
    echo "$CONFIG_FILE is already in .gitignore"
  fi
else
  echo "Creating .gitignore and adding $CONFIG_FILE"
  echo "examples/$CONFIG_FILE" > "$GITIGNORE_FILE"
fi

# Instructions for the user
echo ""
echo "To use this configuration with Claude:"
echo "1. Copy this file to Claude's configuration directory:"
echo "   cp $CONFIG_FILE ~/Library/Application\\ Support/Claude/claude_desktop_config.json"
echo "2. Restart Claude if it's already running"
echo ""
echo "Alternatively, you can reference this file in your Claude installation settings."
echo "See the README.md for more information."

# Make the script executable
chmod +x "$0"
