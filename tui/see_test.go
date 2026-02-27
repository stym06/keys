package tui

import (
	"strings"
	"testing"
	"time"

	"keys/db"

	tea "github.com/charmbracelet/bubbletea"
)

func key(t tea.KeyType) tea.Msg {
	return tea.KeyMsg{Type: t}
}

func char(c rune) tea.Msg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{c}}
}

func sampleKeys() []db.Key {
	now := time.Now().Unix()
	return []db.Key{
		{Name: "API_KEY", Value: "sk-123", UpdatedAt: now},
		{Name: "DB_HOST", Value: "localhost", UpdatedAt: now - 86400*45},
		{Name: "SECRET", Value: "s3cret", UpdatedAt: now - 86400*100},
	}
}

func TestNewSee(t *testing.T) {
	m := NewSee(sampleKeys())
	if m.Done() {
		t.Error("should not be done")
	}
	if m.masked {
		t.Error("NewSee should not be masked")
	}
	if m.cursor != 0 {
		t.Error("cursor should start at 0")
	}
}

func TestNewPeek(t *testing.T) {
	m := NewPeek(sampleKeys())
	if !m.masked {
		t.Error("NewPeek should be masked")
	}
}

func TestFilteredKeys(t *testing.T) {
	m := NewSee(sampleKeys())

	// No filter = all keys
	if len(m.filteredKeys()) != 3 {
		t.Errorf("expected 3 keys, got %d", len(m.filteredKeys()))
	}

	// Type "api"
	m.input = "api"
	filtered := m.filteredKeys()
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered key, got %d", len(filtered))
	}
	if filtered[0].Name != "API_KEY" {
		t.Errorf("expected API_KEY, got %s", filtered[0].Name)
	}
}

func TestFilteredKeysCaseInsensitive(t *testing.T) {
	m := NewSee(sampleKeys())
	m.input = "DB"
	filtered := m.filteredKeys()
	if len(filtered) != 1 {
		t.Fatalf("expected 1, got %d", len(filtered))
	}
	if filtered[0].Name != "DB_HOST" {
		t.Errorf("expected DB_HOST, got %s", filtered[0].Name)
	}
}

func TestCursorNavigation(t *testing.T) {
	m := NewSee(sampleKeys())

	// Move down
	result, _ := m.Update(key(tea.KeyDown))
	m = result.(SeeModel)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.cursor)
	}

	// Move down again
	result, _ = m.Update(key(tea.KeyDown))
	m = result.(SeeModel)
	if m.cursor != 2 {
		t.Errorf("expected cursor 2, got %d", m.cursor)
	}

	// Should not go past end
	result, _ = m.Update(key(tea.KeyDown))
	m = result.(SeeModel)
	if m.cursor != 2 {
		t.Errorf("expected cursor 2 (clamped), got %d", m.cursor)
	}

	// Move up
	result, _ = m.Update(key(tea.KeyUp))
	m = result.(SeeModel)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.cursor)
	}

	// Move up to 0
	result, _ = m.Update(key(tea.KeyUp))
	m = result.(SeeModel)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}

	// Should not go below 0
	result, _ = m.Update(key(tea.KeyUp))
	m = result.(SeeModel)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 (clamped), got %d", m.cursor)
	}
}

func TestCtrlPCtrlNNavigation(t *testing.T) {
	m := NewSee(sampleKeys())

	result, _ := m.Update(key(tea.KeyCtrlN))
	m = result.(SeeModel)
	if m.cursor != 1 {
		t.Errorf("ctrl+n: expected cursor 1, got %d", m.cursor)
	}

	result, _ = m.Update(key(tea.KeyCtrlP))
	m = result.(SeeModel)
	if m.cursor != 0 {
		t.Errorf("ctrl+p: expected cursor 0, got %d", m.cursor)
	}
}

func TestSelection(t *testing.T) {
	m := NewSee(sampleKeys())

	// Select first key
	result, _ := m.Update(key(tea.KeySpace))
	m = result.(SeeModel)
	if !m.selected["API_KEY"] {
		t.Error("API_KEY should be selected")
	}

	// Deselect
	result, _ = m.Update(key(tea.KeySpace))
	m = result.(SeeModel)
	if m.selected["API_KEY"] {
		t.Error("API_KEY should be deselected")
	}
}

