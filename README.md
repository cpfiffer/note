# Letta Note

A [custom tool](https://docs.letta.com/guides/agents/custom-tools/) that gives [Letta agents](https://docs.letta.com/guides/agents/) a file system for their [memory blocks](https://docs.letta.com/guides/agents/memory/).

Agents can store an arbitrary number of notes, organized into folders, and selectively attach them to their [context window](https://docs.letta.com/guides/agents/context-engineering/) when needed. This enables progressive disclosure: instead of loading everything at once, agents mount only what's relevant to the current task.

Notes persist in your Letta server and can be viewed, edited, searched, and reorganized on demand.

## Installation

**WARNING**: Custom tools in Letta are not versioned. If you overwrite an existing tool called `note` with this one, you may not be able to restore your original tool.

```bash
curl -sSL https://raw.githubusercontent.com/cpfiffer/note/main/install.sh | bash
```

The installer will prompt for your API key (or use `LETTA_API_KEY` if set) and attach the tool to your agent via ADE or SDK.

It will check to see if a tool of the same name exists and confirm whether you want to overwrite it.

## Usage

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

## Note Directory

A `/note_directory` block is automatically maintained and attached to your agent. It displays all notes in a tree view with the first 80 characters of each note's first line as a preview.

Write descriptive first lines - they serve as the summary in your directory listing.

## Tips

1. **Notes are blocks** - each note is a memory block with a path-like label, scoped to your agent
2. **Attach = load into context** - attaching a note adds it to your agent's active memory blocks
3. **Detach ≠ delete** - detaching removes from context but the note still exists in storage
4. **Folders are also notes** - `/projects` and `/projects/task1` can both have content
5. **Use descriptive paths** - the label is your only way to find notes later

## Command Permissions

By default, all commands are enabled. You can restrict which commands are available by setting the `ENABLED_COMMANDS` environment variable on your Letta server:

```bash
# Enable all commands (default)
ENABLED_COMMANDS="all"

# Disable delete (safer for production)
ENABLED_COMMANDS="create,view,attach,detach,insert,append,replace,rename,copy,list,search,attached"

# Read-only mode
ENABLED_COMMANDS="view,list,search,attached"
```

Disabled commands will return an error message listing which commands are enabled.
