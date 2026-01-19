# Note Tool for Letta Agents

External storage tool for Letta agents to manage persistent notes with attach/detach semantics.

## Installation

Register the tool with your Letta agent:

```bash
python create_note_tool.py
```

## Commands

| Command | Description |
|---------|-------------|
| `create <path> <content>` | Create new note (not attached) |
| `view <path>` | Read note contents |
| `attach <path> [content]` | Load into context (supports `/folder/*`) |
| `detach <path>` | Remove from context (supports `/folder/*`) |
| `insert <path> <content> [line]` | Insert before line (0-indexed) |
| `append <path> <content>` | Add content to end |
| `replace <path> <old_str> <new_str>` | Find/replace with diff |
| `rename <path> <new_path>` | Move/rename note |
| `copy <path> <new_path>` | Duplicate note |
| `delete <path>` | Permanently remove |
| `list [query]` | List notes (prefix filter, `*` for all) |
| `search <query> [label\|content]` | Search by label or content |
| `attached` | Show notes in context |

## Features

- **Agent-scoped storage** - Each agent's notes are isolated
- **Attach/detach semantics** - Load notes into context or remove them
- **Path-based organization** - Organize with paths like `/projects/webapp`
- **Bulk operations** - Support for wildcards like `attach /folder/*`
- **Auto-maintained directory** - A `/note_directory` block shows all notes

## License

MIT
