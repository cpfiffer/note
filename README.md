# Letta Notes

The note tool enables **context mounting** for Letta agents - a form of progressive disclosure that lets agents manage memory blocks as if they were a file system.

Instead of loading everything into context at once, agents can work with an arbitrary number of text files, attaching only what's needed for the current task. This allows for active, distributed memory management - agents can peek, view, and organize information across a structured hierarchy rather than cramming everything into a single context window.

Think of it as giving your agent a personal file system for structured memory. Notes persist across sessions, can be organized into folders, and are mounted/unmounted from context on demand.

## Installation

```bash
pip install letta-client
```

## Usage

Register the tool with your Letta project, then attach to your agent:

```python
note(command="attach", path="/tasks", content="TODO: Review code")
note(command="view", path="/tasks")
note(command="list")
```

## Commands

- `create` - Create a new note (not attached to context)
- `view` - Read note contents
- `attach` - Load note into agent context (creates if missing)
- `detach` - Remove from context (keeps in storage)
- `insert` - Insert content at a specific line
- `append` - Add content to end of note
- `replace` - Find and replace text
- `rename` - Move/rename note
- `copy` - Duplicate note
- `delete` - Permanently remove note
- `list` - List notes by path prefix
- `search` - Search notes by label or content
- `attached` - Show currently attached notes

## Practical Usage Patterns

**Attach/Detach for Context Management:**
```
note attach /reference/api-docs    # Load into context when needed
note detach /reference/api-docs    # Remove when done to free context space
```

**Bulk Operations:**
```
note attach /folder/*              # Attach all notes in a folder
note detach /folder/*              # Detach all
```

**Progressive Disclosure:**
- Keep detailed references in notes, only attach when relevant
- Detach after use to keep context window lean
- Use `note attached` to see what's currently loaded

**Organizational Patterns:**
```
/projects/          # Project-specific notes
/references/        # Documentation, API specs
/learning/          # Curriculum, lesson plans
/shared/            # Cross-agent shared content
```

## Tips

1. **Notes persist across sessions** - anything stored survives conversation resets
2. **Folders are also notes** - `/projects` and `/projects/task1` can both have content
3. **Search before creating** - use `note search <query>` to find existing content
4. **Prefer append over replace** for logs/journals to avoid data loss
5. **Use descriptive paths** - they're your only way to find things later
