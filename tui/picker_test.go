package tui

import (
	"strings"
	"testing"

	"keys/db"

	tea "github.com/charmbracelet/bubbletea"
)

func pickerKeys() []db.Key {
	return []db.Key{
		{Name: "API_KEY", Value: "sk-123"},
		{Name: "DB_HOST", Value: "localhost"},
		{Name: "SECRET", Value: "s3cret"},
	}
}

func TestNewPicker(t *testing.T) {
	m := NewPicker(pickerKeys())
	if m.cursor != 0 {
		t.Error("cursor should start at 0")
	}
	if m.Aborted() {
		t.Error("should not be aborted")
	}
	if m.Picked() != nil {
		t.Error("should not have picked anything")
	}
}

func TestPickerSearch(t *testing.T) {
	m := NewPicker(pickerKeys())

	// Type "db"
	result, _ := m.Update(char('d'))
	m = result.(PickerModel)
	result, _ = m.Update(char('b'))
	m = result.(PickerModel)

	matches := m.filteredKeys()
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Name != "DB_HOST" {
		t.Errorf("expected DB_HOST, got %s", matches[0].Name)
	}
}

func TestPickerSearchCaseInsensitive(t *testing.T) {
	m := NewPicker(pickerKeys())
	m.input = "API"
	matches := m.filteredKeys()
	if len(matches) != 1 || matches[0].Name != "API_KEY" {
		t.Errorf("case insensitive search failed: %v", matches)
	}
}

func TestPickerCursorNavigation(t *testing.T) {
	m := NewPicker(pickerKeys())

	result, _ := m.Update(key(tea.KeyDown))
	m = result.(PickerModel)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.cursor)
	}

	result, _ = m.Update(key(tea.KeyUp))
	m = result.(PickerModel)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}

	// Clamp at top
	result, _ = m.Update(key(tea.KeyUp))
	m = result.(PickerModel)
	if m.cursor != 0 {
		t.Error("cursor should clamp at 0")
	}
}

func TestPickerCursorClampsAtBottom(t *testing.T) {
	m := NewPicker(pickerKeys())
	m.cursor = 2

	result, _ := m.Update(key(tea.KeyDown))
	m = result.(PickerModel)
	if m.cursor != 2 {
		t.Error("cursor should clamp at last item")
	}
}

func TestPickerCursorResetsOnInput(t *testing.T) {
	m := NewPicker(pickerKeys())
	m.cursor = 2

	result, _ := m.Update(char('a'))
	m = result.(PickerModel)
	if m.cursor != 0 {
		t.Error("cursor should reset to 0 on input")
	}
}

func TestPickerBackspace(t *testing.T) {
	m := NewPicker(pickerKeys())
	m.input = "abc"

	result, _ := m.Update(key(tea.KeyBackspace))
	m = result.(PickerModel)
	if m.input != "ab" {
		t.Errorf("expected 'ab', got %q", m.input)
	}
}

func TestPickerBackspaceEmpty(t *testing.T) {
	m := NewPicker(pickerKeys())

	result, _ := m.Update(key(tea.KeyBackspace))
	m = result.(PickerModel)
	if m.input != "" {
		t.Error("should remain empty")
	}
}

func TestPickerEnterSelectsKey(t *testing.T) {
	m := NewPicker(pickerKeys())

	// Move to second item
	result, _ := m.Update(key(tea.KeyDown))
	m = result.(PickerModel)

	// Press enter
	result, cmd := m.Update(key(tea.KeyEnter))
	m = result.(PickerModel)

	if m.Picked() == nil {
		t.Fatal("should have picked a key")
	}
	if m.Picked().Name != "DB_HOST" {
		t.Errorf("expected DB_HOST, got %s", m.Picked().Name)
	}
	if m.Picked().Value != "localhost" {
		t.Errorf("expected localhost, got %s", m.Picked().Value)
	}
	if !m.done {
		t.Error("should be done")
	}
	if cmd == nil {
		t.Error("should return tea.Quit")
	}
}

func TestPickerEnterWithFilter(t *testing.T) {
	m := NewPicker(pickerKeys())
	m.input = "secret"

	result, _ := m.Update(key(tea.KeyEnter))
	m = result.(PickerModel)

	if m.Picked() == nil {
		t.Fatal("should have picked a key")
	}
	if m.Picked().Name != "SECRET" {
		t.Errorf("expected SECRET, got %s", m.Picked().Name)
	}
}

func TestPickerEnterNoMatches(t *testing.T) {
	m := NewPicker(pickerKeys())
	m.input = "zzzzz"

	result, _ := m.Update(key(tea.KeyEnter))
	m = result.(PickerModel)

	if m.Picked() != nil {
		t.Error("should not pick anything when no matches")
	}
	if m.done {
		t.Error("should not be done when no matches")
	}
}

func TestPickerEscAborts(t *testing.T) {
	m := NewPicker(pickerKeys())

	result, cmd := m.Update(key(tea.KeyEsc))
	m = result.(PickerModel)

	if !m.Aborted() {
		t.Error("should be aborted")
	}
	if !m.done {
		t.Error("should be done")
	}
	if cmd == nil {
		t.Error("should return tea.Quit")
	}
}

func TestPickerCtrlCAborts(t *testing.T) {
	m := NewPicker(pickerKeys())

	result, _ := m.Update(key(tea.KeyCtrlC))
	m = result.(PickerModel)

	if !m.Aborted() {
		t.Error("should be aborted")
	}
}

func TestPickerViewShowsKeys(t *testing.T) {
	m := NewPicker(pickerKeys())
	view := m.View()

	if !strings.Contains(view, "API_KEY") {
		t.Error("should show API_KEY")
	}
	if !strings.Contains(view, "DB_HOST") {
		t.Error("should show DB_HOST")
	}
	if !strings.Contains(view, "SECRET") {
		t.Error("should show SECRET")
	}
}

func TestPickerViewPlaceholder(t *testing.T) {
	m := NewPicker(pickerKeys())
	view := m.View()
	if !strings.Contains(view, "Type to search") {
		t.Error("should show placeholder")
	}
}

func TestPickerViewNoMatches(t *testing.T) {
	m := NewPicker(pickerKeys())
	m.input = "zzzzz"
	view := m.View()
	if !strings.Contains(view, "No matching keys") {
		t.Error("should show no matching keys message")
	}
}

func TestPickerViewDone(t *testing.T) {
	m := NewPicker(pickerKeys())
	m.done = true
	if m.View() != "" {
		t.Error("view should be empty when done")
	}
}

func TestPickerViewHelpText(t *testing.T) {
	m := NewPicker(pickerKeys())
	view := m.View()
	if !strings.Contains(view, "enter") {
		t.Error("should mention enter")
	}
	if !strings.Contains(view, "esc") {
		t.Error("should mention esc")
	}
}

func TestPickerCtrlPCtrlN(t *testing.T) {
	m := NewPicker(pickerKeys())

	result, _ := m.Update(key(tea.KeyCtrlN))
	m = result.(PickerModel)
	if m.cursor != 1 {
		t.Errorf("ctrl+n: expected 1, got %d", m.cursor)
	}

	result, _ = m.Update(key(tea.KeyCtrlP))
	m = result.(PickerModel)
	if m.cursor != 0 {
		t.Errorf("ctrl+p: expected 0, got %d", m.cursor)
	}
}
