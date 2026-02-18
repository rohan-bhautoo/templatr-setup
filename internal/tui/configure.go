package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/templatr/templatr-setup/internal/manifest"
)

// configField wraps a manifest field definition with its input model.
type configField struct {
	key         string // env key or config path
	label       string
	description string
	fieldType   string // text, url, email, secret, number, boolean
	required    bool
	section     string // "env" or config file label
	input       textinput.Model
}

// configureModel manages the configure form.
type configureModel struct {
	fields  []configField
	focused int
	done    bool
	skipped bool
}

func newConfigureModel(m *manifest.Manifest) configureModel {
	var fields []configField

	// .env fields
	for _, env := range m.Env {
		ti := textinput.New()
		ti.Placeholder = env.Default
		ti.CharLimit = 256
		ti.Width = 50

		if env.Type == "secret" {
			ti.EchoMode = textinput.EchoPassword
		}

		// Pre-fill with default if set
		if env.Default != "" {
			ti.SetValue(env.Default)
		}

		fields = append(fields, configField{
			key:         env.Key,
			label:       env.Label,
			description: env.Description,
			fieldType:   env.Type,
			required:    env.Required,
			section:     "Environment Variables (.env)",
			input:       ti,
		})
	}

	// Config file fields
	for _, cfg := range m.Config {
		for _, f := range cfg.Fields {
			ti := textinput.New()
			ti.Placeholder = f.Default
			ti.CharLimit = 256
			ti.Width = 50

			if f.Default != "" {
				ti.SetValue(f.Default)
			}

			fields = append(fields, configField{
				key:         f.Path,
				label:       f.Label,
				description: f.Description,
				fieldType:   f.Type,
				section:     cfg.Label,
				input:       ti,
			})
		}
	}

	// Focus the first field
	if len(fields) > 0 {
		fields[0].input.Focus()
	}

	return configureModel{
		fields: fields,
	}
}

func (m configureModel) Update(msg tea.Msg) (configureModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.fields[m.focused].input.Blur()
			m.focused = (m.focused + 1) % len(m.fields)
			m.fields[m.focused].input.Focus()
			return m, m.fields[m.focused].input.Focus()

		case "shift+tab", "up":
			m.fields[m.focused].input.Blur()
			m.focused = (m.focused - 1 + len(m.fields)) % len(m.fields)
			m.fields[m.focused].input.Focus()
			return m, m.fields[m.focused].input.Focus()

		case "enter":
			// If on last field, submit
			if m.focused == len(m.fields)-1 {
				m.done = true
				return m, nil
			}
			// Otherwise move to next field
			m.fields[m.focused].input.Blur()
			m.focused++
			m.fields[m.focused].input.Focus()
			return m, m.fields[m.focused].input.Focus()
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	m.fields[m.focused].input, cmd = m.fields[m.focused].input.Update(msg)
	return m, cmd
}

func (m configureModel) View() string {
	if len(m.fields) == 0 {
		return mutedStyle.Render("No configuration fields defined.")
	}

	var b strings.Builder

	b.WriteString(boldStyle.Render("Configure Your Template"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Tab/Shift+Tab to navigate, Enter to submit"))
	b.WriteString("\n\n")

	currentSection := ""
	for i, f := range m.fields {
		// Section header
		if f.section != currentSection {
			if currentSection != "" {
				b.WriteString("\n")
			}
			currentSection = f.section
			b.WriteString(infoStyle.Render(fmt.Sprintf("── %s ──", currentSection)))
			b.WriteString("\n\n")
		}

		// Field label
		label := f.label
		if f.required {
			label += " " + errorStyle.Render("*")
		}

		if i == m.focused {
			b.WriteString(highlightStyle.Render(fmt.Sprintf("  %s %s", iconArrow, label)))
		} else {
			b.WriteString(fmt.Sprintf("    %s", boldStyle.Render(label)))
		}
		b.WriteString("\n")

		// Description
		if f.description != "" {
			b.WriteString(fmt.Sprintf("    %s\n", mutedStyle.Render(f.description)))
		}

		// Input
		b.WriteString(fmt.Sprintf("    %s\n", f.input.View()))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.focused == len(m.fields)-1 {
		b.WriteString(highlightStyle.Render("  Press Enter to save configuration"))
	} else {
		b.WriteString(mutedStyle.Render("  Press Tab to move to next field"))
	}

	return b.String()
}

// Values returns the filled-in values as a map.
// Keys are env var keys or config field paths.
func (m configureModel) Values() map[string]string {
	vals := make(map[string]string, len(m.fields))
	for _, f := range m.fields {
		v := f.input.Value()
		if v == "" && f.input.Placeholder != "" {
			v = f.input.Placeholder // use default
		}
		vals[f.key] = v
	}
	return vals
}

// EnvValues returns only the .env field values.
func (m configureModel) EnvValues(manifest *manifest.Manifest) map[string]string {
	vals := make(map[string]string)
	for _, env := range manifest.Env {
		for _, f := range m.fields {
			if f.key == env.Key {
				v := f.input.Value()
				if v == "" {
					v = f.input.Placeholder
				}
				vals[env.Key] = v
				break
			}
		}
	}
	return vals
}

// ConfigValues returns only the config file field values.
func (m configureModel) ConfigValues(manifest *manifest.Manifest) map[string]string {
	vals := make(map[string]string)
	for _, cfg := range manifest.Config {
		for _, field := range cfg.Fields {
			for _, f := range m.fields {
				if f.key == field.Path {
					v := f.input.Value()
					if v == "" {
						v = f.input.Placeholder
					}
					vals[field.Path] = v
					break
				}
			}
		}
	}
	return vals
}
