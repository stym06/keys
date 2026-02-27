package tui

import (
	"fmt"
	"strings"

	"keys/db"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
)

type SelectorModel struct {
	keys     []db.Key
	cursor   int
	selected map[int]bool
	done     bool
	aborted  bool
}

func NewSelector(keys []db.Key) SelectorModel {
	return SelectorModel{
		keys:     keys,
		selected: make(map[int]bool),
	}
}

func (m SelectorModel) SelectedKeys() []db.Key {
	var result []db.Key
	for i, k := range m.keys {
		if m.selected[i] {
			result = append(result, k)
		}
	}
	return result
}

func (m SelectorModel) Aborted() bool {
	return m.aborted
}

func (m SelectorModel) Init() tea.Cmd {
	return nil
}

func (m SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.aborted = true
			m.done = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.keys)-1 {
				m.cursor++
			}
		case " ":
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "enter":
			m.done = true
			return m, tea.Quit
		case "a":
			allSelected := len(m.selected) == len(m.keys)
			if allSelected {
				// check if all are true
				allTrue := true
				for i := range m.keys {
					if !m.selected[i] {
						allTrue = false
						break
					}
				}
				if allTrue {
					m.selected = make(map[int]bool)
				} else {
					for i := range m.keys {
						m.selected[i] = true
					}
				}
			} else {
				for i := range m.keys {
					m.selected[i] = true
				}
			}
		}
	}
	return m, nil
}

func (m SelectorModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Select keys for .env file"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("space: toggle  a: toggle all  enter: confirm  q: quit"))
	b.WriteString("\n\n")

	for i, k := range m.keys {
		cursor := "  "
		if m.cursor == i {
			cursor = cursorStyle.Render("> ")
		}

		check := "[ ]"
		name := k.Name
		if m.selected[i] {
			check = selectedStyle.Render("[x]")
			name = selectedStyle.Render(k.Name)
		}

		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, check, name))
	}

	return b.String()
}
