# Vault Tool for Letta Agents

External storage tool for Letta agents to manage persistent vault items with attach/detach semantics.

> **Note**: This branch uses tags-based ownership which requires an updated Letta client with block tags support.

## Installation

Register the tool with your Letta agent:

```bash
python create_vault_tool.py
```

## Commands

| Command | Description |
|---------|-------------|
| `create <path> <content>` | Create new item (not attached) |
| `view <path>` | Read item contents |
| `attach <path> [content]` | Load into context (supports `/folder/*`) |
| `detach <path>` | Remove from context (supports `/folder/*`) |
| `insert <path> <content> [line]` | Insert before line (0-indexed) |
| `append <path> <content>` | Add content to end |
| `replace <path> <old_str> <new_str>` | Find/replace with diff |
| `rename <path> <new_path>` | Move/rename item |
| `copy <path> <new_path>` | Duplicate item |
| `delete <path>` | Permanently remove |
| `list [query]` | List items (prefix filter, `*` for all) |
| `search <query> [label\|content]` | Search by label or content |
| `attached` | Show items in context |

## Features

- **Agent-scoped storage** - Each agent's items are isolated via tags
- **Attach/detach semantics** - Load items into context or remove them
- **Path-based organization** - Organize with paths like `/projects/webapp`
- **Bulk operations** - Support for wildcards like `attach /folder/*`
- **Auto-maintained directory** - A `/vault_directory` block shows all items

## License

MIT