func TestSelectedFromMatches(t *testing.T) {
	m := NewSee(sampleKeys())
	m.selected["API_KEY"] = true
	m.selected["SECRET"] = true

	matches := m.filteredKeys()
	sel := m.selectedFromMatches(matches)
	if len(sel) != 2 {
		t.Errorf("expected 2 selected, got %d", len(sel))
	}
}

func TestSearchInput(t *testing.T) {
	m := NewSee(sampleKeys())

	// Type "a"
	result, _ := m.Update(char('a'))
	m = result.(SeeModel)
	if m.input != "a" {
		t.Errorf("expected input 'a', got %q", m.input)
	}

	// Type "p"
	result, _ = m.Update(char('p'))
	m = result.(SeeModel)
	if m.input != "ap" {
		t.Errorf("expected input 'ap', got %q", m.input)
	}

	// Cursor should reset on each keystroke
	if m.cursor != 0 {
		t.Errorf("cursor should reset to 0 on input")
	}
}

func TestBackspace(t *testing.T) {
	m := NewSee(sampleKeys())
	m.input = "abc"

	result, _ := m.Update(key(tea.KeyBackspace))
	m = result.(SeeModel)
	if m.input != "ab" {
		t.Errorf("expected 'ab', got %q", m.input)
	}

	// Backspace on empty should not panic
	m.input = ""
	result, _ = m.Update(key(tea.KeyBackspace))
	m = result.(SeeModel)
	if m.input != "" {
		t.Errorf("expected empty, got %q", m.input)
	}
}

func TestEscQuits(t *testing.T) {
	m := NewSee(sampleKeys())

	result, cmd := m.Update(key(tea.KeyEsc))
	final := result.(SeeModel)
	if !final.Done() {
		t.Error("should be done after esc")
	}
	if cmd == nil {
		t.Error("should return tea.Quit cmd")
	}
}

func TestCtrlCQuits(t *testing.T) {
	m := NewSee(sampleKeys())

	result, cmd := m.Update(key(tea.KeyCtrlC))
	final := result.(SeeModel)
	if !final.Done() {
		t.Error("should be done after ctrl+c")
	}
	if cmd == nil {
		t.Error("should return tea.Quit cmd")
	}
}

func TestEnterNoMatchTransitionsToAddName(t *testing.T) {
	m := NewSee(nil) // no keys
	m.input = "NEW_KEY"

	result, _ := m.Update(key(tea.KeyEnter))
	m = result.(SeeModel)
	if m.state != stateAddName {
		t.Error("should transition to stateAddName")
	}
	if m.newName != "NEW_KEY" {
		t.Errorf("expected newName 'NEW_KEY', got %q", m.newName)
	}
}

func TestEnterWithMatchesDoesNothing(t *testing.T) {
	m := NewSee(sampleKeys())

	result, _ := m.Update(key(tea.KeyEnter))
	m = result.(SeeModel)
	if m.state != stateSearch {
		t.Error("should stay in stateSearch when matches exist")
	}
}

func TestAddNameToAddValue(t *testing.T) {
	m := NewSee(nil)
	m.state = stateAddName
	m.newName = "MY_KEY"

	result, _ := m.Update(key(tea.KeyEnter))
	m = result.(SeeModel)
	if m.state != stateAddValue {
		t.Error("should transition to stateAddValue")
	}
}

func TestAddNameTyping(t *testing.T) {
	m := NewSee(nil)
	m.state = stateAddName

	result, _ := m.Update(char('K'))
	m = result.(SeeModel)
	if m.newName != "K" {
		t.Errorf("expected 'K', got %q", m.newName)
	}
}

func TestAddValueTyping(t *testing.T) {
	m := NewSee(nil)
	m.state = stateAddValue
	m.newName = "KEY"

	result, _ := m.Update(char('v'))
	m = result.(SeeModel)
	if m.newVal != "v" {
		t.Errorf("expected 'v', got %q", m.newVal)
	}
}

func TestAddNameBackspace(t *testing.T) {
	m := NewSee(nil)
	m.state = stateAddName
	m.newName = "abc"

	result, _ := m.Update(key(tea.KeyBackspace))
	m = result.(SeeModel)
	if m.newName != "ab" {
		t.Errorf("expected 'ab', got %q", m.newName)
	}
}

