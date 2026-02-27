package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"keys/db"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type seeState int

const (
	stateSearch seeState = iota
	stateAddName
	stateAddValue
)

var (
	labelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	valueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	hintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
	copiedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	checkStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)

	searchBarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1).
			Width(50)

	searchBarFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("12")).
				Padding(0, 1).
				Width(50)

	searchIconStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	searchInputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	placeholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	ageGreenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	ageYellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	ageRedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

type SeeModel struct {
	keys      []db.Key
	input     string
	cursor    int
	selected  map[string]bool // checked keys by name
	state     seeState
	newName   string
	newVal    string
	done      bool
	message   string
	copied    string // flash message
	copiedFmt string // "export" or "env"
	copiedN   int    // number of keys copied

	// Masked/peek mode
	masked   bool
	revealed map[string]bool // keys whose values are revealed

	// Env export via ctrl+e
	envExport     bool
	envExportKeys []db.Key
}

func NewSee(keys []db.Key) SeeModel {
	return SeeModel{keys: keys, state: stateSearch, selected: make(map[string]bool), revealed: make(map[string]bool)}
}

func NewPeek(keys []db.Key) SeeModel {
	return SeeModel{keys: keys, state: stateSearch, selected: make(map[string]bool), masked: true, revealed: make(map[string]bool)}
}

func (m SeeModel) Done() bool          { return m.done }
func (m SeeModel) Message() string     { return m.message }
func (m SeeModel) EnvExport() bool     { return m.envExport }
func (m SeeModel) EnvExportKeys() []db.Key { return m.envExportKeys }

func (m SeeModel) Init() tea.Cmd {
	return nil
}

func copyToClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// selectedFromMatches returns checked keys that are in the current filtered view.
func (m SeeModel) selectedFromMatches(matches []db.Key) []db.Key {
	var result []db.Key
	for _, k := range matches {
		if m.selected[k.Name] {
			result = append(result, k)
		}
	}
	return result
}

func ageIndicator(updatedAt int64) string {
	if updatedAt == 0 {
		return dimStyle.Render("● ")
	}
	age := time.Since(time.Unix(updatedAt, 0))
	days := int(age.Hours() / 24)
	if days < 30 {
		return ageGreenStyle.Render("● ")
	} else if days < 90 {
		return ageYellowStyle.Render("● ")
	}
	return ageRedStyle.Render("● ")
}

func (m SeeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit
		case "up", "ctrl+p":
			if m.state == stateSearch && m.cursor > 0 {
				m.cursor--
				m.copied = ""
			}
		case "down", "ctrl+n":
			if m.state == stateSearch {
				matches := m.filteredKeys()
				if m.cursor < len(matches)-1 {
					m.cursor++
					m.copied = ""
				}
			}
		case " ":
			if m.state == stateSearch {
				matches := m.filteredKeys()
				if len(matches) > 0 && m.cursor < len(matches) {
					name := matches[m.cursor].Name
					m.selected[name] = !m.selected[name]
					if !m.selected[name] {
						delete(m.selected, name)
					}
					m.copied = ""
				}
				return m, nil
			}
		case "r":
			if m.state == stateSearch && m.masked {
				matches := m.filteredKeys()
				if len(matches) > 0 && m.cursor < len(matches) {
					name := matches[m.cursor].Name
					m.revealed[name] = !m.revealed[name]
				}
				return m, nil
			}
			// Fall through to default char handling if not masked
			if m.state != stateSearch || !m.masked {
				if len(msg.String()) == 1 {
					switch m.state {
					case stateSearch:
						m.input += msg.String()
						m.cursor = 0
						m.copied = ""
					case stateAddName:
						m.newName += msg.String()
					case stateAddValue:
						m.newVal += msg.String()
					}
				}
			}
			return m, nil
		case "ctrl+e":
			if m.state == stateSearch {
				matches := m.filteredKeys()
				keys := m.selectedFromMatches(matches)
				if len(keys) == 0 && len(matches) > 0 && m.cursor < len(matches) {
					keys = []db.Key{matches[m.cursor]}
				}
				if len(keys) > 0 {
					m.envExport = true
					m.envExportKeys = keys
					m.done = true
					return m, tea.Quit
				}
				return m, nil
			}
		case "shift+tab", "ctrl+y":
			if m.state == stateSearch {
				matches := m.filteredKeys()
				keys := m.selectedFromMatches(matches)
				// If nothing checked, copy the key under cursor
				if len(keys) == 0 && len(matches) > 0 && m.cursor < len(matches) {
					keys = []db.Key{matches[m.cursor]}
				}
				if len(keys) > 0 {
					var lines []string
					for _, k := range keys {
						lines = append(lines, fmt.Sprintf("export %s=%s", k.Name, k.Value))
					}
					if err := copyToClipboard(strings.Join(lines, "\n")); err == nil {
						m.copied = "done"
						m.copiedFmt = "export"
						m.copiedN = len(keys)
					}
				}
				return m, nil
			}
		case "tab":
			if m.state == stateSearch {
				matches := m.filteredKeys()
				keys := m.selectedFromMatches(matches)
				if len(keys) == 0 && len(matches) > 0 && m.cursor < len(matches) {
					keys = []db.Key{matches[m.cursor]}
				}
				if len(keys) > 0 {
					var lines []string
					for _, k := range keys {
						lines = append(lines, fmt.Sprintf("%s=%s", k.Name, k.Value))
					}
					if err := copyToClipboard(strings.Join(lines, "\n")); err == nil {
						m.copied = "done"
						m.copiedFmt = "env"
						m.copiedN = len(keys)
					}
				}
				return m, nil
			}
		case "backspace":
			switch m.state {
			case stateSearch:
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
					m.cursor = 0
					m.copied = ""
				}
			case stateAddName:
				if len(m.newName) > 0 {
					m.newName = m.newName[:len(m.newName)-1]
				}
			case stateAddValue:
				if len(m.newVal) > 0 {
					m.newVal = m.newVal[:len(m.newVal)-1]
				}
			}
		case "enter":
			switch m.state {
			case stateSearch:
				matches := m.filteredKeys()
				if len(matches) == 0 && m.input != "" {
					m.state = stateAddName
					m.newName = m.input
				}
			case stateAddName:
				if m.newName != "" {
					m.state = stateAddValue
				}
			case stateAddValue:
				if m.newName != "" && m.newVal != "" {
					if err := db.AddKey(m.newName, m.newVal); err != nil {
						m.message = fmt.Sprintf("Error: %v", err)
					} else {
						m.message = fmt.Sprintf("Added %s", m.newName)
					}
					m.done = true
					return m, tea.Quit
				}
			}
		default:
			if len(msg.String()) == 1 {
				switch m.state {
				case stateSearch:
					m.input += msg.String()
					m.cursor = 0
					m.copied = ""
				case stateAddName:
					m.newName += msg.String()
				case stateAddValue:
					m.newVal += msg.String()
				}
			}
		}
	}
	return m, nil
}

