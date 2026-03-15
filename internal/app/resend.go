package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/MonsieurTib/service-bus-tui/internal/azure"
	"github.com/MonsieurTib/service-bus-tui/internal/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type resendState int

const (
	stateChooseMode resendState = iota
	stateSending
	stateDone
	stateError
)

type ResendDismissedMsg struct{}

type ResendOverlayModel struct {
	state       resendState
	messages    []azure.MessageInfo
	destination string
	client      *azure.ServiceBusClient
	preserveIDs bool
	sent        int
	total       int
	err         error
	spinner     spinner.Model
	progressCh  <-chan azure.SendProgress
}

func NewResendOverlayModel(messages []azure.MessageInfo, destination string, client *azure.ServiceBusClient) *ResendOverlayModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	return &ResendOverlayModel{
		state:       stateChooseMode,
		messages:    messages,
		destination: destination,
		client:      client,
		preserveIDs: true,
		total:       len(messages),
		spinner:     s,
	}
}

func (m *ResendOverlayModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *ResendOverlayModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case ResendProgressMsg:
		m.sent = msg.Sent
		if msg.Err != nil {
			m.err = msg.Err
			m.state = stateError
			return nil
		}
		if msg.Done {
			m.state = stateDone
			return nil
		}
		return waitForProgress(m.progressCh)

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd
	}
}

func (m *ResendOverlayModel) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch m.state {
	case stateChooseMode:
		switch msg.String() {
		case "up", "k", "down", "j":
			m.preserveIDs = !m.preserveIDs
		case "enter":
			return m.startSending()
		case "esc":
			return func() tea.Msg { return ResendDismissedMsg{} }
		}

	case stateSending:
		// No cancellation — ignore keys during send
		return nil

	case stateDone, stateError:
		switch msg.String() {
		case "esc", "enter":
			return func() tea.Msg { return ResendDismissedMsg{} }
		}
	}

	return nil
}

func (m *ResendOverlayModel) startSending() tea.Cmd {
	m.state = stateSending
	m.progressCh = m.client.SendMessages(
		context.Background(),
		m.destination,
		m.messages,
		m.preserveIDs,
	)
	return tea.Batch(m.spinner.Tick, waitForProgress(m.progressCh))
}

func (m *ResendOverlayModel) View(width, height int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Primary).
		Bold(true).
		MarginBottom(1)

	switch m.state {
	case stateChooseMode:
		content.WriteString(titleStyle.Render(fmt.Sprintf("Resend %d message(s) to %s", m.total, m.destination)))
		content.WriteString("\n\n")
		content.WriteString(styles.Subtle.Render("Choose Message ID strategy:"))
		content.WriteString("\n\n")

		options := []struct {
			label    string
			selected bool
		}{
			{"Keep original Message IDs", m.preserveIDs},
			{"Generate new Message IDs", !m.preserveIDs},
		}
		for _, opt := range options {
			if opt.selected {
				content.WriteString(lipgloss.NewStyle().
					Foreground(styles.Primary).
					Bold(true).
					Render("▸ " + opt.label))
			} else {
				content.WriteString("  " + opt.label)
			}
			content.WriteString("\n")
		}
		content.WriteString("\n")
		content.WriteString(styles.Subtle.Render("enter: confirm · esc: cancel"))

	case stateSending:
		content.WriteString(titleStyle.Render("Resending messages"))
		content.WriteString("\n\n")
		content.WriteString(m.spinner.View() + " ")
		content.WriteString(fmt.Sprintf("Sending %d/%d to %s...", m.sent, m.total, m.destination))
		content.WriteString("\n\n")
		content.WriteString(m.renderProgressBar(36))

	case stateDone:
		content.WriteString(titleStyle.Render("Resend complete"))
		content.WriteString("\n\n")
		content.WriteString(lipgloss.NewStyle().Foreground(styles.Success).Bold(true).
			Render(fmt.Sprintf("✓ Successfully sent %d/%d messages", m.sent, m.total)))
		content.WriteString("\n\n")
		content.WriteString(styles.Subtle.Render("press esc or enter to close"))

	case stateError:
		content.WriteString(titleStyle.Render("Resend failed"))
		content.WriteString("\n\n")
		content.WriteString(lipgloss.NewStyle().Foreground(styles.ErrorColor).Bold(true).
			Render(fmt.Sprintf("✗ Error after %d/%d", m.sent, m.total)))
		content.WriteString("\n")
		content.WriteString(styles.Error.Render(m.err.Error()))
		content.WriteString("\n\n")
		content.WriteString(styles.Subtle.Render("press esc or enter to close"))
	}

	panelWidth := min(52, width-4)
	panelHeight := min(14, height-4)

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(panelWidth).
		Height(panelHeight).
		Background(lipgloss.Color("235"))

	panel := panelStyle.Render(content.String())

	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		panel,
	)
}

func (m *ResendOverlayModel) renderProgressBar(barWidth int) string {
	if m.total == 0 {
		return ""
	}
	pct := float64(m.sent) / float64(m.total)
	filled := int(pct * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	pctStr := fmt.Sprintf(" %d%%", int(pct*100))

	return lipgloss.NewStyle().Foreground(styles.Primary).Render(bar) +
		styles.Subtle.Render(pctStr)
}

func waitForProgress(ch <-chan azure.SendProgress) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return ResendProgressMsg{Done: true}
		}
		return ResendProgressMsg{
			Sent:  p.Sent,
			Total: p.Total,
			Err:   p.Err,
			Done:  p.Done,
		}
	}
}
