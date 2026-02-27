package tui

import (
	"strings"

	"keys/db"

	tea "github.com/charmbracelet/bubbletea"
)

type PickerModel struct {
	keys    []db.Key
	input   string
	cursor  int
	done    bool
	aborted bool
	picked  *db.Key
}

func NewPicker(keys []db.Key) PickerModel {
	return PickerModel{keys: keys}
}

func (m PickerModel) Picked() *db.Key { return m.picked }
func (m PickerModel) Aborted() bool   { return m.aborted }

func (m PickerModel) Init() tea.Cmd { return nil }

func (m PickerModel) filteredKeys() []db.Key {
	if m.input == "" {
		return m.keys
	}
	query := strings.ToLower(m.input)
	var result []db.Key
	for _, k := range m.keys {
		if strings.Contains(strings.ToLower(k.Name), query) {
			result = append(result, k)
		}
	}
	return result
}

func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.aborted = true
			m.done = true
			return m, tea.Quit
		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "ctrl+n":
			matches := m.filteredKeys()
			if m.cursor < len(matches)-1 {
				m.cursor++
			}
		case "enter":
			matches := m.filteredKeys()
			if len(matches) > 0 && m.cursor < len(matches) {
				picked := matches[m.cursor]
				m.picked = &picked
				m.done = true
				return m, tea.Quit
			}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
				m.cursor = 0
			}
		default:
			if len(msg.String()) == 1 {
				m.input += msg.String()
				m.cursor = 0
			}
		}
	}
	return m, nil
}

func (m PickerModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder

	// Search bar
	icon := searchIconStyle.Render(" ")
	var inputText string
	if m.input == "" {
		inputText = placeholderStyle.Render("Type to search...")
	} else {
		inputText = searchInputStyle.Render(m.input)
	}
	barContent := icon + " " + inputText + searchInputStyle.Render("_")
	b.WriteString(searchBarFocusedStyle.Render(barContent))
	b.WriteString("\n\n")

	matches := m.filteredKeys()
	if len(matches) == 0 {
		b.WriteString(dimStyle.Render("  No matching keys."))
		b.WriteString("\n")
	} else {
		for i, k := range matches {
			pointer := "  "
			if i == m.cursor {
				pointer = cursorStyle.Render("> ")
			}

			name := k.Name
			if i == m.cursor {
				name = labelStyle.Render(k.Name)
			} else {
				name = dimStyle.Render(k.Name)
			}

			b.WriteString(pointer + name + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  enter select  esc quit"))
	return b.String()
}