func (m SeeModel) filteredKeys() []db.Key {
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

func (m SeeModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder

	switch m.state {
	case stateSearch:
		// Build search bar content
		icon := searchIconStyle.Render(" ")
		var inputText string
		if m.input == "" {
			inputText = placeholderStyle.Render("Search keys...")
		} else {
			inputText = searchInputStyle.Render(m.input)
		}
		barContent := icon + " " + inputText + searchInputStyle.Render("_")

		bar := searchBarFocusedStyle.Render(barContent)
		b.WriteString(bar)
		b.WriteString("\n")

		// Show copied status bar
		if m.copied != "" {
			if m.copiedN == 1 {
				if m.copiedFmt == "env" {
					b.WriteString(copiedStyle.Render("  Copied 1 key as KEY=VAL"))
				} else {
					b.WriteString(copiedStyle.Render("  Copied 1 key as export"))
				}
			} else {
				if m.copiedFmt == "env" {
					b.WriteString(copiedStyle.Render(fmt.Sprintf("  Copied %d keys as KEY=VAL", m.copiedN)))
				} else {
					b.WriteString(copiedStyle.Render(fmt.Sprintf("  Copied %d keys as export", m.copiedN)))
				}
			}
		}
		b.WriteString("\n")

		// Count selected
		matches := m.filteredKeys()
		numSelected := 0
		for _, k := range matches {
			if m.selected[k.Name] {
				numSelected++
			}
		}
		if numSelected > 0 {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  %d selected", numSelected)))
			b.WriteString("\n")
		}
		b.WriteString("\n")

		if len(matches) == 0 {
			if m.input != "" {
				b.WriteString(dimStyle.Render("  No keys found."))
				b.WriteString("\n")
				b.WriteString(hintStyle.Render("  Press enter to add a new key"))
				b.WriteString("\n")
			} else {
				b.WriteString(dimStyle.Render("  No keys stored yet."))
				b.WriteString("\n")
			}
		} else {
			for i, k := range matches {
				pointer := "  "
				if i == m.cursor {
					pointer = cursorStyle.Render("> ")
				}

				check := dimStyle.Render("[ ] ")
				if m.selected[k.Name] {
					check = checkStyle.Render("[x] ")
				}

				age := ageIndicator(k.UpdatedAt)

				name := k.Name
				val := k.Value

				// Handle masked mode
				if m.masked && !m.revealed[k.Name] {
					val = "***"
				}

				if i == m.cursor {
					name = labelStyle.Render(k.Name)
					val = valueStyle.Render(val)
				} else {
					name = dimStyle.Render(k.Name)
					val = dimStyle.Render(val)
				}

				b.WriteString(fmt.Sprintf("%s%s%s%s = %s\n", pointer, check, age, name, val))
			}
		}

	case stateAddName:
		inputContent := searchIconStyle.Render("+") + " " + labelStyle.Render("Name: ") + m.newName + searchInputStyle.Render("_")
		b.WriteString(searchBarFocusedStyle.Render(inputContent))
		b.WriteString("\n")

	case stateAddValue:
		nameContent := searchIconStyle.Render("+") + " " + labelStyle.Render("Name: ") + dimStyle.Render(m.newName)
		b.WriteString(searchBarStyle.Render(nameContent))
		b.WriteString("\n")
		valContent := searchIconStyle.Render("+") + " " + labelStyle.Render("Value: ") + m.newVal + searchInputStyle.Render("_")
		b.WriteString(searchBarFocusedStyle.Render(valContent))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.masked {
		b.WriteString(dimStyle.Render("  space select  r reveal  S-tab/ctrl+y copy export  tab copy KEY=VAL  ctrl+e export .env  esc quit"))
	} else {
		b.WriteString(dimStyle.Render("  space select  S-tab/ctrl+y copy export  tab copy KEY=VAL  ctrl+e export .env  enter add key  esc quit"))
	}
	return b.String()
}
