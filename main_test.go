package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func testModel() model {
	return model{size: 5, p: puzzle{n: 5, regions: make([]int, 25)}, marks: make([]mark, 25), width: 100, height: 40}
}
func mouse(m model, action tea.MouseAction, cell int) model {
	v, _ := m.Update(tea.MouseMsg{X: boardX + (cell%5)*cellW + 1, Y: boardY + (cell/5)*2, Action: action, Button: tea.MouseButtonLeft})
	return v.(model)
}
func TestClickCycle(t *testing.T) {
	m := testModel()
	for _, want := range []mark{cross, queen, empty} {
		m = mouse(m, tea.MouseActionPress, 0)
		m = mouse(m, tea.MouseActionRelease, 0)
		if m.marks[0] != want {
			t.Fatalf("got %v, want %v", m.marks[0], want)
		}
	}
}
func TestDragPaintEraseAndPreserveQueens(t *testing.T) {
	m := testModel()
	m.marks[1] = queen
	m = mouse(m, tea.MouseActionPress, 0)
	m = mouse(m, tea.MouseActionMotion, 3)
	m = mouse(m, tea.MouseActionRelease, 4)
	for _, i := range []int{0, 2, 3, 4} {
		if m.marks[i] != cross {
			t.Fatalf("cell %d not painted", i)
		}
	}
	if m.marks[1] != queen {
		t.Fatal("drag overwrote queen")
	}
	m = mouse(m, tea.MouseActionPress, 0)
	m = mouse(m, tea.MouseActionMotion, 4)
	m = mouse(m, tea.MouseActionMotion, 0)
	m = mouse(m, tea.MouseActionRelease, 0)
	for _, i := range []int{0, 2, 3, 4} {
		if m.marks[i] != empty {
			t.Fatalf("cell %d not erased", i)
		}
	}
	if m.marks[1] != queen {
		t.Fatal("erase overwrote queen")
	}
}
func TestDragStartingOnXNeverPlacesQueen(t *testing.T) {
	m := testModel()
	m.marks[0] = cross
	m = mouse(m, tea.MouseActionPress, 0)
	m = mouse(m, tea.MouseActionRelease, 3)
	for _, v := range m.marks {
		if v != empty {
			t.Fatal("erase drag placed a mark")
		}
	}
}
func TestKeyboard(t *testing.T) {
	m := testModel()
	for _, key := range []tea.KeyMsg{{Type: tea.KeyRight}, {Type: tea.KeyRunes, Runes: []rune{'s'}}, {Type: tea.KeySpace}} {
		v, _ := m.Update(key)
		m = v.(model)
	}
	if m.cursor != 6 || m.marks[6] != cross {
		t.Fatal("keyboard navigation or space failed")
	}
}
func TestConflictsAndWin(t *testing.T) {
	m := testModel()
	for r, c := range []int{0, 2, 4, 1, 3} {
		i := r*5 + c
		m.marks[i] = queen
		m.p.regions[i] = r
	}
	if len(m.conflicts()) != 0 {
		t.Fatal("legal solution flagged")
	}
	m.marks[1] = queen
	if len(m.conflicts()) == 0 {
		t.Fatal("conflict missed")
	}
}

func press(m model, key tea.KeyMsg) (model, tea.Cmd) {
	v, cmd := m.Update(key)
	return v.(model), cmd
}

func TestSettings(t *testing.T) {
	m := testModel()
	if m.autoCross {
		t.Fatal("auto Xs must default off")
	}
	m, _ = press(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m, _ = press(m, tea.KeyMsg{Type: tea.KeySpace})
	if !m.settings || !m.autoCross || m.marks[0] != empty {
		t.Fatal("settings did not toggle independently of board")
	}
	m, _ = press(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = press(m, tea.KeyMsg{Type: tea.KeyRight})
	var cmd tea.Cmd
	m, cmd = press(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.settings || m.size != 6 || !m.loading || cmd == nil {
		t.Fatal("new size was not applied on return")
	}
	v, _ := m.Update(generatedMsg(puzzle{n: 5}))
	if !v.(model).loading {
		t.Fatal("accepted old puzzle generation")
	}
	v, _ = m.Update(generatedMsg(puzzle{n: 6, regions: make([]int, 36)}))
	if v.(model).loading || len(v.(model).marks) != 36 {
		t.Fatal("new board was not initialized")
	}
}

func TestAutoCross(t *testing.T) {
	m := testModel()
	for i := range m.p.regions {
		m.p.regions[i] = i / 5
	}
	m.p.regions[24] = 0
	m.marks[20] = queen
	m.marks[0] = cross
	m.cycle(0)
	if m.marks[1] != empty {
		t.Fatal("auto Xs enabled by default")
	}
	m.marks[0] = cross
	m.autoCross = true
	m.cycle(0)
	for _, i := range []int{1, 4, 5, 6, 10, 24} {
		if m.marks[i] != cross {
			t.Fatalf("cell %d not automatically crossed", i)
		}
	}
	if m.marks[0] != queen || m.marks[20] != queen || m.marks[12] != empty {
		t.Fatal("auto Xs overwrote a queen or unrelated square")
	}
}

func TestRemovingCrownClearsOnlyItsAutomaticCrosses(t *testing.T) {
	for _, keyboard := range []bool{false, true} {
		m := testModel()
		m.autoCross = true
		m.marks[1] = cross // Manual X predates the crown.
		cycle := func(cell int) {
			if keyboard {
				m.cursor = cell
				m, _ = press(m, tea.KeyMsg{Type: tea.KeySpace})
			} else {
				m = mouse(m, tea.MouseActionPress, cell)
				m = mouse(m, tea.MouseActionRelease, cell)
			}
		}
		cycle(0)
		cycle(0)
		if m.marks[2] != cross {
			t.Fatal("crown did not place automatic X")
		}
		// Manually painting an automatic X makes it a manual mark.
		m.paint = cross
		m.visited = make(map[int]bool)
		m.paintCell(3)
		// Cleanup must still work if the setting is subsequently disabled.
		m.autoCross = false
		cycle(0)
		for i, got := range m.marks {
			want := empty
			if i == 1 || i == 3 {
				want = cross
			}
			if got != want {
				t.Fatalf("keyboard=%v cell %d: got %v, want %v", keyboard, i, got, want)
			}
		}
	}
}

func TestAutomaticCrossesSharedByCrowns(t *testing.T) {
	m := testModel()
	m.autoCross = true
	m.cycle(0)
	m.cycle(0)
	m.cycle(12) // Automatic X becomes a second crown.
	m.cycle(0)
	if m.marks[1] != cross || m.marks[12] != queen {
		t.Fatal("removing first crown erased shared Xs or second crown")
	}
	m.cycle(12)
	for i, got := range m.marks {
		if got != empty {
			t.Fatalf("cell %d retained automatic mark after both crowns removed", i)
		}
	}
}

func TestResetClearsAutomaticCrossTracking(t *testing.T) {
	m := testModel()
	m.autoCross = true
	m.cycle(0)
	m.cycle(0)
	m, _ = press(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if len(m.autoOwners) != 0 {
		t.Fatal("reset retained automatic X ownership")
	}
	m.marks[1] = cross
	m.cycle(0)
	m.cycle(0)
	m.cycle(0)
	if m.marks[1] != cross {
		t.Fatal("stale ownership erased manual X after reset")
	}
}
