package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"git-sage/internal/config"
	"git-sage/internal/detect"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	itemStyle     = lipgloss.NewStyle().PaddingLeft(2)
	selectedStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#EE6FF8"))
)

// GeneratorFunc is the function signature for generating the commit message.
type GeneratorFunc func(commitType, emoji string) (string, error)

// steps: 0=select type, 1=select scope, 2=waiting AI, 3=confirm
type model struct {
	step          int
	choices       []string // "emoji type"
	scopeChoices  []string // available scopes + "(none)"
	cursor        int
	selectedType  string
	selectedEmoji string
	selectedScope string

	generatedMsg string

	err      error
	quitting bool

	cfg         *config.Config
	initialType detect.CommitType
	generator   GeneratorFunc
}

func initialModel(cfg *config.Config, initialType detect.CommitType, generator GeneratorFunc) model {
	var choices []string
	order := []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"}
	seen := make(map[string]bool)
	for _, k := range order {
		if emoji, ok := cfg.Emojis[k]; ok {
			choices = append(choices, fmt.Sprintf("%s %s", emoji, k))
			seen[k] = true
		}
	}
	for k, v := range cfg.Emojis {
		if !seen[k] {
			choices = append(choices, fmt.Sprintf("%s %s", v, k))
		}
	}

	cursor := 0
	for i, c := range choices {
		if strings.Contains(c, string(initialType)) {
			cursor = i
			break
		}
	}

	// Build scope choices: "(none)" first, then configured scopes
	scopeChoices := []string{"(none)"}
	scopeChoices = append(scopeChoices, cfg.Scopes...)

	return model{
		step:         0,
		choices:      choices,
		scopeChoices: scopeChoices,
		cursor:       cursor,
		cfg:          cfg,
		initialType:  initialType,
		generator:    generator,
	}
}

func (m model) Init() tea.Cmd { return nil }

type aiResultMsg string
type aiErrorMsg error

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		switch m.step {
		case 0: // select type
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			case "enter":
				parts := strings.SplitN(m.choices[m.cursor], " ", 2)
				m.selectedEmoji = parts[0]
				m.selectedType = parts[1]
				if len(m.scopeChoices) > 1 {
					// Has scopes configured — go to scope step
					m.cursor = 0
					m.step = 1
				} else {
					// No scopes — skip straight to AI generation
					m.step = 2
					return m, generateCmd(m.generator, m.selectedType, m.selectedEmoji)
				}
			}

		case 1: // select scope
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.scopeChoices)-1 {
					m.cursor++
				}
			case "enter":
				chosen := m.scopeChoices[m.cursor]
				if chosen != "(none)" {
					m.selectedScope = chosen
				}
				m.step = 2
				return m, generateCmd(m.generator, m.selectedType, m.selectedEmoji)
			}

		case 3: // confirm
			switch msg.String() {
			case "y", "enter":
				return m, tea.Quit
			case "n", "q":
				m.quitting = true
				return m, tea.Quit
			}
		}

	case aiResultMsg:
		m.generatedMsg = string(msg)
		m.step = 3

	case aiErrorMsg:
		m.err = error(msg)
		return m, tea.Quit
	}
	return m, nil
}

func generateCmd(generator GeneratorFunc, t, e string) tea.Cmd {
	return func() tea.Msg {
		res, err := generator(t, e)
		if err != nil {
			return aiErrorMsg(err)
		}
		return aiResultMsg(res)
	}
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}
	if m.quitting {
		return "Aborted.\n"
	}

	var s string
	switch m.step {
	case 0:
		s = titleStyle.Render("Select Commit Type:") + "\n\n"
		for i, choice := range m.choices {
			if m.cursor == i {
				s += selectedStyle.Render(fmt.Sprintf("> %s", choice)) + "\n"
			} else {
				s += itemStyle.Render(fmt.Sprintf("  %s", choice)) + "\n"
			}
		}
		s += "\n(arrows to navigate, enter to select)\n"

	case 1:
		s = titleStyle.Render("Select Scope (optional):") + "\n\n"
		for i, scope := range m.scopeChoices {
			if m.cursor == i {
				s += selectedStyle.Render(fmt.Sprintf("> %s", scope)) + "\n"
			} else {
				s += itemStyle.Render(fmt.Sprintf("  %s", scope)) + "\n"
			}
		}
		s += "\n(arrows to navigate, enter to select)\n"

	case 2:
		s = fmt.Sprintf("Generating commit message for %s %s...\nPlease wait...", m.selectedEmoji, m.selectedType)

	case 3:
		typeStr := m.selectedType
		if m.selectedScope != "" {
			typeStr = fmt.Sprintf("%s(%s)", m.selectedType, m.selectedScope)
		}
		fullMsg := fmt.Sprintf("%s %s: %s", m.selectedEmoji, typeStr, m.generatedMsg)
		s = titleStyle.Render("Confirm Commit Message:") + "\n\n"
		s += fmt.Sprintf("  %s\n\n", fullMsg)
		s += "Commit this? (y/n): "
	}

	return s
}

func (m model) Result() (string, string, string, string, error, bool) {
	return m.selectedType, m.selectedEmoji, m.selectedScope, m.generatedMsg, m.err, !m.quitting
}

// Run starts the TUI and returns (type, emoji, scope, message, error).
func Run(cfg *config.Config, initialType detect.CommitType, generator GeneratorFunc) (string, string, string, string, error) {
	p := tea.NewProgram(initialModel(cfg, initialType, generator))
	finalModel, err := p.Run()
	if err != nil {
		return "", "", "", "", err
	}

	if m, ok := finalModel.(model); ok {
		t, e, scope, msg, modErr, ok := m.Result()
		if modErr != nil {
			return "", "", "", "", modErr
		}
		if !ok {
			return "", "", "", "", fmt.Errorf("aborted by user")
		}
		return t, e, scope, msg, nil
	}

	return "", "", "", "", fmt.Errorf("unknown model state")
}
