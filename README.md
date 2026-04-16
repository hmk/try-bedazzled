# try-bedazzled

**A beautiful, themeable scratch-directory manager for your terminal.**

A Go rewrite of [tobi/try-cli](https://github.com/tobi/try-cli) (C) using [Charmbracelet](https://charm.sh) libraries —
almost as fast as C, 15× faster than Ruby, and it actually has a TUI.

---

## What it does

`try` gives every experiment a dated home directory and a fast fuzzy selector to jump back into it:

```
─────────────────────────────────────────
 ▸ redis                         4 matches
─────────────────────────────────────────
▸ 📂 2026-04-14  redis-cluster         3h ago
  📦 2026-04-12  redis-sentinel        2d ago
  🦀 2026-04-11  redis-rs-bench        3d ago
  📂 2026-04-09  redis-lua-scripts     5d ago
─────────────────────────────────────────
  enter select  •  ctrl-d delete  •  ctrl-r rename  •  ctrl-p preview on  •  ctrl-g settings  •  esc quit
```

Type to filter. Press Enter to `cd`. Type a new name and press Enter to create. That's it.

---

## Install

```bash
go install github.com/hmk/try-bedazzled/cmd/try@latest
```

Or download a pre-built binary from [Releases](https://github.com/hmk/try-bedazzled/releases):

```bash
tar xzf try-bedazzled_*.tar.gz
sudo mv try /usr/local/bin/
```

---

## Setup

Add one line to your shell config and you're done:

**bash / zsh** (`.bashrc` or `.zshrc`):
```bash
eval "$(try init)"
```

**fish** (`config.fish`):
```fish
try init | source
```

Use a custom directory:
```bash
eval "$(try init ~/workspace/experiments)"
```

---

## Commands

| Command | What it does |
|---|---|
| `try` | Open the fuzzy selector |
| `try redis` | Pre-filter the selector for "redis" |
| `try clone <url> [name]` | Git clone into a dated directory |
| `try worktree <name>` | Git worktree (or mkdir if not in a repo) |
| `try theme` | Interactive theme picker with live preview |
| `try settings` | Open the settings menu |
| `try init [path]` | Print shell integration script |

---

## Keyboard shortcuts

| Key | Action |
|---|---|
| Type | Filter (fuzzy, real-time) |
| `↑` / `↓` or `Ctrl-K` / `Ctrl-J` | Navigate |
| `Enter` | Select / create / confirm delete |
| `Ctrl-D` | Mark/unmark entry for deletion |
| `Ctrl-R` | Rename (preserves date prefix) |
| `Ctrl-P` | Toggle file-tree preview panel |
| `Ctrl-G` | Open settings menu |
| `Esc` | Cancel |

---

## Differentiators

### Themes

Four built-in themes, switchable live with `try theme`:

| Theme | Description |
|---|---|
| `default` | Purple accents, clean and modern |
| `catppuccin` | Catppuccin Mocha pastels |
| `dracula` | Dark and saturated |
| `minimal` | Gruvbox-ish, ASCII-safe |

Set in config or override per-session:

```toml
# ~/.config/try/config.toml
theme = "catppuccin"
```

```bash
TRY_THEME=dracula try redis
```

Drop a `.toml` in `~/.config/try/themes/` to add your own:

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
cursor  = "▸"
folder  = "📂"

[layout]
show_icons   = true
show_date    = "right"
show_time    = true
search_style = "bordered"
```

### Content-aware icons

The icon registry reads slug words to pick a project-appropriate emoji automatically:

| Slug word | Icon |
|---|---|
| `go`, `golang` | 🐹 |
| `rust` | 🦀 |
| `python`, `py` | 🐍 |
| `postgres`, `pg` | 🐘 |
| `redis` | 📦 |
| `docker`, `k8s` | 🐳 |
| `react`, `vue`, `svelte` | ⚛️ |
| `ml`, `llm`, `ai` | 🤖 |
| …and 50+ more | |

Override or extend with your own in config:

```toml
[custom_icons]
django   = "🎭"
temporal = "⏱"
shopify  = "🛍"
```

### Live folder preview

Press `Ctrl-P` to toggle a live file-tree preview of the highlighted directory inline in the selector. State persists across launches — turn it off once, it stays off.

### Settings menu

`try settings` (or `Ctrl-,` inside the selector) opens a Huh form that covers every preference and writes `~/.config/try/config.toml`:

- Theme
- Display mode (fullscreen / inline)
- Preview panel default
- Emoji icons toggle
- Add custom slug→icon mappings

---

## Configuration

Full config reference:

```toml
# ~/.config/try/config.toml

tries_path      = "~/tries"          # where your directories live
theme           = "default"          # built-in or custom theme name
display_mode    = "inline"           # "inline" | "fullscreen" (alt screen)
inline_min_rows = 15                 # minimum rows in inline mode
preview_enabled = true               # file-tree preview panel
show_emojis     = true               # folder/type icons

[custom_icons]
django = "🎭"
rust   = "🦀"
```

---

## Performance

Benchmarked on macOS (Apple Silicon M4) with 100 try directories:

| Operation | C (try-cli) | Go (try-bedazzled) | Ruby (try) |
|---|---|---|---|
| Startup + scan | 1.5 ms | 2.4 ms | 40 ms |
| Fuzzy match + select | 2.0 ms | 3.7 ms | 44 ms |
| Version (startup only) | 1.2 ms | 2.0 ms | 36 ms |
| Binary size | 92 KB | 5.1 MB | interpreted |

> Go is ~1.6× slower than C for raw operations, but **15–20× faster than Ruby** — and you get a
> themeable TUI, configurable layouts, content-aware icons, and a settings menu for that trade-off.
>
> Measured with [hyperfine](https://github.com/sharkdp/hyperfine). Reproduce with `bench/bench.sh`.

---

## Architecture

```
cmd/try/main.go
internal/
  cli/       Cobra commands: init, exec, clone, worktree, theme, settings
  dirs/      Directory scanning, naming, date-prefix parsing, slug normalization
  fuzzy/     Scoring algorithm (ported from tobi/try-cli's C)
  shell/     Shell script generation (cd, mkdir, delete, rename, init wrappers)
  theme/     TOML theming, icon registry, config read/write, built-in themes
  tui/       Bubble Tea model, Lip Gloss styles, theme picker table, file tree
```

---

## Credits

Inspired by [Tobias Lutke's](https://github.com/tobi) original [try](https://github.com/tobi/try) (Ruby) and [try-cli](https://github.com/tobi/try-cli) (C).
Built with [Charmbracelet](https://charm.sh) libraries: Bubble Tea, Lip Gloss, Bubbles, Huh.

## License

[BSD 3-Clause](LICENSE)
