# VHS tape for try-bedazzled
#
# Requires:
#   - `try` binary on PATH (built from ./cmd/try)
#   - fixture tries dir at /tmp/demo-tries (see .github/workflows/demo.yml)
#   - Fragment Mono font installed
#
# Regenerate locally:
#   go build -o /tmp/try ./cmd/try
#   ./bench/setup.sh   # creates /tmp/try-bench-fixtures
#   TRIES_PATH=/tmp/try-bench-fixtures vhs demo.tape

Output demo.gif

Require try

Set Shell "bash"
Set FontFamily "Fragment Mono"
Set FontSize 16
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms
Set PlaybackSpeed 1.0

# Silent warm-up so the prompt is at a clean state.
Hide
Type "clear"
Enter
Show

# Launch the selector against the fixture tries dir.
Type "try exec --path /tmp/demo-tries"
Enter
Sleep 1500ms

# Fuzzy filter with a space (matches the dash).
Type "redis cache"
Sleep 1500ms

# Toggle preview panel.
Ctrl+P
Sleep 2s

# Navigate down a couple of times.
Down
Sleep 500ms
Down
Sleep 500ms
Up
Sleep 500ms

# Clear filter and try a different one.
Backspace@80ms 11
Type "go api"
Sleep 1500ms
Down
Sleep 600ms

# Pop open settings briefly.
Ctrl+G
Sleep 2s
Escape
Sleep 500ms

# Exit cleanly.
Escape
Sleep 400ms
