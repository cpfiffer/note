from typing import Literal, Optional
import re


def vault(
    command: str,
    path: Optional[str] = None,
    content: Optional[str] = None,
    old_str: Optional[str] = None,
    new_str: Optional[str] = None,
    new_path: Optional[str] = None,
    insert_line: Optional[int] = None,
    query: Optional[str] = None,
    search_type: str = "label",
) -> str:
    """
    Manage items in your vault. All items are automatically scoped to your agent.
    
    Commands:
      create <path> <content>             - create new item (not attached)
      view <path>                         - read item contents
      attach <path> [content]             - load into context (supports /folder/*)
      detach <path>                       - remove from context (supports /folder/*)
      insert <path> <content> [line]      - insert before line (0-indexed) or append
      append <path> <content>             - add content to end of item
      replace <path> <old_str> <new_str>  - find/replace, shows diff
      rename <path> <new_path>            - move/rename item
      copy <path> <new_path>              - duplicate item
      delete <path>                       - permanently remove
      list [query]                        - list items (prefix filter, * for all)
      search <query> [label|content]      - grep items by label or content
      attached                            - show items currently in context
    
    Args:
        command: The operation to perform
        path: Path to the item (e.g., /projects/webapp, /todo)
        content: Content to insert or initial content when creating
        old_str: Text to find (for replace)
        new_str: Text to replace with (for replace)
        new_path: Destination path (for rename/copy)
        insert_line: Line number to insert before (0-indexed, omit to append)
        query: Search query (for list/search)
        search_type: Search by "label" or "content"
    
    Returns:
        str: Result of the operation
    """
    import os
    
    agent_id = os.environ.get("LETTA_AGENT_ID")
    owner_tag = f"owner:{agent_id}"
    
    # Check enabled commands ("all" or "*" enables everything)
    all_commands = ["create", "view", "attach", "detach", "insert", "append", "replace", "rename", "copy", "delete", "list", "search", "attached"]
    enabled_env = os.environ.get("ENABLED_COMMANDS", "create,view,attach,detach,insert,append,replace,rename,copy,list,search,attached")
    enabled = all_commands if enabled_env in ("all", "*") else enabled_env.split(",")
    if command not in enabled:
        return f"Error: '{command}' is disabled. Enabled: {enabled}"
    
    # Pattern to filter out legacy UUID paths
    uuid_pattern = re.compile(r'/\[?agent-[a-f0-9-]+\]?/')
    
    # Parameter validation
    path_required = ["create", "view", "attach", "detach", "insert", "append", "replace", "rename", "copy", "delete"]
    if command in path_required and not path:
        return f"Error: '{command}' requires path parameter"
    
    if command == "replace" and (not old_str or new_str is None):
        return "Error: 'replace' requires old_str and new_str parameters"
    
    if command in ["create", "insert", "append"] and not content:
        return f"Error: '{command}' requires content parameter"
    
    if command in ["rename", "copy"] and not new_path:
        return f"Error: '{command}' requires new_path parameter"
    
    if command == "search" and not query:
        return "Error: 'search' requires query parameter"
    
    # Track if directory needs updating
    update_directory = False
    result = None
    
    try:
        if command == "create":
            # Check for existing note with same path
            existing = list(client.blocks.list(label=path, tags=[owner_tag]).items)
            if existing:
                return f"Error: Item already exists: {path}"
            
            client.blocks.create(
                label=path,
                value=content,
                tags=[owner_tag]
            )
            update_directory = True
            result = f"Created: {path}"
        
        elif command == "view":
            blocks = list(client.blocks.list(label=path, tags=[owner_tag]).items)
            if not blocks:
                return f"Item not found: {path}"
            return blocks[0].value
        
        elif command == "attach":
            # Get currently attached block IDs to avoid duplicate attach errors
            agent = client.agents.retrieve(agent_id=agent_id)
            attached_ids = {b.id for b in agent.memory.blocks}
            
            # Handle bulk wildcard: /folder/*
            if path.endswith("/*"):
                prefix = path[:-1]  # "/folder/*" → "/folder/"
                all_blocks = list(client.blocks.list(tags=[owner_tag]).items)
                blocks = [b for b in all_blocks if b.label and b.label.startswith(prefix)
                          and not uuid_pattern.search(b.label)]
                if not blocks:
                    return f"No items matching: {path}"
                
                to_attach = [b for b in blocks if b.id not in attached_ids]
                skipped = len(blocks) - len(to_attach)
                
                for block in to_attach:
                    client.agents.blocks.attach(agent_id=agent_id, block_id=block.id)
                
                msg = f"Attached {len(to_attach)} items matching {path}"
                if skipped:
                    msg += f" ({skipped} already attached)"
                return msg
            
            # Single note attach
            existing = list(client.blocks.list(label=path, tags=[owner_tag]).items)
            if existing:
                block_id = existing[0].id
                if block_id in attached_ids:
                    return f"Already attached: {path}"
            else:
                new_block = client.blocks.create(
                    label=path,
                    value=content or "",
                    tags=[owner_tag]
                )
                block_id = new_block.id
                update_directory = True  # New note created
            
            client.agents.blocks.attach(agent_id=agent_id, block_id=block_id)
            result = f"Attached: {path}"
        
        elif command == "detach":
            # Handle bulk wildcard: /folder/*
            if path.endswith("/*"):
                prefix = path[:-1]
                # Get currently attached block IDs
                agent = client.agents.retrieve(agent_id=agent_id)
                attached_ids = {b.id for b in agent.memory.blocks}
                
                all_blocks = list(client.blocks.list(tags=[owner_tag]).items)
                blocks = [b for b in all_blocks if b.label and b.label.startswith(prefix)
                          and not uuid_pattern.search(b.label)
                          and b.id in attached_ids]  # Only detach if actually attached
                if not blocks:
                    return f"No attached items matching: {path}"
                for block in blocks:
                    client.agents.blocks.detach(agent_id=agent_id, block_id=block.id)
                return f"Detached {len(blocks)} items matching {path}"
            
            # Single note detach
            blocks = list(client.blocks.list(label=path, tags=[owner_tag]).items)
            if not blocks:
                return f"Item not found: {path}"
            
            client.agents.blocks.detach(agent_id=agent_id, block_id=blocks[0].id)
            return f"Detached: {path}"
        
        elif command == "insert":
            blocks = list(client.blocks.list(label=path, tags=[owner_tag]).items)
            if not blocks:
                return f"Item not found: {path}. Use 'attach' first."
            
            block = blocks[0]
            lines = block.value.split("\n") if block.value else []
            
            if insert_line is not None:
                lines.insert(insert_line, content)
                line_info = f"line {insert_line}"
            else:
                lines.append(content)
                line_info = "end"
            
            client.blocks.update(block_id=block.id, value="\n".join(lines))
            
            preview = content[:80] + "..." if len(content) > 80 else content
            return f"Inserted at {line_info} in {path}:\n  + {preview}"
        
        elif command == "append":
            blocks = list(client.blocks.list(label=path, tags=[owner_tag]).items)
            if not blocks:
                return f"Item not found: {path}. Use 'attach' first."
            
            block = blocks[0]
            if block.value:
                new_value = block.value + "\n" + content
            else:
                new_value = content
            
            client.blocks.update(block_id=block.id, value=new_value)
            
            preview = content[:80] + "..." if len(content) > 80 else content
            return f"Appended to {path}:\n  + {preview}"
        
        elif command == "rename":
            # Check source exists
            blocks = list(client.blocks.list(label=path, tags=[owner_tag]).items)
            if not blocks:
                return f"Item not found: {path}"
            
            # Check destination doesn't exist
            dest_blocks = list(client.blocks.list(label=new_path, tags=[owner_tag]).items)
            if dest_blocks:
                return f"Error: Destination already exists: {new_path}"
            
            # Update the label
            block = blocks[0]
            client.blocks.update(block_id=block.id, label=new_path)
            update_directory = True
            result = f"Renamed: {path} → {new_path}"
        
        elif command == "copy":
            # Check source exists
            blocks = list(client.blocks.list(label=path, tags=[owner_tag]).items)
            if not blocks:
                return f"Item not found: {path}"
            
            # Check destination doesn't exist
            dest_blocks = list(client.blocks.list(label=new_path, tags=[owner_tag]).items)
            if dest_blocks:
                return f"Error: Destination already exists: {new_path}"
            
            # Create copy (not attached)
            source = blocks[0]
            client.blocks.create(
                label=new_path,
                value=source.value,
                tags=[owner_tag]
            )
            update_directory = True
            result = f"Copied: {path} → {new_path}"
        
        elif command == "replace":
            blocks = list(client.blocks.list(label=path, tags=[owner_tag]).items)
            if not blocks:
                return f"Item not found: {path}"
            
            block = blocks[0]
            if old_str not in block.value:
                return f"Error: old_str not found in note. Exact match required."
            
            new_value = block.value.replace(old_str, new_str, 1)
            client.blocks.update(block_id=block.id, value=new_value)
            
            return f"Replaced in {path}:\n  - {old_str}\n  + {new_str}"
        
        elif command == "delete":
            blocks = list(client.blocks.list(label=path, tags=[owner_tag]).items)
            if not blocks:
                return f"Item not found: {path}"
            
            client.blocks.delete(block_id=blocks[0].id)
            update_directory = True
            result = f"Deleted: {path}"
        
        elif command == "list":
            all_blocks = list(client.blocks.list(tags=[owner_tag]).items)
            
            # Filter to path-like labels, exclude legacy UUID paths
            blocks = [b for b in all_blocks 
                      if b.label and b.label.startswith("/")
                      and not uuid_pattern.search(b.label)]
            
            # Apply prefix filter if query provided
            if query and query != "*":
                blocks = [b for b in blocks if b.label.startswith(query)]
            
            if not blocks:
                return "No items found" if not query or query == "*" else f"No items matching: {query}"
            
            # Deduplicate and sort
            labels = sorted(set(b.label for b in blocks))
            return "\n".join(labels)
        
        elif command == "search":
            all_blocks = list(client.blocks.list(tags=[owner_tag]).items)
            
            # Filter out legacy UUID paths
            all_blocks = [b for b in all_blocks 
                          if b.label and b.label.startswith("/")
                          and not uuid_pattern.search(b.label)]
            
            if search_type == "label":
                blocks = [b for b in all_blocks if query in b.label]
            else:
                blocks = [b for b in all_blocks if b.value and query in b.value]
            
            if not blocks:
                return f"No items matching: {query}"
            
            results = []
            for b in blocks:
                preview = b.value[:100].replace("\n", " ") if b.value else ""
                if len(b.value or "") > 100:
                    preview += "..."
                results.append(f"{b.label}: {preview}")
            
            return "\n".join(results)
        
        elif command == "attached":
            agent = client.agents.retrieve(agent_id=agent_id)
            note_blocks = [b for b in agent.memory.blocks 
                          if b.label and b.label.startswith("/")
                          and not uuid_pattern.search(b.label)]
            
            if not note_blocks:
                return "No items currently attached"
            
            return "\n".join(sorted(b.label for b in note_blocks))
        
        else:
            return f"Error: Unknown command '{command}'"
        
        # Update vault_directory if needed
        if update_directory:
            dir_label = "/vault_directory"
            # Get all items
            all_blocks = list(client.blocks.list(tags=[owner_tag]).items)
            items = [b for b in all_blocks 
                     if b.label and b.label.startswith("/") 
                     and b.label != dir_label
                     and not uuid_pattern.search(b.label)]
            
            # Header for the directory
            header = "External storage. Attach to load into context, detach when done.\nFolders are also items (e.g., /projects and /projects/task1 can both have content).\nCommands: view, attach, detach, insert, append, replace, rename, copy, delete, list, search\nBulk: attach /folder/*, detach /folder/*"
            
            if items:
                # Group items by folder
                folders = {}
                for b in sorted(items, key=lambda x: x.label):
                    parts = b.label.rsplit("/", 1)
                    if len(parts) == 2:
                        folder, name = parts[0] + "/", parts[1]
                    else:
                        folder, name = "/", b.label[1:]  # Root level
                    if folder not in folders:
                        folders[folder] = []
                    first_line = (b.value or "").split("\n")[0][:40]
                    if len((b.value or "").split("\n")[0]) > 40:
                        first_line += "..."
                    folders[folder].append((name, first_line))
                
                # Build tree view
                lines = []
                for folder in sorted(folders.keys()):
                    lines.append(folder)
                    folder_items = folders[folder]
                    max_name_len = max(len(name) for name, _ in folder_items)
                    for name, summary in folder_items:
                        lines.append(f"  {name.ljust(max_name_len)} | {summary}")
                
                dir_content = header + "\n\n" + "\n".join(lines)
            else:
                dir_content = header + "\n\n(no items)"
            
            # Find or create directory block
            dir_blocks = list(client.blocks.list(label=dir_label, tags=[owner_tag]).items)
            if dir_blocks:
                client.blocks.update(block_id=dir_blocks[0].id, value=dir_content)
            else:
                # Create and attach directory block
                dir_block = client.blocks.create(
                    label=dir_label,
                    value=dir_content,
                    tags=[owner_tag]
                )
                client.agents.blocks.attach(agent_id=agent_id, block_id=dir_block.id)
        
        if result:
            return result
        
    except Exception as e:
        return f"Error executing '{command}': {str(e)}"
