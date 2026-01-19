#!/usr/bin/env python3
"""
Script to create the vault tool in Letta.
"""
from letta_client import Letta
import os

# Initialize client
client = Letta(api_key=os.environ["LETTA_API_KEY"])

# Read the tool source code
with open("vault_tool.py", "r") as f:
    source_code = f.read()

# Create or update the tool
tool = client.tools.upsert(
    source_code=source_code,
)

print(f"Created tool: {tool.name}")
print(f"  ID: {tool.id}")
print(f"  Description: {tool.description}")
print("\nTool is ready to use!")
print("\nExample usage:")
print('  vault(command="attach", path="/tasks", content="TODO: Review code")')
print('  vault(command="view", path="/tasks")')
print('  vault(command="list")')
