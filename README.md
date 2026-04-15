# try-bedazzled

A beautiful, themeable directory manager for scratch experiments. A Go rewrite of [tobi/try-cli](https://github.com/tobi/try-cli) using [Charmbracelet](https://charm.sh) libraries.

> Almost as fast as C. Way faster than Ruby. And look at this TUI.

## Features

- **Fuzzy finder** — type to filter, select with arrow keys
- **Smart icons** — auto-detects project type from name (🐹 Go, 🐍 Python, 🦀 Rust, 🐘 Postgres, ...)
- **4 built-in themes** — default, catppuccin, dracula, minimal
- **Configurable layout** — columns, icons, date position, search style — all in TOML
- **Ghost autocomplete** — fish-style dim suggestions as you type
- **Delete confirmation** — bordered dialog before removing directories
- **Git integration** — `try clone <url>`, `try worktree <name>`
- **Shell wrapper** — changes your shell's cwd (bash, zsh, fish)

## Install

### From source

```bash
go install github.com/hmk/try-bedazzled/cmd/try@latest
```

### From release

Download from [Releases](https://github.com/hmk/try-bedazzled/releases), then:

```bash
# macOS / Linux
tar xzf try-bedazzled_*.tar.gz
sudo mv try /usr/local/bin/
```

### Arch Linux

```bash
# .deb and .rpm packages available in releases
sudo dpkg -i try-bedazzled_*.deb    # Debian/Ubuntu
sudo rpm -i try-bedazzled_*.rpm     # Fedora/RHEL
```

## Setup

Add to your shell config (`.bashrc`, `.zshrc`, or `config.fish`):

```bash
eval "$(try init)"

# Or with a custom directory:
eval "$(try init ~/my/experiments)"
```

## Usage

```bash
try                      # Show help
try redis                # Fuzzy find or create "YYYY-MM-DD-redis"
try cd redis             # Explicit selector
try clone <url> [name]   # Git clone into dated directory
try worktree <name>      # Git worktree (or mkdir if not in repo)
try theme                # Interactive theme picker with live preview
try --help               # Full help
```

### Keyboard shortcuts

| Key | Action |
|---|---|
| Type | Filter entries (fuzzy) |
| `↑`/`↓` | Navigate |
| `Enter` | Select / create new / confirm delete |
| `Ctrl-D` | Mark for deletion |
| `Ctrl-R` | Rename |
| `Esc` | Cancel |

## Themes

Pick a theme interactively:

```bash
try theme
```

Or set it in `~/.config/try/config.toml`:

```toml
theme = "catppuccin"   # default, catppuccin, dracula, minimal
```

Or override with an environment variable:

```bash
TRY_THEME=dracula try redis
```

### Custom themes

Drop a `.toml` file in `~/.config/try/themes/`:

```toml
# ~/.config/try/themes/nord.toml

[colors]
accent  = "#88C0D0"
dim     = "#4C566A"
text    = "#ECEFF4"
match   = "#EBCB8B"
danger  = "#BF616A"
success = "#A3BE8C"

[symbols]
cursor  = ">"
folder  = "📂"
created = "+"
deleted = "x"

[layout]
show_icons   = true
show_date    = "right"
show_time    = true
columns      = ["icon", "name", "date", "time"]
search_style = "bordered"
```

## Performance

Benchmarked on macOS (Apple Silicon M4) with 100 try directories:

| Operation | C | Go (this) | Ruby |
|---|---|---|---|
| Startup + scan | 1.5ms | 2.4ms | 40ms |
| Fuzzy match + select | 2.0ms | 3.7ms | 44ms |
| Version (startup only) | 1.2ms | 2.0ms | 36ms |
| Binary size | 92KB | 5.1MB | interpreted |

> Go is ~1.6x slower than C for raw operations, but **15-20x faster than Ruby** — and you get a themeable TUI, config files, and content-aware icons for that trade-off.
>
> Measured with [hyperfine](https://github.com/sharkdp/hyperfine). See `bench/` for reproduction.

## Architecture

```
cmd/try/main.go          Entry point
internal/
  cli/                   Cobra commands (init, exec, clone, worktree, theme)
  dirs/                  Directory scanning, naming, slug normalization
  fuzzy/                 Scoring algorithm (ported from C)
  shell/                 Shell script generation, escaping, init wrappers
  theme/                 TOML theming, icon registry, config read/write
  tui/                   Bubble Tea model, Lip Gloss styles, theme picker
```

## Credits

Inspired by [Tobias Lutke's](https://github.com/tobi) original [try](https://github.com/tobi/try) (Ruby) and [try-cli](https://github.com/tobi/try-cli) (C). Built with [Charmbracelet](https://charm.sh) libraries.

## License

[BSD 3-Clause](LICENSE)
