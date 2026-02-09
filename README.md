# qBittorrent File Matcher

A desktop application that matches torrent files in qBittorrent with existing files on disk by file size, then renames files in qBittorrent to match. This solves the problem of seeding torrents when you have the files but with different names.

## Features

- **File Matching by Size** - Automatically finds files on disk that match torrent files by size
- **Smart Renaming** - Renames files in qBittorrent to point to your existing files
- **GUI & CLI** - Use the graphical interface or command line
- **Recheck Support** - Trigger torrent recheck after renaming to verify file integrity
- **Skip Unmatched** - Option to set priority to 0 for files without matches
- **Extension Filtering** - Optionally require matching file extensions

## Installation

### Download

Download the latest release from the [Releases](https://github.com/username/qbittorrent-file-matcher/releases) page.

### Build from Source

Requirements:

- Go 1.21+
- Node.js 18+
- [Wails 3](https://v3.wails.io/)

```bash
# Clone the repository
git clone https://github.com/username/qbittorrent-file-matcher.git
cd qbittorrent-file-matcher

# Build GUI application
wails3 build

# Build CLI-only application (no WebView dependency)
wails3 task windows:build:cli
```

## Usage

### GUI Application

1. Launch the application
2. Enter your qBittorrent WebUI URL, username, and password
3. Select a torrent from the list
4. Enter the directory path where your files are located
5. Click "Scan" to find matches
6. Review matches and click "Apply Renames"
7. Optionally click "Recheck Torrent" to verify file integrity

### CLI Application

```bash
# Basic usage
qbittorrent-file-matcher-cli match \
  --url http://localhost:8080 \
  --username admin \
  --password secret \
  --hash <torrent-hash> \
  --path /path/to/files

# Using environment variables (recommended for password)
export QBT_URL=http://localhost:8080
export QBT_USERNAME=admin
export QBT_PASSWORD=secret
qbittorrent-file-matcher-cli match --hash <torrent-hash> --path /path/to/files

# Additional options
qbittorrent-file-matcher-cli match \
  --url http://localhost:8080 \
  --hash <torrent-hash> \
  --path /path/to/files \
  --auto              # Auto-select first match (no prompts)
  --dry-run           # Preview changes without applying
  --skip-unmatched    # Set priority 0 for unmatched files
  --recheck           # Trigger recheck after renaming
  --no-same-ext       # Allow matching files with different extensions
```

### CLI Options

| Flag                    | Description                                           |
| ----------------------- | ----------------------------------------------------- |
| `--url <url>`           | qBittorrent WebUI URL (e.g., <http://localhost:8080>) |
| `--hash <hash>`         | Torrent hash to match                                 |
| `--path <path>`         | Directory path to scan for files                      |
| `-u, --username <user>` | qBittorrent username                                  |
| `-p, --password <pass>` | qBittorrent password                                  |
| `--same-ext`            | Only match files with same extension (default)        |
| `--no-same-ext`         | Allow matching files with different extensions        |
| `--skip-unmatched`      | Set priority to 0 for unmatched files                 |
| `-r, --recheck`         | Trigger torrent recheck after applying renames        |
| `--dry-run`             | Show what would be done without making changes        |
| `-a, --auto`            | Auto-select first match (no interactive prompts)      |

### Environment Variables

| Variable       | Description                                      |
| -------------- | ------------------------------------------------ |
| `QBT_URL`      | Default qBittorrent WebUI URL                    |
| `QBT_USERNAME` | Default username                                 |
| `QBT_PASSWORD` | Default password (more secure than command line) |

## How It Works

1. **Scan Directory** - Recursively scans the specified directory and indexes all files by size
2. **Match Files** - For each torrent file, finds disk files with matching size
3. **Auto-Match** - If only one file matches a size, it's automatically selected
4. **Manual Selection** - If multiple files match, you can choose which one to use
5. **Rename in qBittorrent** - Updates file paths in qBittorrent to point to your files
6. **Recheck** - Optionally triggers a hash recheck to verify file integrity

## Requirements

- qBittorrent with WebUI enabled (Tools > Options > Web UI)
- Files must match by size (the content should be identical)

## Tech Stack

- **Backend**: Go with [Wails 3](https://v3.wails.io/)
- **Frontend**: React + TypeScript + Vite + [shadcn/ui](https://ui.shadcn.com/) + Tailwind CSS
- **qBittorrent API**: [autobrr/go-qbittorrent](https://github.com/autobrr/go-qbittorrent)

## Development

```bash
# Run in development mode (hot reload)
wails3 dev

# Run tests
go test ./...

# Run linter
golangci-lint run ./...

# Frontend linting
cd frontend && npm run lint
```

## License

MIT License
