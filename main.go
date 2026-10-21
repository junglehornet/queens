package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mark uint8

const (
	empty mark = iota
	cross
	queen
)
const boardX, boardY, cellW = 2, 5, 4

type generatedMsg puzzle
type model struct {
	p              puzzle
	size           int
	marks          []mark
	cursor         int
	width, height  int
	loading        bool
	settings       bool
	settingCursor  int
	settingsSize   int
	autoCross      bool
	autoOwners     map[int]map[int]bool // Automatically marked cell → crowns responsible.
	down, dragging bool
	start, last    int
	paint          mark
	visited        map[int]bool
}

func newModel(n int) model { return model{size: n, loading: true} }
func generateCmd(n int) tea.Cmd {
	return func() tea.Msg { return generatedMsg(generate(n, rand.New(rand.NewSource(time.Now().UnixNano())))) }
}
func (m model) Init() tea.Cmd { return generateCmd(m.size) }

func (m model) fits() bool { return m.width >= boardX+m.size*cellW+2 && m.height >= boardY+m.size*2+6 }
func (m model) cellAt(x, y int) int {
	if x < boardX || x >= boardX+m.size*cellW || y < boardY || y >= boardY+m.size*2 {
		return -1
	}
	return ((y-boardY)/2)*m.size + (x-boardX)/cellW
}

func (m *model) cycle(i int) {
	previous := m.marks[i]
	// A direct edit replaces any automatic mark on this cell.
	delete(m.autoOwners, i)
	m.marks[i] = (previous + 1) % 3
	if previous == queen {
		for cell, owners := range m.autoOwners {
			delete(owners, i)
			if len(owners) == 0 {
				if m.marks[cell] == cross {
					m.marks[cell] = empty
				}
				delete(m.autoOwners, cell)
			}
		}
	}
	if m.marks[i] != queen || !m.autoCross {
		return
	}
	if m.autoOwners == nil {
		m.autoOwners = make(map[int]map[int]bool)
	}
	r, c := i/m.size, i%m.size
	for j, v := range m.marks {
		rr, cc := j/m.size, j%m.size
		if (v == empty || m.autoOwners[j] != nil) && (r == rr || c == cc || m.p.regions[i] == m.p.regions[j] || (abs(r-rr) <= 1 && abs(c-cc) <= 1)) {
			m.marks[j] = cross
			if m.autoOwners[j] == nil {
				m.autoOwners[j] = make(map[int]bool)
			}
			m.autoOwners[j][i] = true
		}
	}
}

func (m model) updateSettings(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "m", "esc":
		m.settings = false
		if m.settingsSize != m.size {
			m.size = m.settingsSize
			m.loading = true
			return m, generateCmd(m.size)
		}
	case "up", "w", "down", "s", "tab", "shift+tab":
		m.settingCursor = 1 - m.settingCursor
	case "left", "a", "right", "d", " ", "enter":
		if m.settingCursor == 0 {
			m.autoCross = !m.autoCross
		} else {
			delta := 1
			if key == "left" || key == "a" {
				delta = -1
			}
			n := m.settingsSize + delta
			if n >= 5 && n <= 9 {
				m.settingsSize = n
			}
		}
	}
	return m, nil
}
func (m *model) paintCell(i int) {
	if i < 0 || m.visited[i] {
		return
	}
	m.visited[i] = true
	if m.marks[i] != queen {
		delete(m.autoOwners, i)
		m.marks[i] = m.paint
	}
}

// Interpolate skipped mouse reports so fast drags leave a continuous stroke.
func (m *model) stroke(to int) {
	if to < 0 {
		m.last = -1
		return
	}
	if m.last < 0 {
		m.paintCell(to)
		m.last = to
		return
	}
	x, y := m.last%m.size, m.last/m.size
	tx, ty := to%m.size, to/m.size
	dx, dy := abs(tx-x), -abs(ty-y)
	sx, sy := -1, -1
	if x < tx {
		sx = 1
	}
	if y < ty {
		sy = 1
	}
	err := dx + dy
	for {
		m.paintCell(y*m.size + x)
		if x == tx && y == ty {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x += sx
		}
		if e2 <= dx {
			err += dx
			y += sy
		}
	}
	m.last = to
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case generatedMsg:
		if puzzle(msg).n != m.size {
			return m, nil
		}
		m.p = puzzle(msg)
		m.marks = make([]mark, m.size*m.size)
		m.autoOwners = nil
		m.loading = false
		m.cursor = 0
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.down = false
	case tea.KeyMsg:
		key := msg.String()
		if key == "q" || key == "ctrl+c" {
			return m, tea.Quit
		}
		if m.settings {
			return m.updateSettings(key)
		}
		if key == "m" {
			m.settings = true
			m.settingsSize = m.size
			m.down = false
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "n":
			if !m.loading {
				m.loading = true
				m.down = false
				return m, generateCmd(m.size)
			}
		case "r":
			if !m.loading {
				m.marks = make([]mark, m.size*m.size)
				m.autoOwners = nil
				m.down = false
			}
		default:
			if m.loading || !m.fits() {
				break
			}
			m.down = false
			r, c := m.cursor/m.size, m.cursor%m.size
			switch msg.String() {
			case "up", "w":
				if r > 0 {
					m.cursor -= m.size
				}
			case "down", "s":
				if r < m.size-1 {
					m.cursor += m.size
				}
			case "left", "a":
				if c > 0 {
					m.cursor--
				}
			case "right", "d":
				if c < m.size-1 {
					m.cursor++
				}
			case " ", "enter":
				m.cycle(m.cursor)
			}
		}
	case tea.MouseMsg:
		if m.settings || m.loading || !m.fits() {
			break
		}
		i := m.cellAt(msg.X, msg.Y)
		switch msg.Action {
		case tea.MouseActionPress:
			if msg.Button != tea.MouseButtonLeft || i < 0 {
				break
			}
			m.down = true
			m.dragging = false
			m.start = i
			m.last = i
			m.cursor = i
			m.paint = cross
			if m.marks[i] == cross {
				m.paint = empty
			}
			m.visited = make(map[int]bool)
		case tea.MouseActionMotion:
			if !m.down {
				break
			}
			if i != m.start && !m.dragging {
				m.dragging = true
				m.paintCell(m.start)
			}
			if m.dragging {
				m.stroke(i)
			}
			if i >= 0 {
				m.cursor = i
			}
		case tea.MouseActionRelease:
			if !m.down {
				break
			}
			if i != m.start && !m.dragging {
				m.dragging = true
				m.paintCell(m.start)
			}
			if m.dragging {
				m.stroke(i)
			} else if i == m.start {
				m.cycle(i)
			}
			m.down = false
		}
	}
	return m, nil
}

