# VHS tape for try-bedazzled
#
# Requires:
#   - `try` binary on PATH (built from ./cmd/try)
#   - fixture tries dir at /tmp/demo-tries (see .github/workflows/demo.yml)
#   - Fragment Mono font installed
#
# Regenerate locally:
#   go build -o /tmp/try ./cmd/try && sudo install /tmp/try /usr/local/bin/try
#   rm -rf /tmp/demo-tries && mkdir -p /tmp/demo-tries
#   for d in 2026-04-14-redis-cache 2026-04-13-go-api 2026-04-12-react-dashboard \
#            2026-04-11-ml-pipeline 2026-04-10-rust-bench 2026-04-08-postgres-migration \
#            2026-04-05-terraform-modules 2026-03-28-docker-compose-auth \
#            2026-03-14-python-cli 2026-02-28-kubernetes-operator; do
#     mkdir -p "/tmp/demo-tries/$d" && touch "/tmp/demo-tries/$d/README.md"
#   done
#   vhs demo.tape

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

# Use a throwaway config dir so the demo never touches a real user's settings,
# and starts from the bedazzled default every run.
Env XDG_CONFIG_HOME "/tmp/try-demo-config"
Hide
Type "rm -rf /tmp/try-demo-config"
Enter
Show

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

# Pop open settings and switch to the rainbow theme.
# (Theme order: bedazzled, catppuccin, dracula, minimal, rainbow.)
Ctrl+G
Sleep 1500ms
Enter         # enter the Theme picker
Sleep 800ms
Down          # catppuccin
Sleep 400ms
Down          # dracula
Sleep 400ms
Down          # minimal
Sleep 400ms
Down          # rainbow
Sleep 1200ms
Enter         # apply
Sleep 800ms
Escape        # save & close settings (reopens selector)
Sleep 1500ms

# Show off rainbow_matches with a fresh fuzzy filter.
Type "py cli"
Sleep 2s
Down
Sleep 600ms
Up
Sleep 800ms

# Exit cleanly.
Escape
Sleep 400ms