func TestAddValueBackspace(t *testing.T) {
	m := NewSee(nil)
	m.state = stateAddValue
	m.newVal = "xyz"

	result, _ := m.Update(key(tea.KeyBackspace))
	m = result.(SeeModel)
	if m.newVal != "xy" {
		t.Errorf("expected 'xy', got %q", m.newVal)
	}
}

func TestAgeIndicator(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name      string
		updatedAt int64
		wantColor string // partial string check
	}{
		{"zero", 0, "●"}, // dim style
		{"fresh", now - 86400*5, "●"},
		{"medium", now - 86400*45, "●"},
		{"old", now - 86400*100, "●"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ageIndicator(tc.updatedAt)
			if !strings.Contains(result, "●") {
				t.Errorf("expected dot indicator, got %q", result)
			}
		})
	}
}

func TestAgeIndicatorBoundaries(t *testing.T) {
	now := time.Now().Unix()

	// Exactly at boundaries
	tests := []struct {
		name      string
		updatedAt int64
	}{
		{"29 days", now - 86400*29},  // green
		{"30 days", now - 86400*30},  // yellow
		{"89 days", now - 86400*89},  // yellow
		{"90 days", now - 86400*90},  // red
		{"365 days", now - 86400*365}, // red
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ageIndicator(tc.updatedAt)
			if result == "" {
				t.Error("ageIndicator should not return empty string")
			}
			if !strings.Contains(result, "●") {
				t.Error("should contain dot indicator")
			}
		})
	}
}

func TestMaskedModeValues(t *testing.T) {
	m := NewPeek(sampleKeys())

	view := m.View()
	// In masked mode, values should show as ***
	if !strings.Contains(view, "***") {
		t.Error("masked mode should show *** for values")
	}
	// Actual values should NOT appear
	if strings.Contains(view, "sk-123") {
		t.Error("masked mode should not show actual values")
	}
}

func TestRevealInPeekMode(t *testing.T) {
	m := NewPeek(sampleKeys())

	// Press 'r' to reveal first key
	result, _ := m.Update(char('r'))
	m = result.(SeeModel)

	if !m.revealed["API_KEY"] {
		t.Error("API_KEY should be revealed after pressing r")
	}

	view := m.View()
	// The revealed key's value should appear
	if !strings.Contains(view, "sk-123") {
		t.Error("revealed key should show actual value")
	}

	// Press 'r' again to hide
	result, _ = m.Update(char('r'))
	m = result.(SeeModel)
	if m.revealed["API_KEY"] {
		t.Error("API_KEY should be hidden after second r press")
	}
}

func TestRKeyInNonMaskedModeTypesChar(t *testing.T) {
	m := NewSee(sampleKeys())

	result, _ := m.Update(char('r'))
	m = result.(SeeModel)
	// In non-masked mode, 'r' should type into search
	if m.input != "r" {
		t.Errorf("expected input 'r', got %q", m.input)
	}
}

func TestCtrlEExport(t *testing.T) {
	m := NewSee(sampleKeys())

	// Select a key first
	result, _ := m.Update(key(tea.KeySpace))
	m = result.(SeeModel)

	// Press ctrl+e
	result, cmd := m.Update(key(tea.KeyCtrlE))
	m = result.(SeeModel)

	if !m.EnvExport() {
		t.Error("envExport should be true")
	}
	if !m.Done() {
		t.Error("should be done after ctrl+e")
	}
	if cmd == nil {
		t.Error("should return tea.Quit")
	}
	if len(m.EnvExportKeys()) != 1 {
		t.Errorf("expected 1 export key, got %d", len(m.EnvExportKeys()))
	}
}

func TestCtrlEExportCursorFallback(t *testing.T) {
	m := NewSee(sampleKeys())
	// No selection, cursor on first key

	result, _ := m.Update(key(tea.KeyCtrlE))
	m = result.(SeeModel)

	if !m.EnvExport() {
		t.Error("should export cursor key when nothing selected")
	}
	if len(m.EnvExportKeys()) != 1 {
		t.Errorf("expected 1 key, got %d", len(m.EnvExportKeys()))
	}
	if m.EnvExportKeys()[0].Name != "API_KEY" {
		t.Errorf("expected API_KEY, got %s", m.EnvExportKeys()[0].Name)
	}
}

