# JSIShell

A modern interactive shell interpreter written in Go with standard Unix commands, inline autocompletion, and cross-platform support.

## Features

- **Standard Unix Commands**: `cd`, `pwd`, `ls`, `cp`, `mv`, `rm`, `mkdir`, `search`
- **Glob Expansion**: `ls *.go` expands wildcards automatically
- **Tilde Expansion**: `~/path` expands to home directory
- **Command Abbreviations**: Type `l` for `ls`, with ambiguity detection
- **Inline Autocompletion**: Ghost text suggestions, Tab completion, PATH executable completion
- **Line Editing**: Cursor movement, word operations (Ctrl+arrows), kill/yank
- **Persistent History**: Configurable, filterable, with duplicate handling
- **Color Output**: Auto-detection with TTY, NO_COLOR, and TERM support
- **Colored Prompt**: Customizable with variables and colors
- **YAML Configuration**: `~/.config/jsishell/config.yaml`
- **Cross-Platform**: Linux, macOS, and Windows support

## Installation

### From Releases

Download the latest binary for your platform from the [Releases](https://github.com/sdejongh/jsishell/releases) page.

| Platform | Architecture | Binary |
|----------|--------------|--------|
| Linux | amd64 | `jsishell-linux-amd64` |
| Linux | arm64 | `jsishell-linux-arm64` |
| macOS | amd64 | `jsishell-darwin-amd64` |
| macOS | arm64 (Apple Silicon) | `jsishell-darwin-arm64` |
| Windows | amd64 | `jsishell-windows-amd64.exe` |

### From Source

```bash
# Clone the repository
git clone https://github.com/sdejongh/jsishell.git
cd jsishell

# Build
make build

# Or install to GOPATH/bin
make install
```

**Requirements**: Go 1.24+

## Usage

```bash
# Start the shell
./jsishell

# Show version
./jsishell --version

# Show help
./jsishell --help
```

Once in the shell, type `help` to see available commands.

## Built-in Commands

| Category | Commands |
|----------|----------|
| Utilities | `echo`, `exit`, `help`, `clear`, `env`, `reload`, `history` |
| Navigation | `cd`, `pwd` |
| File Operations | `ls`, `cp`, `mv`, `rm`, `mkdir`, `search` |
| Configuration | `init` |

All commands support `--help` for detailed usage information.

### ls Command

List directory contents with extensive options:

| Option | Description |
|--------|-------------|
| `-a, --all` | Include hidden files |
| `-d, --directory` | List directories only |
| `-l, --long` | Long format (permissions, owner, group, size, date, name) |
| `-R, --recursive` | List subdirectories recursively |
| `-v, --verbose` | Verbose mode (implies -l, adds file type indicator) |
| `-q, --quiet` | Only show file names |
| `-s, --sort=<spec>` | Sort by: `name`, `size`, `time`, `dir` (prefix with `!` to reverse) |
| `-e, --exclude=<glob>` | Exclude files matching glob pattern (can be repeated) |

**Sort examples**:
```bash
ls --sort=name           # Sort by name (default)
ls --sort=!time          # Sort by modification time, newest first
ls --sort=dir,name       # Directories first, then by name
ls --sort=!size,name     # By size descending, then by name
```

### cp Command

Copy files and directories:

| Option | Description |
|--------|-------------|
| `-r, --recursive` | Copy directories recursively |
| `-v, --verbose` | Print file names as they are copied |
| `-f, --force` | Overwrite existing files without prompting |
| `-e, --exclude=<glob>` | Exclude files matching glob pattern (can be repeated) |

### rm Command

Remove files and directories:

| Option | Description |
|--------|-------------|
| `-r, --recursive` | Remove directories and their contents recursively |
| `-f, --force` | Ignore nonexistent files, never prompt |
| `-y, --yes` | Skip confirmation prompts (same as --force) |
| `-v, --verbose` | Print file names as they are removed |
| `-q, --quiet` | Suppress error messages |
| `-e, --exclude=<glob>` | Exclude files matching glob pattern (can be repeated) |

### search Command

Find files and directories matching glob patterns with logical expressions:

| Option | Description |
|--------|-------------|
| `-r, --recursive` | Search recursively in subdirectories |
| `-l, --level=<n>` | Maximum depth level (0 = unlimited) |
| `-a, --absolute` | Display absolute paths |

**Type predicates** (case-insensitive):
- `isFile` - Match regular files only
- `isDir` - Match directories only
- `isLink` / `isSymlink` - Match symbolic links
- `isExec` - Match executable files

**Logical operators** (case-insensitive):
- `AND`, `&&` - Both patterns must match
- `OR`, `||` - Either pattern must match (default)
- `XOR`, `^` - Exactly one pattern must match
- `NOT`, `!` - Negate the following pattern
- `( )` - Group expressions

**Operator precedence** (highest to lowest): NOT > AND > XOR > OR

**Examples**:
```bash
# Simple searches
search . "*.go"                         # Find Go files in current directory
search . "*.go" "*.md" -r               # Find Go OR Markdown files recursively

# With logical operators
search . "*.go" AND NOT "*_test.go" -r  # Find Go files excluding tests
search . "test_*" AND "*.py" -r         # Find Python test files

# With type predicates
search . isDir -r                       # Find all directories
search . isExec -r                      # Find all executable files
search . "*.go" AND isFile -r           # Find Go files only (not directories)

# Complex expressions
search . "(" "*.go" OR "*.md" ")" AND NOT "*_test*" -r

# Quoting for literal patterns (find files named AND, OR, etc.)
search . "AND"                          # Find file named 'AND'
search . "isFile"                       # Find file named 'isFile'
```

## Command Line Editing

| Key | Action |
|-----|--------|
| `Left/Right` | Move cursor one character |
| `Ctrl+Left/Right` | Move cursor one word |
| `Home` / `Ctrl+A` | Move to beginning of line |
| `End` / `Ctrl+E` | Move to end of line |
| `Backspace` | Delete character before cursor |
| `Delete` | Delete character at cursor |
| `Ctrl+K` | Delete from cursor to end of line |
| `Ctrl+U` | Delete from cursor to beginning of line |
| `Ctrl+W` | Delete word before cursor |
| `Up/Down` | Navigate command history |
| `Tab` | Accept inline suggestion |
| `Tab Tab` | Show all completion candidates |
| `Ctrl+C` | Interrupt current command |

## Configuration

Generate a default configuration file:

```bash
init
```

This creates `~/.config/jsishell/config.yaml`:

```yaml
prompt: "%{green}%u@%h%{/}:%{blue}%~%{/}%$ "
history:
  max_size: 1000
  file: "~/.jsishell_history"
  ignore_duplicates: true
  ignore_space_prefix: true
colors:
  enabled: true
  directory: blue
  file: white
  executable: green
  error: red
abbreviations:
  enabled: true
```

Use `reload` command to apply configuration changes without restarting.

### Prompt Variables

| Variable | Description |
|----------|-------------|
| `%d` | Current directory (full path) |
| `%D` | Current directory (basename) |
| `%~` | Current directory with ~ for home |
| `%u` | Username |
| `%h` | Hostname (short) |
| `%H` | Hostname (full) |
| `%t` | Time (HH:MM) |
| `%T` | Time (HH:MM:SS) |
| `%$` | Shell indicator ($ for user, # for root) |
| `%n` | Newline |
| `%%` | Literal % |

### Prompt Colors

Use `%{color}` to start a color and `%{/}` or `%{reset}` to reset:

- **Colors**: `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`
- **Bright**: `bright_black`, `bright_red`, `bright_green`, etc.
- **Styles**: `bold`, `dim`, `underline`

**Example**: `%{bold}%{green}%u@%h%{/}:%{blue}%~%{/}%$ `

## Command Abbreviations

When enabled, you can type abbreviated command names:

- `l` executes `ls` (only command starting with 'l')
- `cl` executes `clear` (disambiguates from `cd`, `cp`)
- `c` shows error: "Ambiguous command 'c'. Did you mean: cd, clear, cp?"

Abbreviations only apply to built-in commands, not external programs.

## Windows Support

JSIShell works on Windows with the following adaptations:

- **Drive navigation**: Type `c:` or `D:` to change drives
- **Windows paths**: Backslashes work correctly (`d:\pictures\2024`)
- **PATH variable**: Uses `Path` (Windows convention) instead of `PATH`
- **Executable detection**: Recognizes `.exe`, `.cmd`, `.bat`, `.com`, `.ps1`
- **Owner/group display**: Shows `-` placeholders in `ls -l` (Windows uses SIDs)

## Architecture

```
cmd/jsishell/           # Entry point
internal/
├── shell/              # REPL orchestration
├── lexer/              # Tokenization
├── parser/             # AST generation (with tilde/glob expansion)
├── executor/           # Command execution engine
├── builtins/           # Built-in commands (17 commands)
├── completion/         # Inline autocompletion with PATH caching
├── history/            # Persistent command history
├── config/             # YAML configuration loader
├── terminal/           # Terminal I/O, line editor, colors
├── env/                # Environment variables
└── errors/             # Sentinel errors
```

## Performance

| Metric | Target | Achieved |
|--------|--------|----------|
| Shell startup time | < 100ms | **0.14ms** |
| Input latency | < 10ms | **< 0.01ms** |
| Autocompletion | < 100ms | **< 1ms** |
| Binary size | < 10MB | **~2.8MB** |

## Dependencies

Minimal external dependencies:
- `golang.org/x/term` - Cross-platform terminal handling
- `gopkg.in/yaml.v3` - YAML configuration

## Development

```bash
# Run all checks (format, vet, lint, test)
make check

# Run tests
make test

# Run tests with race detector
make test-race

# Generate coverage report
make test-coverage

# Build for all platforms
make release VERSION=1.0.0
```

## Future Features

The following features may be implemented on request:
- Pipes and redirections (`|`, `>`, `<`, `>>`)
- Shell scripting (conditionals, loops, functions)
- Aliases
- Job control (background processes, `&`, `fg`, `bg`)
- Additional built-ins (`cat`, `touch`, `grep`, `which`, etc.)

## License

MIT License - See [LICENSE](LICENSE) for details.
