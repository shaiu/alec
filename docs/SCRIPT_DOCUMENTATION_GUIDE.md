# Script Documentation Guide

## Overview
Alec automatically extracts and displays script descriptions and previews in the TUI. This guide shows you how to document your scripts for optimal display.

## Shell Scripts (Bash, Sh, Zsh)

### Basic Header Comments
The simplest way to document a shell script is with header comments:

```bash
#!/bin/bash
# This script performs daily database backups
# It runs every night at midnight via cron
# and uploads backups to S3

# Your script code here
backup_database() {
    ...
}
```

**Result in Alec:**
> Description: "This script performs daily database backups It runs every night at midnight via cron and uploads backups to S3"

### Using Description Markers
For more explicit documentation, use special markers:

```bash
#!/bin/bash
# Description: Automated backup script with S3 upload
# Creates compressed archives and uploads to AWS S3

# Your script code here
```

**Supported Markers:**
- `# Description: ...`
- `# @desc ...`
- `# @description ...`
- `# Summary: ...`
- `# @summary ...`

### Best Practices

✅ **Good:**
```bash
#!/bin/bash
# Description: Deploys the application to production
# Pulls latest code, runs migrations, and restarts services

set -e
...
```

❌ **Avoid:**
```bash
#!/bin/bash
# Script
# TODO: Add description later

# Random comments in the middle
...
```

## Python Scripts

### Module-Level Docstrings (Preferred)
Python scripts should use module-level docstrings:

```python
#!/usr/bin/env python3
"""
Data processing pipeline for customer analytics.

This script extracts data from the database, performs
transformations, and generates daily reports.
"""

import pandas as pd
...
```

**Result in Alec:**
> Description: "Data processing pipeline for customer analytics. This script extracts data from the database, performs transformations, and generates daily reports."

### Single-Line Docstrings
For simple scripts:

```python
#!/usr/bin/env python3
"""Quick utility to sync files between servers."""

import os
...
```

### Alternative: Header Comments
If you prefer comments over docstrings:

```python
#!/usr/bin/env python3
# Description: Sync files between development and staging
# Runs hourly via systemd timer

import os
...
```

### Best Practices

✅ **Good:**
```python
#!/usr/bin/env python3
"""
User management CLI tool.

Provides commands for creating, updating, and deleting
user accounts with proper validation and logging.
"""
```

✅ **Also Good:**
```python
#!/usr/bin/env python3
# @desc Database migration helper
# Applies pending migrations and validates schema

import psycopg2
...
```

❌ **Avoid:**
```python
#!/usr/bin/env python3

# Some code here
def main():
    """Function docstring doesn't count as module description"""
    ...
```

## Tips for Great Script Documentation

### 1. Keep it Concise
Descriptions are automatically truncated to 300 characters. Make your first sentence count!

```bash
# Description: Monitors server health and sends alerts
# Checks CPU, memory, disk, and network metrics every 5 minutes
```

### 2. Start with the Purpose
Begin with what the script does, not how it does it:

✅ `# Description: Generates monthly sales reports`
❌ `# Description: This script uses pandas and matplotlib to...`

### 3. Include Key Information
Mention important aspects like:
- **What** the script does
- **When** it runs (if automated)
- **Where** it operates (which systems/services)

```bash
#!/bin/bash
# Description: Nightly log rotation for web servers
# Compresses and archives logs older than 7 days
# Runs on all production web nodes at 3 AM
```

### 4. Use Empty Lines Freely
Empty lines within header comments are fine:

```bash
#!/bin/bash
# Main deployment script for staging environment

# Handles code deployment, database migrations,
# and service restarts with rollback capability

set -e
...
```

### 5. Marker Position
Markers can appear anywhere in the header comment block:

```bash
#!/bin/bash
# Backup Script
# @desc Creates encrypted backups of customer data
# Version: 2.0
```

## What Gets Displayed in Alec

When you select a script in Alec, you'll see:

```
🐚 backup-script

📁 Location: /home/user/scripts/backup-script.sh
🔧 Type: shell
⚙️  Interpreter: /bin/bash
📅 Modified: 2025-10-05 14:30:00
📏 Size: 2.1 KB
📊 Lines: 87

📝 Description:
Automated backup script with S3 upload. Creates compressed
archives and uploads to AWS S3 with encryption.

──────────────────────────────────────────────────
📄 Script Preview (showing 50 of 87 lines)

#!/bin/bash
# Description: Automated backup script with S3 upload
# Creates compressed archives and uploads to AWS S3
...

──────────────────────────────────────────────────
⚡ Press Enter to execute this script
```

## Script Preview Behavior

### Short Scripts (≤30 lines)
Full script content is shown:

```
📄 Full Script

[entire script content]
```

### Long Scripts (>30 lines)
First 50 lines are shown with indicator:

```
📄 Script Preview (showing 50 of 120 lines)

[first 50 lines]

... (script continues)
```

## Examples by Use Case

### Deployment Script
```bash
#!/bin/bash
# Description: Production deployment automation
# Pulls code, builds assets, runs migrations, restarts services
# Includes automatic rollback on failure
```

### Monitoring Script
```python
#!/usr/bin/env python3
"""
System health monitoring and alerting.

Checks CPU, memory, disk, and network every 5 minutes.
Sends Slack alerts when thresholds are exceeded.
"""
```

### Data Processing
```python
#!/usr/bin/env python3
# @desc ETL pipeline for customer analytics
# Extracts from PostgreSQL, transforms with pandas, loads to warehouse
# Runs daily at 2 AM via Airflow
```

### Utility Script
```bash
#!/bin/bash
# Quick utility to find and clean old Docker images
# Removes images older than 30 days to free disk space
```

### Maintenance Script
```bash
#!/bin/bash
# Description: Database maintenance and optimization
# Vacuums tables, rebuilds indexes, updates statistics
# Safe to run on production with minimal downtime
```

## Testing Your Documentation

To see how your script will appear in Alec:

1. Add your script to a configured directory
2. Refresh Alec (press `r`)
3. Navigate to your script
4. View the description and preview in the main content area

Or run the parser directly:

```go
config := parser.DefaultParseConfig()
metadata, err := parser.ParseScript("/path/to/script.sh", "shell", config)
fmt.Println(metadata.Description)
```

## Common Mistakes to Avoid

### ❌ Description Too Far Down
```bash
#!/bin/bash

# Some code here
...

# Description: This won't be found
```
*Parser only looks at header comments (first 20 lines)*

### ❌ Code Before Description
```python
#!/usr/bin/env python3
import os  # <-- Code here

"""This docstring won't be found"""
```
*Docstring must appear before any code*

### ❌ Only Function Docstrings
```python
#!/usr/bin/env python3

def main():
    """This won't be used as module description"""
    ...
```
*Need module-level docstring, not function docstring*

### ❌ Multiline Without Marker
```bash
#!/bin/bash
This is not a comment
So it won't be parsed
```
*Must use # for each line*

## Summary

**For Shell Scripts:**
- Add header comments after shebang
- Use `# Description:` or other markers for clarity
- Keep first line focused and concise

**For Python Scripts:**
- Use module-level docstrings (triple quotes)
- Place immediately after shebang, before imports
- Or use `# Description:` header comments

**General Tips:**
- First 20 lines are scanned for documentation
- First 300 characters of description are shown
- Empty lines in headers are OK
- Multiple consecutive comment lines are joined

---

Happy documenting! Your future self (and teammates) will thank you. 🎉
