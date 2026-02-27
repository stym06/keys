package tui

import (
	"strings"
	"testing"

	"keys/db"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewEdit(t *testing.T) {
	k := db.Key{Name: "MY_KEY", Value: "secret"}
	m := NewEdit(k)

	if m.name != "MY_KEY" {
		t.Errorf("expected name MY_KEY, got %q", m.name)
	}
	if m.value != "secret" {
		t.Errorf("expected value secret, got %q", m.value)
	}
	if m.oldName != "MY_KEY" {
		t.Errorf("expected oldName MY_KEY, got %q", m.oldName)
	}
	if m.focus != editFieldName {
		t.Error("focus should start on name field")
	}
}

func TestEditTabSwitchFocus(t *testing.T) {
	k := db.Key{Name: "KEY", Value: "val"}
	m := NewEdit(k)

	// Initially on name
	if m.focus != editFieldName {
		t.Error("should start on name field")
	}

	// Tab to value
	result, _ := m.Update(key(tea.KeyTab))
	m = result.(EditModel)
	if m.focus != editFieldValue {
		t.Error("should switch to value field")
	}

	// Tab back to name
	result, _ = m.Update(key(tea.KeyTab))
	m = result.(EditModel)
	if m.focus != editFieldName {
		t.Error("should switch back to name field")
	}
}

func TestEditTypingName(t *testing.T) {
	k := db.Key{Name: "", Value: "val"}
	m := NewEdit(k)

	result, _ := m.Update(char('A'))
	m = result.(EditModel)
	if m.name != "A" {
		t.Errorf("expected 'A', got %q", m.name)
	}

	result, _ = m.Update(char('B'))
	m = result.(EditModel)
	if m.name != "AB" {
		t.Errorf("expected 'AB', got %q", m.name)
	}
}

func TestEditTypingValue(t *testing.T) {
	k := db.Key{Name: "KEY", Value: ""}
	m := NewEdit(k)

	// Switch to value field
	result, _ := m.Update(key(tea.KeyTab))
	m = result.(EditModel)

	result, _ = m.Update(char('x'))
	m = result.(EditModel)
	if m.value != "x" {
		t.Errorf("expected 'x', got %q", m.value)
	}
}

func TestEditBackspaceName(t *testing.T) {
	k := db.Key{Name: "ABC", Value: "val"}
	m := NewEdit(k)

	result, _ := m.Update(key(tea.KeyBackspace))
	m = result.(EditModel)
	if m.name != "AB" {
		t.Errorf("expected 'AB', got %q", m.name)
	}
}

func TestEditBackspaceValue(t *testing.T) {
	k := db.Key{Name: "KEY", Value: "xyz"}
	m := NewEdit(k)

	// Switch to value field
	result, _ := m.Update(key(tea.KeyTab))
	m = result.(EditModel)

	result, _ = m.Update(key(tea.KeyBackspace))
	m = result.(EditModel)
	if m.value != "xy" {
		t.Errorf("expected 'xy', got %q", m.value)
	}
}

func TestEditBackspaceEmpty(t *testing.T) {
	k := db.Key{Name: "", Value: ""}
	m := NewEdit(k)

	// Should not panic on empty
	result, _ := m.Update(key(tea.KeyBackspace))
	m = result.(EditModel)
	if m.name != "" {
		t.Error("should remain empty")
	}
}

func TestEditEscCancel(t *testing.T) {
	k := db.Key{Name: "KEY", Value: "val"}
	m := NewEdit(k)

	result, cmd := m.Update(key(tea.KeyEsc))
	m = result.(EditModel)

	if !m.done {
		t.Error("should be done after esc")
	}
	if m.Message() != "Cancelled" {
		t.Errorf("expected 'Cancelled', got %q", m.Message())
	}
	if cmd == nil {
		t.Error("should return tea.Quit")
	}
}

func TestEditCtrlCCancel(t *testing.T) {
	k := db.Key{Name: "KEY", Value: "val"}
	m := NewEdit(k)

	result, cmd := m.Update(key(tea.KeyCtrlC))
	m = result.(EditModel)

	if !m.done {
		t.Error("should be done after ctrl+c")
	}
	if cmd == nil {
		t.Error("should return tea.Quit")
	}
}

func TestEditEnterEmptyDoesNotSave(t *testing.T) {
	k := db.Key{Name: "", Value: "val"}
	m := NewEdit(k)

	result, _ := m.Update(key(tea.KeyEnter))
	m = result.(EditModel)
	if m.done {
		t.Error("should not save with empty name")
	}
}

func TestEditEnterEmptyValueDoesNotSave(t *testing.T) {
	k := db.Key{Name: "KEY", Value: ""}
	m := NewEdit(k)

	result, _ := m.Update(key(tea.KeyEnter))
	m = result.(EditModel)
	if m.done {
		t.Error("should not save with empty value")
	}
}

func TestEditEnterSaves(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// First create the key in DB
	db.AddKey("EDIT_ME", "old_val")

	k := db.Key{Name: "EDIT_ME", Value: "new_val"}
	m := NewEdit(k)

	result, cmd := m.Update(key(tea.KeyEnter))
	m = result.(EditModel)

	if !m.done {
		t.Error("should be done after enter with valid data")
	}
	if cmd == nil {
		t.Error("should return tea.Quit")
	}
	if !strings.Contains(m.Message(), "Updated") {
		t.Errorf("expected Updated message, got %q", m.Message())
	}

	// Verify in DB
	got, err := db.GetKey("EDIT_ME")
	if err != nil {
		t.Fatalf("GetKey: %v", err)
	}
	if got.Value != "new_val" {
		t.Errorf("expected 'new_val', got %q", got.Value)
	}
}

func TestEditViewNotEmpty(t *testing.T) {
	k := db.Key{Name: "KEY", Value: "val"}
	m := NewEdit(k)

	view := m.View()
	if view == "" {
		t.Error("view should not be empty")
	}
	if !strings.Contains(view, "Name:") {
		t.Error("should show Name label")
	}
	if !strings.Contains(view, "Value:") {
		t.Error("should show Value label")
	}
	if !strings.Contains(view, "KEY") {
		t.Error("should show key name")
	}
	if !strings.Contains(view, "val") {
		t.Error("should show key value")
	}
}

func TestEditViewDone(t *testing.T) {
	k := db.Key{Name: "KEY", Value: "val"}
	m := NewEdit(k)
	m.done = true

	view := m.View()
	if view != "" {
		t.Error("view should be empty when done")
	}
}

func TestEditViewHelpText(t *testing.T) {
	k := db.Key{Name: "KEY", Value: "val"}
	m := NewEdit(k)

	view := m.View()
	if !strings.Contains(view, "tab") {
		t.Error("should mention tab in help")
	}
	if !strings.Contains(view, "enter") {
		t.Error("should mention enter in help")
	}
	if !strings.Contains(view, "esc") {
		t.Error("should mention esc in help")
	}
}

func TestEditViewFocusHighlight(t *testing.T) {
	k := db.Key{Name: "KEY", Value: "val"}
	m := NewEdit(k)

	viewName := m.View()

	// Switch to value
	result, _ := m.Update(key(tea.KeyTab))
	m = result.(EditModel)
	viewVal := m.View()

	// Views should differ (different field highlighted)
	if viewName == viewVal {
		t.Error("view should change when focus switches")
	}
}
