#!/usr/bin/env python3
"""
Script to create the note management tool in Letta.
"""
from letta_client import Letta
import os

# Initialize client
# client = Letta(api_key=os.environ["CAMERON_API_KEY"])
client = Letta(api_key=os.environ["LETTA_API_KEY"])

# Read the tool source code
with open("note_tool.py", "r") as f:
    source_code = f.read()

# Create or update the tool
tool = client.tools.upsert(
    source_code=source_code,
)

print(f"✓ Created tool: {tool.name}")
print(f"  ID: {tool.id}")
print(f"  Description: {tool.description}")
print("\nTool is ready to use!")
print("\nExample usage:")
print('  note(command="attach", path="/self/tasks", content="TODO: Review code")')
print('  note(command="view", path="/self/tasks")')
print('  note(command="list", query="/self/")')
