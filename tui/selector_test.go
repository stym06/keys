package tui

import (
	"strings"
	"testing"

	"keys/db"

	tea "github.com/charmbracelet/bubbletea"
)

func selectorKeys() []db.Key {
	return []db.Key{
		{Name: "KEY_A", Value: "a"},
		{Name: "KEY_B", Value: "b"},
		{Name: "KEY_C", Value: "c"},
	}
}

func TestNewSelector(t *testing.T) {
	m := NewSelector(selectorKeys())
	if m.cursor != 0 {
		t.Error("cursor should start at 0")
	}
	if m.done {
		t.Error("should not be done")
	}
	if m.aborted {
		t.Error("should not be aborted")
	}
}

func TestSelectorNavigationDown(t *testing.T) {
	m := NewSelector(selectorKeys())

	result, _ := m.Update(key(tea.KeyDown))
	m = result.(SelectorModel)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.cursor)
	}

	result, _ = m.Update(key(tea.KeyDown))
	m = result.(SelectorModel)
	if m.cursor != 2 {
		t.Errorf("expected cursor 2, got %d", m.cursor)
	}

	// Should not go past end
	result, _ = m.Update(key(tea.KeyDown))
	m = result.(SelectorModel)
	if m.cursor != 2 {
		t.Errorf("expected cursor 2 (clamped), got %d", m.cursor)
	}
}

func TestSelectorNavigationUp(t *testing.T) {
	m := NewSelector(selectorKeys())
	m.cursor = 2

	result, _ := m.Update(key(tea.KeyUp))
	m = result.(SelectorModel)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.cursor)
	}

	result, _ = m.Update(key(tea.KeyUp))
	m = result.(SelectorModel)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}

	// Should not go below 0
	result, _ = m.Update(key(tea.KeyUp))
	m = result.(SelectorModel)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 (clamped), got %d", m.cursor)
	}
}

func TestSelectorJKNavigation(t *testing.T) {
	m := NewSelector(selectorKeys())

	result, _ := m.Update(char('j'))
	m = result.(SelectorModel)
	if m.cursor != 1 {
		t.Errorf("j: expected cursor 1, got %d", m.cursor)
	}

	result, _ = m.Update(char('k'))
	m = result.(SelectorModel)
	if m.cursor != 0 {
		t.Errorf("k: expected cursor 0, got %d", m.cursor)
	}
}

func TestSelectorToggle(t *testing.T) {
	m := NewSelector(selectorKeys())

	// Toggle first
	result, _ := m.Update(key(tea.KeySpace))
	m = result.(SelectorModel)
	if !m.selected[0] {
		t.Error("item 0 should be selected")
	}

	// Toggle again to deselect
	result, _ = m.Update(key(tea.KeySpace))
	m = result.(SelectorModel)
	if m.selected[0] {
		t.Error("item 0 should be deselected")
	}
}

func TestSelectorToggleAll(t *testing.T) {
	m := NewSelector(selectorKeys())

	// Toggle all on
	result, _ := m.Update(char('a'))
	m = result.(SelectorModel)
	for i := range selectorKeys() {
		if !m.selected[i] {
			t.Errorf("item %d should be selected", i)
		}
	}

	// Toggle all off
	result, _ = m.Update(char('a'))
	m = result.(SelectorModel)
	if len(m.selected) != 0 {
		t.Errorf("all items should be deselected, got %d selected", len(m.selected))
	}
}

func TestSelectorEnter(t *testing.T) {
	m := NewSelector(selectorKeys())
	m.selected[0] = true
	m.selected[2] = true

	result, cmd := m.Update(key(tea.KeyEnter))
	m = result.(SelectorModel)

	if !m.done {
		t.Error("should be done after enter")
	}
	if m.Aborted() {
		t.Error("should not be aborted")
	}
	if cmd == nil {
		t.Error("should return tea.Quit")
	}

	selected := m.SelectedKeys()
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected keys, got %d", len(selected))
	}
	if selected[0].Name != "KEY_A" || selected[1].Name != "KEY_C" {
		t.Errorf("unexpected selected keys: %v", selected)
	}
}

func TestSelectorAbortQ(t *testing.T) {
	m := NewSelector(selectorKeys())

	result, cmd := m.Update(char('q'))
	m = result.(SelectorModel)

	if !m.Aborted() {
		t.Error("should be aborted after q")
	}
	if !m.done {
		t.Error("should be done")
	}
	if cmd == nil {
		t.Error("should return tea.Quit")
	}
}

func TestSelectorAbortEsc(t *testing.T) {
	m := NewSelector(selectorKeys())

	result, _ := m.Update(key(tea.KeyEsc))
	m = result.(SelectorModel)

	if !m.Aborted() {
		t.Error("should be aborted after esc")
	}
}

func TestSelectorAbortCtrlC(t *testing.T) {
	m := NewSelector(selectorKeys())

	result, _ := m.Update(key(tea.KeyCtrlC))
	m = result.(SelectorModel)

	if !m.Aborted() {
		t.Error("should be aborted after ctrl+c")
	}
}

func TestSelectedKeysEmpty(t *testing.T) {
	m := NewSelector(selectorKeys())
	selected := m.SelectedKeys()
	if len(selected) != 0 {
		t.Errorf("expected 0 selected, got %d", len(selected))
	}
}

func TestSelectorViewNotEmpty(t *testing.T) {
	m := NewSelector(selectorKeys())
	view := m.View()
	if view == "" {
		t.Error("view should not be empty")
	}
	if !strings.Contains(view, "Select keys") {
		t.Error("should show title")
	}
}

func TestSelectorViewDone(t *testing.T) {
	m := NewSelector(selectorKeys())
	m.done = true
	view := m.View()
	if view != "" {
		t.Error("view should be empty when done")
	}
}

func TestSelectorViewShowsKeys(t *testing.T) {
	m := NewSelector(selectorKeys())
	view := m.View()
	if !strings.Contains(view, "KEY_A") {
		t.Error("should show KEY_A")
	}
	if !strings.Contains(view, "KEY_B") {
		t.Error("should show KEY_B")
	}
	if !strings.Contains(view, "KEY_C") {
		t.Error("should show KEY_C")
	}
}

func TestSelectorViewSelectedCheck(t *testing.T) {
	m := NewSelector(selectorKeys())
	m.selected[0] = true

	view := m.View()
	if !strings.Contains(view, "[x]") {
		t.Error("selected item should show [x]")
	}
}

func TestSelectorViewHelpText(t *testing.T) {
	m := NewSelector(selectorKeys())
	view := m.View()
	if !strings.Contains(view, "space") {
		t.Error("should mention space")
	}
	if !strings.Contains(view, "enter") {
		t.Error("should mention enter")
	}
}
