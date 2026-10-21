# Queens

A Go / Bubble Tea terminal puzzle with connected, colored regions. Every
generated puzzle is checked to have exactly one solution.

This codebase was mostly coded by AI

```sh
go run .
go run . -size 8  # sizes 5–9; default 7
```

Place one 👑 in every row, column, and region. Queens cannot touch, including
diagonally. Unlike chess queens, distant queens on a diagonal are allowed.
Regions are distinguished by color; conflicting queens show `!` underneath.
Each cell is four character columns wide and two lines tall.
The app recognizes completion automatically.

* Left click cycles empty → X → 👑 → empty (commits on mouse release).
* Drag from an empty square to paint Xs. Drag from an X to erase Xs.
  Drags preserve queens and never place them. Starting on a queen paints Xs
  on the other squares while preserving the queen.
* Arrow keys or WASD move the bracketed selection; space cycles its mark.
  Keyboard and mouse input are both always available.
* `m` opens settings. Use up/down to select and left/right to change.
  Automatic X placement defaults to off; when enabled, placing a crown marks
  empty cells in its row, column, region, and adjacent squares. Removing a
  crown clears its automatic Xs, preserving manual Xs and automatic Xs still
  covered by another crown. Board sizes range from 5–9; changing size
  starts a new puzzle when you return with `m` or Escape.
* `r` clears the board; `n` generates a new puzzle; `q`, Escape, or Ctrl+C quits.

Use a terminal with mouse reporting, Unicode emoji, and color support.
The default board needs 32 columns and 25 rows. Smaller windows show a resize
prompt. Crown rendering depends on the terminal's emoji font and width handling.

```sh
go test ./...
go build -o queens .
go get github.com/charmbracelet/bubbletea@latest
go mod tidy
```
