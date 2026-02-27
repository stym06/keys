package tui

import (
	"fmt"
	"strings"

	"keys/db"

	tea "github.com/charmbracelet/bubbletea"
)

type editField int

const (
	editFieldName editField = iota
	editFieldValue
)

type EditModel struct {
	oldName string
	name    string
	value   string
	focus   editField
	done    bool
	message string
}

func NewEdit(key db.Key) EditModel {
	return EditModel{
		oldName: key.Name,
		name:    key.Name,
		value:   key.Value,
		focus:   editFieldName,
	}
}

func (m EditModel) Message() string { return m.message }

func (m EditModel) Init() tea.Cmd {
	return nil
}

func (m EditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			m.done = true
			m.message = "Cancelled"
			return m, tea.Quit
		case "tab":
			if m.focus == editFieldName {
				m.focus = editFieldValue
			} else {
				m.focus = editFieldName
			}
		case "enter":
			if m.name != "" && m.value != "" {
				if err := db.UpdateKey(m.oldName, m.name, m.value); err != nil {
					m.message = fmt.Sprintf("Error: %v", err)
				} else {
					m.message = fmt.Sprintf("Updated %s", m.name)
				}
				m.done = true
				return m, tea.Quit
			}
		case "backspace":
			if m.focus == editFieldName && len(m.name) > 0 {
				m.name = m.name[:len(m.name)-1]
			} else if m.focus == editFieldValue && len(m.value) > 0 {
				m.value = m.value[:len(m.value)-1]
			}
		default:
			if len(msg.String()) == 1 {
				if m.focus == editFieldName {
					m.name += msg.String()
				} else {
					m.value += msg.String()
				}
			}
		}
	}
	return m, nil
}

func (m EditModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder

	// Name field
	nameIcon := searchIconStyle.Render("✎")
	nameLabel := labelStyle.Render("Name: ")
	nameContent := nameIcon + " " + nameLabel + m.name
	if m.focus == editFieldName {
		nameContent += searchInputStyle.Render("_")
		b.WriteString(searchBarFocusedStyle.Render(nameContent))
	} else {
		b.WriteString(searchBarStyle.Render(nameContent))
	}
	b.WriteString("\n")

	// Value field
	valIcon := searchIconStyle.Render("✎")
	valLabel := labelStyle.Render("Value: ")
	valContent := valIcon + " " + valLabel + m.value
	if m.focus == editFieldValue {
		valContent += searchInputStyle.Render("_")
		b.WriteString(searchBarFocusedStyle.Render(valContent))
	} else {
		b.WriteString(searchBarStyle.Render(valContent))
	}
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  tab switch field  enter save  esc cancel"))
	return b.String()
}
