#!/bin/bash
#
# Install the Letta Note tool
# Usage: curl -sSL https://raw.githubusercontent.com/cpfiffer/note/main/install.sh | bash
#

set -e

echo "Letta Note Tool Installer"
echo "========================="
echo ""

# Get API key
API_KEY="${LETTA_API_KEY:-}"

if [ -z "$API_KEY" ]; then
  read -p "Enter your Letta API key: " API_KEY
  echo ""
fi

if [ -z "$API_KEY" ]; then
  echo "Error: API key required"
  exit 1
fi

echo "Checking for existing 'note' tool..."
EXISTING=$(curl -sS "https://api.letta.com/v1/tools/?name=note" \
  -H "Authorization: Bearer $API_KEY")

if echo "$EXISTING" | grep -q '"id"'; then
  EXISTING_ID=$(echo "$EXISTING" | python3 -c 'import json,sys; d=json.loads(sys.stdin.read()); print(d[0]["id"] if d else "")')
  if [ -n "$EXISTING_ID" ]; then
    echo ""
    echo "Warning: A tool named 'note' already exists (ID: $EXISTING_ID)"
    read -p "Overwrite it? [y/N] " CONFIRM
    if [ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ]; then
      echo "Aborted."
      exit 0
    fi
    echo ""
  fi
fi

echo "Fetching note_tool.py..."
SOURCE_CODE=$(curl -sSL https://raw.githubusercontent.com/cpfiffer/note/main/note_tool.py)

if [ -z "$SOURCE_CODE" ]; then
  echo "Error: Failed to fetch note_tool.py"
  exit 1
fi

echo "Installing tool..."

# Escape for JSON
SOURCE_CODE_JSON=$(echo "$SOURCE_CODE" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read()))')

# Upsert the tool
RESPONSE=$(curl -sS https://api.letta.com/v1/tools/ \
  -X PUT \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d "{\"source_code\": $SOURCE_CODE_JSON}")

# Check for errors
if echo "$RESPONSE" | grep -q '"error"'; then
  echo "Error: $RESPONSE"
  exit 1
fi

TOOL_ID=$(echo "$RESPONSE" | python3 -c 'import json,sys; print(json.loads(sys.stdin.read()).get("id", "unknown"))')

echo ""
echo "✓ Note tool installed!"
echo ""
echo "Tool ID: $TOOL_ID"
echo ""
echo "Next: Attach to your agent via ADE or SDK:"
echo "  client.agents.tools.attach(agent_id=AGENT_ID, tool_id=\"$TOOL_ID\")"