func (m model) conflicts() map[int]bool {
	bad := make(map[int]bool)
	for i, a := range m.marks {
		if a != queen {
			continue
		}
		for j := i + 1; j < len(m.marks); j++ {
			if m.marks[j] != queen {
				continue
			}
			r, c, rr, cc := i/m.size, i%m.size, j/m.size, j%m.size
			if r == rr || c == cc || m.p.regions[i] == m.p.regions[j] || (abs(r-rr) <= 1 && abs(c-cc) <= 1) {
				bad[i] = true
				bad[j] = true
			}
		}
	}
	return bad
}

var palette = []string{"#D6A5D9", "#97C9EB", "#F2C078", "#A8D5A2", "#E99B9B", "#B5AEEC", "#80CBC4", "#E5D58A", "#C6B4A1"}

func (m model) View() string {
	if m.settings {
		auto := "Off"
		if m.autoCross {
			auto = "On"
		}
		selectors := []string{"  ", "  "}
		selectors[m.settingCursor] = "> "
		return fmt.Sprintf("\n  SETTINGS\n\n  %sAuto place Xs when crown is placed: %s\n  %sBoard size: %d × %d\n\n  Up / down: select • left / right: change\n  Space / enter: toggle or increase\n  m / Esc: return • q quit\n\n  Changing board size starts a new puzzle on return.\n", selectors[0], auto, selectors[1], m.settingsSize, m.settingsSize)
	}
	if m.loading {
		return "\n  QUEENS\n\n  Generating a uniquely solvable puzzle…\n\n  m settings • q quit\n"
	}
	if !m.fits() {
		return fmt.Sprintf("\n  Resize terminal to at least %d × %d.\n\n  m settings • q quit\n", boardX+m.size*cellW+2, boardY+m.size*2+6)
	}
	bad := m.conflicts()
	count := 0
	for _, v := range m.marks {
		if v == queen {
			count++
		}
	}
	status := fmt.Sprintf("%d / %d queens", count, m.size)
	if len(bad) > 0 {
		status += "  •  Conflicting queens marked !"
	} else if count == m.size {
		status = "Solved! 👑  Press n for a new puzzle."
	}
	var b strings.Builder
	b.WriteString("\n  QUEENS  /  a little logic, a little royalty\n\n")
	b.WriteString("  One queen per row, column & region. Queens cannot touch.\n\n")
	for r := 0; r < m.size; r++ {
		for line := 0; line < 2; line++ {
			b.WriteString("  ")
			for c := 0; c < m.size; c++ {
				i := r*m.size + c
				left, right := " ", " "
				if i == m.cursor {
					left, right = "[", "]"
				}
				content := "  "
				if m.marks[i] == cross {
					content = "X "
				}
				if m.marks[i] == queen {
					content = "👑 "
				}
				if line == 1 {
					content = "  "
					if bad[i] {
						content = "! "
					}
				}
				style := lipgloss.NewStyle().Background(lipgloss.Color(palette[m.p.regions[i]])).Foreground(lipgloss.Color("#191724"))
				b.WriteString(style.Render(left + content + right))
			}
			b.WriteByte('\n')
		}
	}
	fmt.Fprintf(&b, "\n  %s\n\n  Click / space: empty → X → 👑 → empty\n  Drag: paint Xs; start on X to erase. Queens are preserved.\n  Arrows / WASD move • r reset • n new puzzle • m settings • q quit\n", status)
	return b.String()
}

func main() {
	size := flag.Int("size", 7, "board size (5–9)")
	flag.Parse()
	if *size < 5 || *size > 9 {
		fmt.Fprintln(os.Stderr, "size must be between 5 and 9")
		os.Exit(2)
	}
	if _, err := tea.NewProgram(newModel(*size), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