func TestCtrlENoKeysDoesNothing(t *testing.T) {
	m := NewSee(nil) // no keys

	result, _ := m.Update(key(tea.KeyCtrlE))
	m = result.(SeeModel)

	if m.EnvExport() {
		t.Error("should not export when no keys")
	}
	if m.Done() {
		t.Error("should not be done when no keys to export")
	}
}

func TestViewNotEmptyWhenNotDone(t *testing.T) {
	m := NewSee(sampleKeys())
	view := m.View()
	if view == "" {
		t.Error("View should not be empty when not done")
	}
}

func TestViewEmptyWhenDone(t *testing.T) {
	m := NewSee(sampleKeys())
	m.done = true
	view := m.View()
	if view != "" {
		t.Error("View should be empty when done")
	}
}

func TestViewShowsSearchPlaceholder(t *testing.T) {
	m := NewSee(sampleKeys())
	view := m.View()
	if !strings.Contains(view, "Search keys") {
		t.Error("should show search placeholder")
	}
}

func TestViewShowsKeyNames(t *testing.T) {
	m := NewSee(sampleKeys())
	view := m.View()
	if !strings.Contains(view, "API_KEY") {
		t.Error("should show API_KEY")
	}
	if !strings.Contains(view, "DB_HOST") {
		t.Error("should show DB_HOST")
	}
}

func TestViewEmptyState(t *testing.T) {
	m := NewSee(nil)
	view := m.View()
	if !strings.Contains(view, "No keys stored yet") {
		t.Error("should show empty message")
	}
}

func TestViewNoMatchesHint(t *testing.T) {
	m := NewSee(sampleKeys())
	m.input = "zzzzzzz"
	view := m.View()
	if !strings.Contains(view, "No keys found") {
		t.Error("should show no keys found message")
	}
	if !strings.Contains(view, "enter to add") {
		t.Error("should show add hint")
	}
}

func TestViewSelectedCount(t *testing.T) {
	m := NewSee(sampleKeys())
	m.selected["API_KEY"] = true
	m.selected["DB_HOST"] = true

	view := m.View()
	if !strings.Contains(view, "2 selected") {
		t.Error("should show selection count")
	}
}

func TestViewMaskedHelpText(t *testing.T) {
	m := NewPeek(sampleKeys())
	view := m.View()
	if !strings.Contains(view, "r reveal") {
		t.Error("masked mode should show reveal hint")
	}
}

func TestViewNormalHelpText(t *testing.T) {
	m := NewSee(sampleKeys())
	view := m.View()
	if !strings.Contains(view, "enter add key") {
		t.Error("normal mode should show add key hint")
	}
	if !strings.Contains(view, "ctrl+e") {
		t.Error("should show ctrl+e hint")
	}
}

func TestViewAddNameState(t *testing.T) {
	m := NewSee(nil)
	m.state = stateAddName
	m.newName = "TEST"

	view := m.View()
	if !strings.Contains(view, "Name:") {
		t.Error("should show Name label")
	}
	if !strings.Contains(view, "TEST") {
		t.Error("should show entered name")
	}
}

func TestViewAddValueState(t *testing.T) {
	m := NewSee(nil)
	m.state = stateAddValue
	m.newName = "KEY"
	m.newVal = "val"

	view := m.View()
	if !strings.Contains(view, "Name:") {
		t.Error("should show Name label")
	}
	if !strings.Contains(view, "Value:") {
		t.Error("should show Value label")
	}
}

func TestViewAgeIndicators(t *testing.T) {
	m := NewSee(sampleKeys())
	view := m.View()
	// Should contain age dots
	if !strings.Contains(view, "●") {
		t.Error("view should contain age indicators")
	}
}

func TestCopiedFlashClears(t *testing.T) {
	m := NewSee(sampleKeys())
	m.copied = "done"
	m.copiedFmt = "env"
	m.copiedN = 1

	// Moving cursor should clear copied
	result, _ := m.Update(key(tea.KeyDown))
	m = result.(SeeModel)
	if m.copied != "" {
		t.Error("copied should be cleared on cursor move")
	}
}
